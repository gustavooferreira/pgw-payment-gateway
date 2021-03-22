# Docker builder for Golang
FROM golang:1.16-buster AS builder

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-w -s" -installsuffix 'static' \
    -o /api-server cmd/api-server/main.go

# Final stage: running container
FROM scratch AS final

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the compiled executable from the first stage.
COPY --from=builder /api-server /api-server

# Declare the port on which the webserver will be exposed.
# As we're going to run the executable as an unprivileged user, we can't bind
# to ports below 1024.
EXPOSE 8080/TCP

USER nobody:nobody

# Set env vars
ENV PGW_PAYMENT_GATEWAY_APP_WEBSERVER_HOST 0.0.0.0
ENV PGW_PAYMENT_GATEWAY_APP_WEBSERVER_PORT 8080

CMD ["/api-server"]

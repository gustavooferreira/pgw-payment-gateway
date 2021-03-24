# Payment Gateway Exercise - Payment Gateway Service

Please check out documentation for the whole system [here](https://github.com/gustavooferreira/pgw-docs).

This repository structure follows [this convention](https://github.com/golang-standards/project-layout).

---

## Tip

> If you run `make` without any targets, it will display all options available on the makefile followed by a short description.

# Build

To build a binary, run:

```bash
make build
```

The `api-server` binary will be placed inside the `bin/` folder.

---

# Tests

To run tests:

```bash
make test
```

To get coverage:

```bash
make coverage
```

# Docker

To build the docker image, run:

```bash
make build-docker
```

The docker image is named `pgw/payment-gateway-api-server`.

To start a docker container, run:

```bash
docker run --rm --name pgw-payment-gateway-service -p 127.0.0.1:9000:8080/tcp pgw/payment-gateway-api-server
```

Once the container is running, you can make a request like this:

```bash
curl -i -X POST -u bill:pass1 http://localhost:9000/api/v1/authorise -d '{"credit_card": {"name":"customer1", "number": 4000000000000001, "expiry_month":10, "expiry_year":2030, "cvv":123}, "currency": "EUR", "amount": 10.50}'
```

# Design

## MySQL tables

Table `credit_cards`:

| Field        | Type        |
| ------------ | ----------- |
| number       | bigint      |
| name         | varchar(40) |
| expiry_month | bigint      |
| expiry_year  | bigint      |
| cvv          | bigint      |

Table `authorisations`:

| Field              | Type        |
| ------------------ | ----------- |
| uid                | varchar(50) |
| amount             | double      |
| currency           | longtext    |
| state              | longtext    |
| merchant_name      | longtext    |
| credit_card_number | bigint      |

Table `transactions`:

| Field             | Type            |
| ----------------- | --------------- |
| id                | bigint unsigned |
| amount            | double          |
| type              | varchar(20)     |
| authorisation_uid | varchar(50)     |

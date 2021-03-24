package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
)

// AuthUserKey is the name of the user credential in basic auth.
const AuthUserKey = "user"

func GinBasicAuth(logger log.Logger, httpClient *http.Client, authServiceHost string, authServicePort int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		auth := c.Request.Header.Get("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Basic ")
		if token == auth {
			c.JSON(http.StatusForbidden, gin.H{"message": "could not find BasicAuth Authorization token"})
			c.Abort()
			return
		}

		decodedToken, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "could not decode BasicAuth Authorization token"})
			c.Abort()
			return
		}

		decodedTokenStr := string(decodedToken)
		credentials := strings.Split(decodedTokenStr, ":")
		if len(credentials) != 2 {
			c.JSON(http.StatusForbidden, gin.H{"message": "could not decode BasicAuth Authorization token"})
			c.Abort()
			return
		}

		// Send http request to validate credentials
		valid, err := CheckCredentials(httpClient, authServiceHost, authServicePort, credentials[0], credentials[1])
		if err != nil {
			logger.Error(fmt.Sprintf("basicauth middleware error: %s", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			c.Abort()
			return
		}

		if !valid {
			c.JSON(http.StatusForbidden, gin.H{"message": "provided credentials are not valid"})
			c.Abort()
			return
		}

		// The user credentials are valid, set user's id to key AuthUserKey
		c.Set(AuthUserKey, credentials[0])
	}
}

func CheckCredentials(httpClient *http.Client, host string, port int, username string, password string) (bool, error) {
	requestBodyData := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: username, Password: password}

	requestBody, err := json.Marshal(requestBodyData)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("http://%s:%d/api/v1/auth", host, port)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	responseBodyData := struct {
		Valid bool `json:"valid"`
	}{}

	err = json.Unmarshal(responseBody, &responseBodyData)
	if err != nil {
		return false, err
	}

	if !responseBodyData.Valid {
		return false, nil
	}

	return true, nil
}

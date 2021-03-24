package pprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// NOTE: This client (SDK) needs improvement, like better error reporting in logs and what not

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(host string, port int, httpClient *http.Client) *Client {
	c := &Client{httpClient: httpClient, baseURL: fmt.Sprintf("http://%s:%d/api/v1", host, port)}
	return c
}

func (c *Client) AuthorisePayment(authReq AuthorisationRequest) (authID string, success bool) {
	requestBody, err := json.Marshal(authReq)
	if err != nil {
		return "", false
	}

	url := c.baseURL + "/authorise"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", false
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	var responseBodyData AuthorisationResponse

	err = json.Unmarshal(responseBody, &responseBodyData)
	if err != nil {
		return "", false
	}

	if responseBodyData.Code != 1 {
		return "", false
	}

	return responseBodyData.AuthorisationID, true
}

func (c *Client) CaptureTransaction(capReq CaptureRequest) (success bool) {
	requestBody, err := json.Marshal(capReq)
	if err != nil {
		return false
	}

	url := c.baseURL + "/capture"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var responseBodyData CaptureResponse

	err = json.Unmarshal(responseBody, &responseBodyData)
	if err != nil {
		return false
	}

	if responseBodyData.Code != 1 {
		return false
	}

	return true
}

func (c *Client) RefundTransaction(refReq RefundRequest) (success bool) {
	requestBody, err := json.Marshal(refReq)
	if err != nil {
		return false
	}

	url := c.baseURL + "/refund"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var responseBodyData RefundResponse

	err = json.Unmarshal(responseBody, &responseBodyData)
	if err != nil {
		return false
	}

	if responseBodyData.Code != 1 {
		return false
	}

	return true
}

func (c *Client) VoidPayment(voidReq VoidRequest) (success bool) {
	requestBody, err := json.Marshal(voidReq)
	if err != nil {
		return false
	}

	url := c.baseURL + "/void"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var responseBodyData VoidResponse

	err = json.Unmarshal(responseBody, &responseBodyData)
	if err != nil {
		return false
	}

	if responseBodyData.Code != 1 {
		return false
	}

	return true
}

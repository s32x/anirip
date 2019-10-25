package httpclient /* import "s32x.com/httpclient" */

import (
	"encoding/json"
	"net/http"
)

// GetString calls GetString using the DefaultClient
func GetString(url string) (string, error) {
	return DefaultClient.GetString(url, nil)
}

// GetString performs a GET request and returns the response as a string
func (c *Client) GetString(path string, headers map[string]string) (string, error) {
	// Retrieve the bytes and decode the response
	body, err := c.GetBytes(path, headers)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetJSON calls GetJSON using the DefaultClient
func GetJSON(url string, out interface{}) error {
	return DefaultClient.GetJSON(url, nil, out)
}

// GetJSON performs a basic http GET request and decodes the JSON response into
// the out interface
func (c *Client) GetJSON(path string, headers map[string]string, out interface{}) error {
	// Retrieve the bytes and decode the response
	body, err := c.GetBytes(path, headers)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

// GetBytes calls GetBytes using the DefaultClient
func GetBytes(url string) ([]byte, error) {
	return DefaultClient.GetBytes(url, nil)
}

// GetBytes performs a GET request using the passed path and headers. It
// expects a 200 code status in the response and returns the bytes on the
// response
func (c *Client) GetBytes(path string, headers map[string]string) ([]byte, error) {
	// Execute the request and return the response
	res, err := c.DoWithStatus(NewRequest(http.MethodGet, path, headers, nil),
		http.StatusOK)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

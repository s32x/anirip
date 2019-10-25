package httpclient /* import "s32x.com/httpclient" */

// httpclient is a convenience package for executing HTTP requests. It's safe
// in that it always closes response bodies and returns byte slices, strings or
// decodes responses into interfaces

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"
)

// Client is an http.Client wrapper
type Client struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// DefaultClient is a basic Client for use without needing to define a Client
var DefaultClient = New()

// New creates a new Client reference given a client timeout
func New() *Client {
	return &Client{client: &http.Client{}}
}

// SetTimeout sets the timeout on the httpclients client
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.Timeout = timeout
	return c
}

// SetTransport sets the Transport on the httpclients client
func (c *Client) SetTransport(transport *http.Transport) *Client {
	c.client.Transport = transport
	return c
}

// SetBaseURL sets the baseURL on the Client which will be used on all
// subsequent requests
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// SetHeaders sets the headers on the Client which will be used on all
// subsequent requests
func (c *Client) SetHeaders(headers map[string]string) *Client {
	c.headers = headers
	return c
}

// DoWithStatus performs the request and asserts the status code on the response
func (c *Client) DoWithStatus(req *Request, expectedStatus int) (*Response, error) {
	// Perform the request
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Assert the status and return the response
	if res.StatusCode != expectedStatus {
		return res, errors.New("Unexpected status-code on response")
	}
	return res, nil
}

// Do performs the passed request and returns a fully populated response
func (c *Client) Do(req *Request) (*Response, error) {
	// Encode the body if one was passed
	var b io.ReadWriter
	if req.Body != nil {
		b = bytes.NewBuffer(req.Body)
	}

	// Generate a new request using the new URL
	r, err := http.NewRequest(req.Method, c.baseURL+req.Path, b)
	if err != nil {
		return nil, err
	}

	// Add all desired headers prioritizing request headers over global client
	// headers
	if c.headers != nil {
		for k, v := range c.headers {
			r.Header.Set(k, v)
		}
	}
	if req.Headers != nil {
		for k, v := range req.Headers {
			r.Header.Set(k, v)
		}
	}

	// Execute the fully constructed request
	res, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode the response into a Response and return
	return NewResponse(res)
}

package httpclient /* import "s32x.com/httpclient" */

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// Response is a basic HTTP response struct containing just the important data
// returned from an HTTP request
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

// NewResponse creates a more basic HTTP Response from the passed http.Response
func NewResponse(res *http.Response) (*Response, error) {
	// Create a map of all headers
	headers := make(map[string][]string)
	for k, v := range res.Header {
		if len(v) > 0 {
			headers[k] = strings.Split(v[0], ", ")
		}
	}

	// Decode the body into a byte slice
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Return the fully populated Response
	return &Response{
		StatusCode: res.StatusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

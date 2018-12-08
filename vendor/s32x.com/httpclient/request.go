package httpclient /* import "s32x.com/httpclient" */

// Request is a basic HTTP request struct containing just what is needed to
// perform a standard HTTP request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// NewRequest builds a new Request with the passed data
func NewRequest(method, path string, headers map[string]string, body []byte) *Request {
	return &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}

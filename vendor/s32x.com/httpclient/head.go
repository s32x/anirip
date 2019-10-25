package httpclient /* import "s32x.com/httpclient" */

import "net/http"

// Head calls Head using the DefaultClient
func Head(url string) error {
	return DefaultClient.Head(url, nil)
}

// Head performs a HEAD request using the passed path and headers. It expects a
// 200 status code in the response
func (c *Client) Head(path string, headers map[string]string) error {
	_, err := c.DoWithStatus(NewRequest(http.MethodHead, path, headers, nil),
		http.StatusOK)
	return err
}

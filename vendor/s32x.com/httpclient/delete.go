package httpclient /* import "s32x.com/httpclient" */

import "net/http"

// Delete calls Delete using the DefaultClient
func Delete(url string) error {
	return DefaultClient.Delete(url, nil)
}

// Delete performs a DELETE request using the passed path and headers. It
// expects a 200 code status in the response
func (c *Client) Delete(path string, headers map[string]string) error {
	_, err := c.DoWithStatus(NewRequest(http.MethodDelete, path, headers, nil),
		http.StatusOK)
	return err
}

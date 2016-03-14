package anirip

import (
	"io"
	"net/http"
)

// A shorthand function for writing http requests. NOTE: Uses a default user-agent for every request
func GetHTTPResponse(method, urlStr string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	// Builds out request based on first 3 params
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, Error{Message: "There was an error creating the " + method + " request on " + urlStr, Err: err}
	}

	// Sets the headers passed as the request headers
	request.Header = header
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")

	// Attaches all cookies passed
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	// Executes the new request and returns the response, retries recursively if theres a failure
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, Error{Message: "There was an error performing the " + method + " request on " + urlStr, Err: err}
	}
	return response, nil
}

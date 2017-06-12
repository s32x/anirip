package anirip

import (
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// HTTPClient performs
type HTTPClient struct {
	Client  *http.Client
	Header  http.Header
	Cookies []*http.Cookie
}

// NewHTTPClient generates a new HTTPClient Requester that contains a random
// user-agent to partially mask anirip requests
func NewHTTPClient() (*HTTPClient, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/cvandeplas/pystemon/master/user-agents.txt")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Splits the user-agents into a slice and returns an HTTPClient with a random
	// user-agent on the header
	ua := strings.Split(string(b), "\n")
	rand.Seed(time.Now().UnixNano())
	header := http.Header{}
	header.Add("user-agent", ua[rand.Intn(len(ua))])
	return &HTTPClient{
		Client:  &http.Client{},
		Header:  header,
		Cookies: []*http.Cookie{},
	}, nil
}

// Post performs a POST request using the passed url and body using the
// in-memory HTTPClient
func (c *HTTPClient) Post(url string, body io.Reader) (*http.Response, error) {
	return c.request(http.MethodPost, url, body)
}

// Get performs a GET request using the passed url using the in-memory
// HTTPClient
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	return c.request(http.MethodGet, url, nil)
}

func (c *HTTPClient) request(method, url string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	r.Header = c.Header
	for _, cookie := range c.Cookies {
		r.AddCookie(cookie)
	}
	return c.Client.Do(r)
}

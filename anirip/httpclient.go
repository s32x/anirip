package anirip

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HTTPClient performs
type HTTPClient struct {
	Client    *http.Client
	UserAgent string
}

// NewHTTPClient generates a new HTTPClient Requester that contains a random
// user-agent to emulate browser requests
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

	// Create the client and attach a cookie jar
	client := &http.Client{}
	client.Jar, _ = cookiejar.New(nil)

	// Splits the user-agents into a slice and returns an HTTPClient with a random
	// user-agent on the header
	ua := strings.Split(string(b), "\n")
	rand.Seed(time.Now().UnixNano())
	return &HTTPClient{
		Client:    client,
		UserAgent: ua[rand.Intn(len(ua))],
	}, nil
}

// Post performs a POST request using the passed url and body using the
// in-memory HTTPClient
func (c *HTTPClient) Post(url string, header http.Header, body io.Reader) (*http.Response, error) {
	// Assemble our request and attach all headers and cookies
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	return c.request(req)
}

// Get performs a GET request using the passed url using the in-memory
// HTTPClient
func (c *HTTPClient) Get(url string, header http.Header) (*http.Response, error) {
	// Assemble our request and attach all headers and cookies
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	return c.request(req)
}

// Executes standard requests and returns the response
func (c *HTTPClient) request(req *http.Request) (*http.Response, error) {
	// Adds headers used on ALL requests
	req.Header.Add("User-Agent", c.UserAgent)

	// Executes the request
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// If the server is in IUAM mode, solve the challenge and retry
	if res.StatusCode == 503 && res.Header.Get("Server") == "cloudflare-nginx" {
		defer res.Body.Close()
		var rb []byte
		rb, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return c.bypassCF(req, rb)
	}
	return res, err
}

// bypass attempts to re-execute a standard request after first bypassing
// Cloudflares I'm Under Attack Mode
func (c *HTTPClient) bypassCF(req *http.Request, body []byte) (*http.Response, error) {
	_, ov, _ := otto.Run(extractChallenge(body))
	answerI, _ := strconv.Atoi(strings.TrimSpace(ov.String()))

	// Extracts the challenge variables needed from the HTML
	vc, _ := regexp.Compile(`name="jschl_vc" value="(\w+)"`)
	pass, _ := regexp.Compile(`name="pass" value="(.+?)"`)
	vcMatch := vc.FindSubmatch(body)
	passMatch := pass.FindSubmatch(body)

	if !(len(vcMatch) == 2 && len(passMatch) == 2) {
		return nil, errors.New("Failed to extract Cloudflare IUAM challenge")
	}

	// Assemble the CFClearence request
	u, _ := url.Parse(fmt.Sprintf("%s://%s/cdn-cgi/l/chk_jschl", req.URL.Scheme, req.URL.Host))
	query := u.Query()
	query.Set("jschl_vc", string(vcMatch[1]))
	query.Set("jschl_answer", fmt.Sprintf("%d", answerI+len(u.Host)))
	query.Set("pass", string(passMatch[1]))
	u.RawQuery = query.Encode()
	req.Header.Add("Referer", req.URL.String())

	// Execute, populate cookies after 5 seconds and re-execute prior request
	time.Sleep(5 * time.Second)
	_, err := c.Get(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

// extractChallenge extracts the IUAM challenge javascript from the page body
func extractChallenge(body []byte) string {
	r1, _ := regexp.Compile(`setTimeout\(function\(\){\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?\r?\n[\s\S]+?a\.value =.+?)\r?\n`)
	r2, _ := regexp.Compile(`a\.value = (parseInt\(.+?\)) \+ .+?;`)
	r3, _ := regexp.Compile(`\s{3,}[a-z](?: = |\.).+`)
	r4, _ := regexp.Compile(`[\n\\']`)
	r1Match := r1.FindSubmatch(body)

	if len(r1Match) != 2 {
		return ""
	}

	js := string(r1Match[1])
	js = r2.ReplaceAllString(js, "$1")
	js = r3.ReplaceAllString(js, "")
	js = r4.ReplaceAllString(js, "")

	lastSemicolon := strings.LastIndex(js, ";")
	if lastSemicolon >= 0 {
		js = js[:lastSemicolon]
	}
	return js
}

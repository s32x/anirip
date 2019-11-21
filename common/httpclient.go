package common /* import "s32x.com/anirip/common" */

import (
	"fmt"
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

	"github.com/robertkrimen/otto"
	"s32x.com/anirip/common/log"
	"s32x.com/httpclient"
)

const (
	uaList    = "https://raw.githubusercontent.com/cvandeplas/pystemon/master/user-agents.txt"
	defaultUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36"
)

// HTTPClient performs
type HTTPClient struct {
	Client    *http.Client
	UserAgent string
}

// NewHTTPClient generates a new HTTPClient Requester that
// contains a random user-agent to emulate browser requests
func NewHTTPClient() *HTTPClient {
	// Create the client and attach a cookie jar
	client := &http.Client{}
	client.Jar, _ = cookiejar.New(nil)

	return &HTTPClient{
		Client:    client,
		UserAgent: randomUA(),
	}
}

// randomUA retrieves a list of user-agents and returns a
// one randomly selected from the list
func randomUA() string {
	// Retrieve the bytes of the user-agent list
	useragents, err := httpclient.GetString(uaList)
	if err != nil {
		return defaultUA
	}

	// Split all user-agents into a slice and return a
	// single random one
	ua := strings.Split(useragents, "\n")
	if len(ua) == 0 {
		return defaultUA
	}
	return ua[rand.Intn(len(ua))]
}

// Post performs a POST request using the passed url and
// body using the in-memory HTTPClient
func (c *HTTPClient) Post(url string, header http.Header, body io.Reader) (*http.Response, error) {
	// Assemble our request and attach all headers and
	// cookies
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	return c.request(req)
}

// Get performs a GET request using the passed url using
// the in-memory HTTPClient
func (c *HTTPClient) Get(url string, header http.Header) (*http.Response, error) {
	// Assemble our request and attach all headers and cookies
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating GET request: %w", err)
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
		return nil, fmt.Errorf("processing request: %w", err)
	}

	// If the server is in IUAM mode, solve the challenge and retry
	if res.StatusCode == 503 && res.Header.Get("Server") == "cloudflare" {
		log.Warn("Performing IUAM bypass (this sometimes takes several tries)...")
		defer res.Body.Close()
		var rb []byte
		rb, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("reading IUAM response: %w", err)
		}
		return c.bypassCF(req, rb)
	}
	return res, nil
}

// bypass attempts to re-execute a standard request after first bypassing
// Cloudflares I'm Under Attack Mode
func (c *HTTPClient) bypassCF(req *http.Request, body []byte) (*http.Response, error) {
	// Strip out the full challenge script
	r1, _ := regexp.Compile(`setTimeout\(function\(\){\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?\r?\n[\s\S]+?a\.value =.+?)\r?\n`)
	r1Match := r1.FindSubmatch(body)
	if len(r1Match) != 2 {
		return nil, fmt.Errorf("failed to match on IUAM challenge")
	}
	js := string(r1Match[1])

	// Remove any DOM manipulation
	r3, _ := regexp.Compile(`\s{3,}[a-z](?: = |\.).+`)
	r4, _ := regexp.Compile(`[\n\\']`)
	js = r3.ReplaceAllString(js, "")
	js = r4.ReplaceAllString(js, "")

	// Trim off anything after the last semicolon
	lastSemicolon := strings.LastIndex(js, ";")
	if lastSemicolon >= 0 {
		js = js[:lastSemicolon]
	}

	// Replace t.length with the length of the url
	js = strings.Replace(js, " + t.length", " + "+strconv.Itoa(len(req.URL.Host)), -1)

	// Replace a.value with the otto selector and return
	js = strings.Replace(js, "a.value", "$1", -1)

	// Run the script with otto, storing the result in ov
	_, ov, _ := otto.Run(js)
	answerI, _ := strconv.ParseFloat(ov.String(), 64)

	// Extracts the challenge variables needed from the HTML
	vc, _ := regexp.Compile(`name="jschl_vc" value="(.+?)"`)
	pass, _ := regexp.Compile(`name="pass" value="(.+?)"`)
	vcMatch := vc.FindSubmatch(body)
	passMatch := pass.FindSubmatch(body)

	if !(len(vcMatch) == 2 && len(passMatch) == 2) {
		return nil, fmt.Errorf("failed to extract IUAM challenge")
	}

	// Assemble the CFClearence request
	u, _ := url.Parse(fmt.Sprintf("%s://%s/cdn-cgi/l/chk_jschl", req.URL.Scheme, req.URL.Host))
	query := u.Query()
	query.Set("jschl_vc", string(vcMatch[1]))
	query.Set("pass", string(passMatch[1]))
	query.Set("jschl_answer", fmt.Sprintf("%g", answerI))
	u.RawQuery = query.Encode()

	// Execute, populate cookies after 5 seconds and re-execute prior request
	time.Sleep(4000 * time.Millisecond)
	if _, err := c.Get(u.String(), nil); err != nil {
		return nil, fmt.Errorf("getting IUAM request: %w", err)
	}
	return c.request(req)
}

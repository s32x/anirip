package crunchyroll

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/anirip/anirip"
)

// Login logs the user in to Crunchyroll and stores the session on the client
func Login(c *anirip.HTTPClient, user, pass string) error {
	// Perform preflight request to retrieve the login page
	_, err := c.Get("https://www.crunchyroll.com/login", nil)
	if err != nil {
		return err
	}

	// Sets the credentials and attempts to generate new cookies
	if err := createSession(c, user, pass); err != nil {
		return err
	}

	// Validates the session created and returns
	if err := validateSession(c); err != nil {
		return err
	}
	return nil
}

// createSession sends a login request to Crunchyroll and stores the cookies
// recieved in the clients cookiejar
func createSession(c *anirip.HTTPClient, user, pass string) error {
	// Assemble the FormData needed for authentication
	body := bytes.NewBufferString(url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {user},
		"password": {pass},
	}.Encode())

	// Execute the post request on the formhandler and store cookies in the jar
	head := http.Header{}
	head.Add("Referer", "https://www.crunchyroll.com/login")
	head.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err := c.Post("https://www.crunchyroll.com/?a=formhandler", head, body)
	if err != nil {
		return anirip.NewError("Failed to execute authentication request", err)
	}
	return nil
}

// validateSession performs a get request on crunchyrolls homepage and checks
// to be sure a non-empty username is found
func validateSession(c *anirip.HTTPClient) error {
	resp, err := c.Get("http://www.crunchyroll.com/", nil)
	if err != nil {
		return anirip.NewError("Failed to execute session validation request", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return anirip.NewError("Failed to parse session validation page", err)
	}

	user := strings.TrimSpace(doc.Find("li.username").First().Text())
	if resp.StatusCode == 200 && user != "" {
		return nil
	}
	return anirip.NewError("Failed to verify session", nil)
}

package crunchyroll /* import "s32x.com/anirip/crunchyroll" */

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"s32x.com/anirip/common"
	"s32x.com/anirip/common/log"
)

// Login logs the user in to Crunchyroll and stores the session on the client
func Login(c *common.HTTPClient, user, pass string) error {
	// Perform preflight request to retrieve the login page
	res, err := c.Get("https://www.crunchyroll.com/login", nil)
	if err != nil {
		return fmt.Errorf("getting login page: %w", err)
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return fmt.Errorf("generating login document: %w", err)
	}

	// Scrape the login token
	token, _ := doc.Find("#login_form__token").First().Attr("value")

	// Sets the credentials and attempts to generate new cookies
	if err := createSession(c, user, pass, token); err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	// Validates the session created and returns
	if err := validateSession(c); err != nil {
		return fmt.Errorf("validating session: %w", err)
	}
	log.Info("Successfully logged in!")
	return nil
}

// createSession sends a login request to Crunchyroll and stores the cookies
// recieved in the clients cookiejar
func createSession(c *common.HTTPClient, user, pass, token string) error {
	// Assemble the FormData needed for authentication
	body := bytes.NewBufferString(url.Values{
		"login_form[name]":         {user},
		"login_form[password]":     {pass},
		"login_form[redirect_url]": {"/"},
		"login_form[_token]":       {token},
	}.Encode())

	// Execute the post request on the formhandler and store cookies in the jar
	head := http.Header{}
	head.Add("Referer", "https://www.crunchyroll.com/login")
	head.Add("Content-Type", "application/x-www-form-urlencoded")
	if _, err := c.Post("https://www.crunchyroll.com/login", head, body); err != nil {
		return fmt.Errorf("posting auth request: %w", err)
	}
	return nil
}

// validateSession performs a get request on crunchyrolls homepage and checks
// to be sure a non-empty username is found
func validateSession(c *common.HTTPClient) error {
	resp, err := c.Get("http://www.crunchyroll.com/", nil)
	if err != nil {
		return fmt.Errorf("getting validation page: %w", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return fmt.Errorf("generating validation document: %w", err)
	}

	user := strings.TrimSpace(doc.Find("li.username").First().Text())
	if resp.StatusCode == 200 && user != "" {
		return nil
	}
	return fmt.Errorf("could not verify session")
}

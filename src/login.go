package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type CrunchyCookie struct {
	User    string
	Pass    string
	Cookies []*http.Cookie
}

// Attempts to log the user in, store a cookie and return the login status
func login(crunchyCookie *CrunchyCookie) error {
	// First checks to see if we already have a cookie config
	exists, err := getStoredCookies(crunchyCookie)
	if err != nil {
		return err
	}

	// If we don't already have cookies, get new ones
	if crunchyCookie.Cookies == nil || crunchyCookie.User == "" {
		// Ask for credentials and request new cookies
		fmt.Printf("Please enter your username and password.\n\n")
		getStandardUserInput("Username : ", &crunchyCookie.User)
		getStandardUserInput("Password : ", &crunchyCookie.Pass)
		err := getNewCookies(crunchyCookie)
		if err != nil {
			return err
		}
	}

	// Test the cookies we currently have at this point
	valid, err := validateCookies(crunchyCookie)
	if err != nil {
		return err
	}

	// If the cookies we have are currently valid but dont exist, store them
	if valid && !exists {
		// Prepares a buffer and marshals the crunchyCookie object
		var crunchyCookieBytes bytes.Buffer
		crunchyCookieEncoder := gob.NewEncoder(&crunchyCookieBytes)
		err = crunchyCookieEncoder.Encode(crunchyCookie)
		if err != nil {
			return CRError{"There was an error encoding your cookies", err}
		}

		// Writes cookies to cookies file
		err := ioutil.WriteFile(cookiesFile, crunchyCookieBytes.Bytes(), 0644)
		if err != nil {
			return CRError{"There was an error writing cookies to file", err}
		}
		return nil
	}
	return nil
}

// Gets stored cookies found in cookiesFile
func getStoredCookies(crunchyCookie *CrunchyCookie) (bool, error) {
	// Checks if file exists - will return it's contents if so
	if _, err := os.Stat(cookiesFile); err == nil {
		crunchyCookieBytes, err := ioutil.ReadFile(cookiesFile)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(cookiesFile)
			return false, CRError{"There was an error reading your cookies file", err}
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		crunchyCookieBuffer := bytes.NewBuffer(crunchyCookieBytes)
		crunchyCookieDecoder := gob.NewDecoder(crunchyCookieBuffer)

		// Decodes the stored cookie and returns it
		err = crunchyCookieDecoder.Decode(&crunchyCookie)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(cookiesFile)
			return false, CRError{"There was an error decoding your cookies file", err}
		}
		// Cookies are able to be decoded so return true
		return true, nil
	}
	return false, nil
}

// Creates new cookies by re-authenticating with Crunchyroll
func getNewCookies(crunchyCookie *CrunchyCookie) error {
	// Construct formdata for the login request
	formData := url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {crunchyCookie.User},
		"password": {crunchyCookie.Pass},
	}

	// Prepare an http request to be modified
	loginReq, err := http.NewRequest("POST", "https://www.crunchyroll.com/?a=formhandler", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return CRError{"There was an error creating login request", err}
	}

	// Adds required headers to get a valid 200 response
	loginReq.Header.Add("referer", "https://www.crunchyroll.com/login")
	loginReq.Header.Add("user-agent", userAgent)
	loginReq.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Attempt to execute the login request
	loginResp, err := http.DefaultTransport.RoundTrip(loginReq)
	if err != nil {
		return CRError{"There was an error performing login request", err}
	}

	// Sets cookies to recieved cookies and returns
	crunchyCookie.Cookies = loginResp.Cookies()
	return nil
}

// Validates the cookies to be sure that we are still logged in
func validateCookies(crunchyCookie *CrunchyCookie) (bool, error) {
	// We use the cookie we recieved to attempt a simple authenticated request
	client := &http.Client{}
	verificationReq, err := http.NewRequest("GET", "http://www.crunchyroll.com/", nil)
	if err != nil {
		return false, CRError{"There was an error creating cookie validation request", err}
	}

	// Sets the headers for our (hopefully) authenticated request
	verificationReq.Header.Add("User-Agent", userAgent)
	verificationReq.Header.Add("Connection", "keep-alive")
	for i := 0; i < len(crunchyCookie.Cookies); i++ {
		verificationReq.AddCookie(crunchyCookie.Cookies[i])
	}

	// Attempt to execute the authenticated verification request
	validationResp, err := client.Do(verificationReq)
	if err != nil {
		return false, CRError{"There was an error performing cookie validation request", err}
	}
	defer validationResp.Body.Close()

	// If we see our username in the document, login was a success
	loginDoc, err := goquery.NewDocumentFromResponse(validationResp)
	if err != nil {
		return false, CRError{"There was an error parsing cookie validation page", err}
	}
	scannedUser := strings.TrimSpace(loginDoc.Find("li.username").First().Text())

	// Checks if the Username used to login is in the home page...
	if strings.ToLower(scannedUser) == strings.ToLower(crunchyCookie.User) {
		fmt.Printf(">> Logged in as " + crunchyCookie.User + "\n\n")
		return true, nil
	}
	return false, nil
}

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

type Session struct {
	User    string
	Pass    string
	Cookies []*http.Cookie
}

// Attempts to log the user in, store a cookie and return the login status
func (session *Session) Login(user string, pass string) error {
	// First checks to see if we already have a cookie config
	exists, err := getStoredCookies(session)
	if err != nil {
		return err
	}

	// If we don't already have cookies, get new ones
	if session.Cookies == nil || session.User == "" {
		// Ask for credentials and request new cookies
		session.User = user
		session.Pass = pass
		err := createNewCookies(session)
		if err != nil {
			return err
		}
	}

	// Test the cookies we currently have at this point
	valid, err := validateCookies(session)
	if err != nil {
		return err
	}

	// If the cookies we have are currently valid but dont exist, store them
	if valid && !exists {
		// Prepares a buffer and marshals the session object
		var sessionBytes bytes.Buffer
		sessionEncoder := gob.NewEncoder(&sessionBytes)
		err = sessionEncoder.Encode(session)
		if err != nil {
			return Error{"There was an error encoding your cookies", err}
		}

		// Writes cookies to cookies file
		err := ioutil.WriteFile("crunchyroll.cookie", sessionBytes.Bytes(), 0644)
		if err != nil {
			return Error{"There was an error writing cookies to file", err}
		}
		return nil
	}
	return nil
}

// Gets stored cookies found in cookiesFile
func getStoredCookies(session *Session) (bool, error) {
	// Checks if file exists - will return it's contents if so
	if _, err := os.Stat("crunchyroll.cookie"); err == nil {
		sessionBytes, err := ioutil.ReadFile("crunchyroll.cookie")
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove("crunchyroll.cookie")
			return false, Error{"There was an error reading your cookies file", err}
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		sessionBuffer := bytes.NewBuffer(sessionBytes)
		sessionDecoder := gob.NewDecoder(sessionBuffer)

		// Decodes the stored cookie and returns it
		err = sessionDecoder.Decode(&session)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove("crunchyroll.cookie")
			return false, Error{"There was an error decoding your cookies file", err}
		}
		// Cookies are able to be decoded so return true
		return true, nil
	}
	return false, nil
}

// Creates new cookies by re-authenticating with Crunchyroll
func createNewCookies(session *Session) error {
	// Construct formdata for the login request
	formData := url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {session.User},
		"password": {session.Pass},
	}

	// Prepare an http request to be modified
	loginReq, err := http.NewRequest("POST", "https://www.crunchyroll.com/?a=formhandler", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return Error{"There was an error creating login request", err}
	}

	// Adds required headers to get a valid 200 response
	loginReq.Header.Add("referer", "https://www.crunchyroll.com/login")
	loginReq.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	loginReq.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Attempt to execute the login request
	loginResp, err := http.DefaultTransport.RoundTrip(loginReq)
	if err != nil {
		return Error{"There was an error performing login request", err}
	}

	// Sets cookies to recieved cookies and returns
	session.Cookies = loginResp.Cookies()
	return nil
}

// Validates the cookies to be sure that we are still logged in
func validateCookies(session *Session) (bool, error) {
	// We use the cookie we recieved to attempt a simple authenticated request
	client := &http.Client{}
	verificationReq, err := http.NewRequest("GET", "http://www.crunchyroll.com/", nil)
	if err != nil {
		return false, Error{"There was an error creating cookie validation request", err}
	}

	// Sets the headers for our (hopefully) authenticated request
	verificationReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	verificationReq.Header.Add("Connection", "keep-alive")
	for _, cookie := range session.Cookies {
		verificationReq.AddCookie(cookie)
	}

	// Attempt to execute the authenticated verification request
	validationResp, err := client.Do(verificationReq)
	if err != nil {
		return false, Error{"There was an error performing cookie validation request", err}
	}
	defer validationResp.Body.Close()

	// If we see our username in the document, login was a success
	loginDoc, err := goquery.NewDocumentFromResponse(validationResp)
	if err != nil {
		return false, Error{"There was an error parsing cookie validation page", err}
	}
	scannedUser := strings.TrimSpace(loginDoc.Find("li.username").First().Text())

	// Checks if the Username used to login is in the home page...
	if strings.ToLower(scannedUser) == strings.ToLower(session.User) {
		fmt.Println("Logged in as " + session.User)
		return true, nil
	}
	return false, nil
}

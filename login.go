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
	Cookies []*http.Cookie
}

// Attempts to log the user in, store a cookie and return the login status
func login() ([]*http.Cookie, error) {
	user := ""
	pass := ""

	// First checks to see if we already have a cookie config
	user, crunchyCookie, exists, err := getStoredCookies()
	if err != nil {
		return nil, err
	}

	// If we don't already have cookies, get new ones
	if !exists || len(crunchyCookie.Cookies) == 0 || user == "" {
		// Ask for credentials and request new cookies
		fmt.Print("Please enter your username and password.\n\n")
		GetStandardUserInput("Username : ", &user)
		GetStandardUserInput("Password : ", &pass)

		crunchyCookie, err = getNewCookies(user, pass)
		if err != nil {
			return nil, err
		}
	}

	// Test the cookies we currently have to be sure they are still valid
	valid, err := validateCookies(user, crunchyCookie)
	if err != nil {
		return nil, err
	}

	if valid && exists {
		fmt.Printf(">> Logged in as " + crunchyCookie.User + "\n\n")
		return crunchyCookie.Cookies, nil
	}

	// If the cookies we have are currently valid,
	// and don't already exist, store them
	if valid && !exists {
		var crunchyCookieBytes bytes.Buffer
		crunchyCookieEncoder := gob.NewEncoder(&crunchyCookieBytes)
		err = crunchyCookieEncoder.Encode(crunchyCookie)
		if err != nil {
			return nil, err
		}

		ioutil.WriteFile(cookiesFile, crunchyCookieBytes.Bytes(), 0644)

		return crunchyCookie.Cookies, nil
	}
	return nil, nil
}

// Gets stored cookies found in cookiesFile
func getStoredCookies() (string, CrunchyCookie, bool, error) {
	// Checks if file exists - will return its contents if so
	if _, err := os.Stat(cookiesFile); err == nil {

		crunchyCookieBytes, err := ioutil.ReadFile(cookiesFile)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(cookiesFile)
			return "", CrunchyCookie{}, false, nil
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		crunchyCookieBuffer := bytes.NewBuffer(crunchyCookieBytes)
		crunchyCookieDecoder := gob.NewDecoder(crunchyCookieBuffer)

		// Decodes the stored cookie and returns it
		var crunchyCookie CrunchyCookie
		err = crunchyCookieDecoder.Decode(&crunchyCookie)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(cookiesFile)
			return "", CrunchyCookie{}, false, nil
		}
		return crunchyCookie.User, crunchyCookie, true, nil
	}
	return "", CrunchyCookie{}, false, nil
}

// Creates new cookies by re-authenticating with Crunchyroll
func getNewCookies(user string, pass string) (CrunchyCookie, error) {
	// Construct formdata for the login request
	formData := url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {user},
		"password": {pass},
	}

	// Prepare an http request to be modified
	loginReq, err := http.NewRequest("POST", "https://www.crunchyroll.com/?a=formhandler", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return CrunchyCookie{User: user}, err
	}

	// Adds required headers to get a valid 200 response
	loginReq.Header.Add("referer", "https://www.crunchyroll.com/login")
	loginReq.Header.Add("user-agent", userAgent)
	loginReq.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Attempt to execute the login request
	loginResp, err := http.DefaultTransport.RoundTrip(loginReq)
	if err != nil {
		return CrunchyCookie{User: user}, err
	}

	// Packs all our cookies into a CookieJar and returns it
	return CrunchyCookie{User: user, Cookies: loginResp.Cookies()}, nil
}

// Validates the cookies to be sure that we are still logged in
func validateCookies(user string, crunchyCookie CrunchyCookie) (bool, error) {
	// We use the cookie we recieved to attempt a simple authenticated request
	client := &http.Client{}
	verificationReq, err := http.NewRequest("GET", "http://www.crunchyroll.com/", nil)
	if err != nil {
		return false, err
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
		return false, err
	}
	defer validationResp.Body.Close()

	// If we see our username in the document, login was a success
	loginDoc, err := goquery.NewDocumentFromResponse(validationResp)
	if err != nil {
		return false, err
	}
	scannedUser := strings.TrimSpace(loginDoc.Find("li.username").First().Text())

	if strings.ToLower(scannedUser) == strings.ToLower(user) {
		return true, nil
	}
	return false, nil
}

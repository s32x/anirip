package daisuki

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/ANIRip/anirip"
)

type DaisukiSession struct {
	User    string
	Pass    string
	Cookies []*http.Cookie
}

// Attempts to log the user in, store a cookie and return the login status
func (session *DaisukiSession) Login(user string, pass string) error {
	// First checks to see if we already have a cookie config
	exists, err := getStoredCookies(session)
	if err != nil {
		return err
	}

	// If we don't already have cookies, get new ones
	if session.Cookies == nil || session.User == "" {
		// Sets the credentials and attempts to generate new cookies
		session.User = user
		session.Pass = pass
		if err := createNewCookies(session); err != nil {
			return err
		}
	}

	// Test the cookies we currently have at this point
	valid, err := validateCookies(session)
	if err != nil || !valid {
		return anirip.Error{Message: "Our Daisuki cookies are invalid", Err: err}
	}

	// If the cookies we have are currently valid but dont exist, store them
	if valid && !exists {
		// Prepares a buffer and marshals the session object
		var sessionBytes bytes.Buffer
		sessionEncoder := gob.NewEncoder(&sessionBytes)
		session.Pass = "" // Clears the password before writing it to our cookie file
		err = sessionEncoder.Encode(session)
		if err != nil {
			return anirip.Error{Message: "There was an error encoding your cookies", Err: err}
		}

		// Writes cookies to cookies file
		err := ioutil.WriteFile("daisuki.cookie", sessionBytes.Bytes(), 0644)
		if err != nil {
			return anirip.Error{Message: "There was an error writing cookies to file", Err: err}
		}
		return nil
	}
	return nil
}

// Gets the daisuki cookies so they can be passed accessed outside this lib
func (session *DaisukiSession) GetCookies() []*http.Cookie {
	return session.Cookies
}

// Gets stored cookies found in cookiesFile
func getStoredCookies(session *DaisukiSession) (bool, error) {
	// Checks if file exists - will return it's contents if so
	if _, err := os.Stat("daisuki.cookie"); err == nil {
		sessionBytes, err := ioutil.ReadFile("daisuki.cookie")
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove("daisuki.cookie")
			return false, anirip.Error{Message: "There was an error reading your cookies file", Err: err}
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		sessionBuffer := bytes.NewBuffer(sessionBytes)
		sessionDecoder := gob.NewDecoder(sessionBuffer)

		// Decodes the stored cookie and returns it
		err = sessionDecoder.Decode(&session)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove("daisuki.cookie")
			return false, anirip.Error{Message: "There was an error decoding your cookies file", Err: err}
		}
		// Cookies are able to be decoded so return true
		return true, nil
	}
	return false, nil
}

// Creates new cookies by re-authenticating with Daisuki
func createNewCookies(session *DaisukiSession) error {
	// Construct formdata for the login request
	formData := url.Values{
		"emailAddress": {session.User},
		"password":     {session.Pass},
	}

	// Performs the HTTP Request that will log the user in
	loginReqHeaders := http.Header{}
	loginReqHeaders.Add("referer", "http://www.daisuki.net/us/en/top.html")
	loginReqHeaders.Add("content-type", "application/x-www-form-urlencoded")
	loginResponse, err := anirip.GetHTTPResponse("POST",
		"https://www.daisuki.net/bin/SignInServlet.html/input",
		bytes.NewBufferString(formData.Encode()),
		loginReqHeaders,
		[]*http.Cookie{})
	if err != nil {
		return err
	}

	// Sets cookies to recieved cookies and returns
	session.Cookies = loginResponse.Cookies()
	return nil
}

// Validates the cookies to be sure that we are still logged in
func validateCookies(session *DaisukiSession) (bool, error) {
	// We use the cookie we recieved to attempt a simple authenticated request
	validationReqHeaders := http.Header{}
	validationReqHeaders.Add("referer", "http://www.daisuki.net/us/en/mypage/info.html")
	validationResponse, err := anirip.GetHTTPResponse("GET",
		"https://www.daisuki.net/us/en/mypage/info.html",
		nil,
		validationReqHeaders,
		session.Cookies)
	if err != nil {
		return false, err
	}

	// Creates a goquery document for scraping
	validationRespDoc, err := goquery.NewDocumentFromResponse(validationResponse)
	if err != nil {
		return false, anirip.Error{Message: "There was an error while accessing the validation page", Err: err}
	}

	// Scrapes the document and attempts to find the username
	userName := validationRespDoc.Find("div#Nickname.clearFix.accountInformation div.list02").First().Text()

	// Checks if the Username used to login is in the account info page...
	if validationResponse.StatusCode == 200 && userName != "" {
		fmt.Printf("Logged in to Daisuki as " + userName + "\n\n")
		return true, nil
	}
	return false, nil
}

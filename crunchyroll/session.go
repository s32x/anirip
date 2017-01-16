package crunchyroll

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/anirip/anirip"
)

// Session contains user login credentials OR active session cookies
type Session struct {
	User    string
	Pass    string
	Cookies []*http.Cookie
}

// Login logs the user in and stores the cookies in a file in our temp folder
// for use on future requests
func (s *Session) Login(user, pass, tempDir string) error {
	// First checks to see if we already have a cookie config
	exists, err := getStoredCookies(s, tempDir)
	if err != nil {
		return err
	}

	// If we don't already have cookies, get new ones
	if s.Cookies == nil || s.User == "" {
		// Sets the credentials and attempts to generate new cookies
		s.User = user
		s.Pass = pass
		err := createNewCookies(s)
		if err != nil {
			return err
		}
	}

	// Test the cookies we currently have at this point
	valid, err := validateCookies(s)
	if err != nil || !valid {
		return anirip.Error{Message: "Our Crunchyroll cookies are invalid", Err: err}
	}

	// If the cookies we have are currently valid but dont exist, store them
	if valid && !exists {
		// Prepares a buffer and marshals the session object
		var sessionBytes bytes.Buffer
		sessionEncoder := gob.NewEncoder(&sessionBytes)
		s.Pass = "" // Clears the password before writing it to our cookie file
		err = sessionEncoder.Encode(s)
		if err != nil {
			return anirip.Error{Message: "There was an error encoding your cookies", Err: err}
		}

		// Writes cookies to cookies file
		err := ioutil.WriteFile(tempDir+string(os.PathSeparator)+"crunchyroll.cookie", sessionBytes.Bytes(), 0644)
		if err != nil {
			return anirip.Error{Message: "There was an error writing cookies to file", Err: err}
		}
		return nil
	}
	return nil
}

// Returns the cookies so we can access them outside of this lib
func (s *Session) GetCookies() []*http.Cookie {
	return s.Cookies
}

// Gets stored cookies found in cookiesFile
func getStoredCookies(session *Session, tempDir string) (bool, error) {
	// Checks if file exists - will return it's contents if so
	if _, err := os.Stat(tempDir + string(os.PathSeparator) + "crunchyroll.cookie"); err == nil {
		sessionBytes, err := ioutil.ReadFile(tempDir + string(os.PathSeparator) + "crunchyroll.cookie")
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(tempDir + string(os.PathSeparator) + "crunchyroll.cookie")
			return false, anirip.Error{Message: "There was an error reading your cookies file", Err: err}
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		sessionBuffer := bytes.NewBuffer(sessionBytes)
		sessionDecoder := gob.NewDecoder(sessionBuffer)

		// Decodes the stored cookie and returns it
		err = sessionDecoder.Decode(&session)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(tempDir + string(os.PathSeparator) + "crunchyroll.cookie")
			return false, anirip.Error{Message: "There was an error decoding your cookies file", Err: err}
		}
		// Cookies are able to be decoded so return true
		return true, nil
	}
	return false, nil
}

// createNewCookies sends a login request to crunchyroll and stores the cookies recieved
func createNewCookies(s *Session) error {
	form := url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {s.User},
		"password": {s.Pass},
	}

	head := http.Header{}
	head.Add("referer", "https://www.crunchyroll.com/login")
	head.Add("content-type", "application/x-www-form-urlencoded")
	resp, err := anirip.GetHTTPResponse("POST", "https://www.crunchyroll.com/?a=formhandler",
		bytes.NewBufferString(form.Encode()), head, []*http.Cookie{})
	if err != nil {
		return err
	}

	s.Cookies = resp.Cookies()
	return nil
}

// validateCookies performs a get request on crunchyrolls homepage and checks
// to be sure a non-empty username is found
func validateCookies(s *Session) (bool, error) {
	head := http.Header{}
	head.Add("Connection", "keep-alive")
	resp, err := anirip.GetHTTPResponse("GET", "http://www.crunchyroll.com/", nil, head, s.Cookies)
	if err != nil {
		return false, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return false, anirip.Error{Message: "There was an error parsing cookie validation page", Err: err}
	}
	user := strings.TrimSpace(doc.Find("li.username").First().Text())

	if resp.StatusCode == 200 && user != "" {
		return true, nil
	}
	return false, nil
}

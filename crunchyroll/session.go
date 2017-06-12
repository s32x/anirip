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
	User, Pass string
	Cookies    []*http.Cookie
}

// Login logs the user in and stores the cookies in a file in our temp folder
// for use on future requests
func (s *Session) Login(client *anirip.HTTPClient, user, pass, tempDir string) error {
	// First checks to see if we already have a cookie config
	exists, err := getStoredCookies(s, tempDir)
	if err != nil {
		return anirip.NewError("Failed to retrieve stored cookies", err)
	}

	// If we don't already have cookies, get new ones
	if s.Cookies == nil || s.User == "" {
		// Sets the credentials and attempts to generate new cookies
		s.User = user
		s.Pass = pass
		if err := createNewCookies(client, s); err != nil {
			return anirip.NewError("Failed to create new cookies", err)
		}
	}

	// Test the cookies we currently have at this point
	valid, err := validateCookies(client, s)
	if err != nil || !valid {
		return anirip.NewError("Your Crunchyroll cookies are invalid", err)
	}

	// If the cookies we have are currently valid but dont exist, store them
	if valid && !exists {
		// Prepares a buffer and marshals the session object
		var sessionBytes bytes.Buffer
		sessionEncoder := gob.NewEncoder(&sessionBytes)
		s.Pass = "" // Clears the password before writing it to our cookie file
		if err := sessionEncoder.Encode(s); err != nil {
			return anirip.NewError("There was an error encoding your cookies", err)
		}

		// Writes cookies to cookies file
		if err := ioutil.WriteFile(tempDir+string(os.PathSeparator)+"crunchyroll.cookie", sessionBytes.Bytes(), 0644); err != nil {
			return anirip.NewError("There was an error writing cookies to file", err)
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
			os.Remove(tempDir + string(os.PathSeparator) + "crunchyroll.cookie")
			return false, anirip.NewError("There was an error reading your cookies file", err)
		}

		// Creates a decoder to decode the bytes found in our cookiesFile
		sessionBuffer := bytes.NewBuffer(sessionBytes)
		sessionDecoder := gob.NewDecoder(sessionBuffer)

		// Decodes the stored cookie and returns it
		err = sessionDecoder.Decode(&session)
		if err != nil {
			// Attempts a deletion of an unreadable cookies file
			_ = os.Remove(tempDir + string(os.PathSeparator) + "crunchyroll.cookie")
			return false, anirip.NewError("There was an error decoding your cookies file", err)
		}
		// Cookies are able to be decoded so return true
		return true, nil
	}
	return false, nil
}

// createNewCookies sends a login request to crunchyroll and stores the cookies recieved
func createNewCookies(client *anirip.HTTPClient, s *Session) error {
	body := bytes.NewBufferString(url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {s.User},
		"password": {s.Pass},
	}.Encode())

	client.Header.Add("referer", "https://www.crunchyroll.com/login")
	client.Header.Add("connection-type", "application/x-www-form-urlencoded")
	resp, err := client.Post("https://www.crunchyroll.com/?a=formhandler", body)
	if err != nil {
		return err
	}
	s.Cookies = resp.Cookies()
	return nil
}

// validateCookies performs a get request on crunchyrolls homepage and checks
// to be sure a non-empty username is found
func validateCookies(client *anirip.HTTPClient, s *Session) (bool, error) {
	client.Header.Add("connection", "keep-alive")
	resp, err := client.Get("http://www.crunchyroll.com/")
	if err != nil {
		return false, anirip.NewError("Failed to validate cookies", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return false, anirip.NewError("There was an error parsing cookie validation page", err)
	}

	user := strings.TrimSpace(doc.Find("li.username").First().Text())
	if resp.StatusCode == 200 && user != "" {
		return true, nil
	}
	return false, nil
}

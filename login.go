package main

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/kr/pretty"
)

// Function that attempts to log the user in, store a
// cookie and return the login status
func login(user string, pass string) (bool, error) {
	// Construct formdata for the login request
	formData := url.Values{
		"formname": {"RpcApiUser_Login"},
		"fail_url": {"http://www.crunchyroll.com/login"},
		"name":     {user},
		"password": {pass},
	}

	// Prepare a client & http request to be modified
	client := &http.Client{}
	loginReq, err := http.NewRequest("POST", "https://www.crunchyroll.com/?a=formhandler", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return false, err
	}

	// Set the required headers so we look like a browser
	loginReq.Header.Add("Referer", "https://www.crunchyroll.com/login")
	loginReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	loginReq.Header.Add("Content-type", "application/x-www-form-urlencoded")

	// Attempt to execute the login request
	loginResp, err := client.Do(loginReq)
	if err != nil {
		return false, err
	}
	defer loginResp.Body.Close()

	pretty.Println(loginResp.Cookies())
	pretty.Println(loginResp.Header)

	// Now we use the cookie we recieved to attempt a simple authenticated request
	// client = &http.Client{}
	// verificationReq, err := http.NewRequest("GET", "http://www.crunchyroll.com/", nil)
	// if err != nil {
	// 	return false, err
	// }
	//
	// // Sets the headers for our (hopefully) authenticated request
	// verificationReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	// verificationReq.Header.Add("Connection", "keep-alive")
	//
	// pretty.Println(verificationReq)
	//
	// // Attempt to execute the verification authenticated request
	// verificationResp, err := client.Do(verificationReq)
	// if err != nil {
	// 	return false, err
	// }
	// defer loginResp.Body.Close()
	// body, err := ioutil.ReadAll(verificationResp.Body)
	// if err != nil {
	// 	return false, err
	// }
	//
	// pretty.Println(string(body))
	//
	// if strings.Contains(string(body), user) {
	// 	log.Println("SUCCESS!")
	// 	return true, nil
	// }
	// log.Println("FAIL")
	return false, nil
}

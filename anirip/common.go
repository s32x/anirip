package anirip

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// A rename func that retries 10 times before returning an error
func Rename(sourcesFile, destinationFile string, i int) error {
	// Attempts a rename and if it fails, it will retry i times
	if err := os.Rename(sourcesFile, destinationFile); err != nil {
		if i > 0 {
			Rename(sourcesFile, destinationFile, i-1)
		}
		return Error{Message: "There was an error renaming " + sourcesFile + " to " + destinationFile, Err: err}
	}
	return nil
}

// A shorthand function for writing http requests. NOTE: Uses a default user-agent for every request
func GetHTTPResponse(method, urlStr string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	// Builds out request based on first 3 params
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, Error{Message: "There was an error creating the " + method + " request on " + urlStr, Err: err}
	}

	// Sets the headers passed as the request headers
	request.Header = header
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")

	// Attaches all cookies passed
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	// Executes the new request and returns the response, retries recursively if theres a failure
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		fmt.Println("There was an error performing the " + method + " request on " + urlStr + " : " + err.Error() + "\n")
		GetHTTPResponse(method, urlStr, body, header, cookies)
	}
	return response, nil
}

// Constructs an episode file name and returns the file name cleaned
func GenerateEpisodeFileName(showTitle string, seasonNumber int, episodeNumber float64, description string) string {
	// Pads season number with a 0 if it's less than 10
	seasonNumberString := strconv.Itoa(seasonNumber)
	if seasonNumber < 10 {
		seasonNumberString = "0" + strconv.Itoa(seasonNumber)
	}

	// Pads episode number with a 0 if it's less than 10
	episodeNumberString := strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	if episodeNumber < 10 {
		episodeNumberString = "0" + strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	}

	// Constructs episode file name and returns it
	fileName := showTitle + " - S" + seasonNumberString + "E" + episodeNumberString + " - " + description
	return CleanFileName(fileName)
}

// Cleans the new file/folder name so there won't be any write issues
func CleanFileName(fileName string) string {
	newFileName := fileName // Strips out any illegal characters and returns our new file name
	for _, illegalChar := range []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"} {
		newFileName = strings.Replace(newFileName, illegalChar, " ", -1)
	}
	return newFileName
}

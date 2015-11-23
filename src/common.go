package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Error struct {
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf(">>> Error : %v : %v", e.Message, e.Err)
	}
	return fmt.Sprintf(">>> Error : %v.", e.Message)
}

// Gets XML data for the requested request type and episode
func getXML(req string, episode *Episode, cookies []*http.Cookie) (string, error) {
	xmlUrl := "http://www.crunchyroll.com/xml/?"

	// formdata to indicate the source page
	formData := url.Values{
		"current_page": {episode.URL},
	}

	// Constructs a queryString for user set settings
	queryString := url.Values{}
	if req == "RpcApiSubtitle_GetXml" {
		queryString = url.Values{
			"req":                {"RpcApiSubtitle_GetXml"},
			"subtitle_script_id": {strconv.Itoa(episode.SubtitleID)},
		}
	} else if req == "RpcApiVideoPlayer_GetStandardConfig" {
		queryString = url.Values{
			"req":           {"RpcApiVideoPlayer_GetStandardConfig"},
			"media_id":      {strconv.Itoa(episode.ID)},
			"video_format":  {"108"},
			"video_quality": {"80"},
			"auto_play":     {"1"},
			"aff":           {"crunchyroll-website"},
			"show_pop_out_controls":   {"1"},
			"pop_out_disable_message": {""},
			"click_through":           {"0"},
		}
	} else {
		queryString = url.Values{
			"req":                  {req},
			"media_id":             {strconv.Itoa(episode.ID)},
			"video_format":         {"108"},
			"video_encode_quality": {"80"},
		}
	}

	// Constructs a client and request that will get the xml we're asking for
	client := &http.Client{}
	xmlReq, err := http.NewRequest("POST", xmlUrl+queryString.Encode(), bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", Error{"There was an error creating our getXML request", err}
	}
	xmlReq.Header.Add("Host", "www.crunchyroll.com")
	xmlReq.Header.Add("Origin", "http://static.ak.crunchyroll.com")
	xmlReq.Header.Add("Content-type", "application/x-www-form-urlencoded")
	xmlReq.Header.Add("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
	xmlReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	xmlReq.Header.Add("X-Requested-With", "ShockwaveFlash/19.0.0.245")
	for i := 0; i < len(cookies); i++ {
		xmlReq.AddCookie(cookies[i])
	}

	// Executes request and returns the result as a string
	resp, err := client.Do(xmlReq)
	if err != nil {
		return "", Error{"There was an error executing our getXML request", err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", Error{"There was an error reading our getXML response", err}
	}
	return string(body), nil
}

// Gets user input from the user and unmarshalls it into the input
func getStandardUserInput(prefixText string, input *string) error {
	fmt.Printf(prefixText)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		*input = scanner.Text()
		break
	}
	if err := scanner.Err(); err != nil {
		return Error{"There was an error getting standard user input", err}
	}
	return nil
}

// Constructs and cleans the file name that we will assign for the episode
func generateEpisodeFileName(showTitle string, seasonNumber int, episodeNumber float64, description string) string {
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
	newFileName := showTitle + " - S" + seasonNumberString + "E" + episodeNumberString + " - " + description

	// Strips out any illegal characters and returns our new file name
	for _, illegalChar := range []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"} {
		newFileName = strings.Replace(newFileName, illegalChar, " ", -1)
	}
	return newFileName
}

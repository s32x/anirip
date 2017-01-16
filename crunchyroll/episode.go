package crunchyroll

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/anirip/anirip"
)

// Episode holds all episode metadata desired for downloading
type Episode struct {
	ID          int
	SubtitleID  int
	Title       string
	Description string
	Number      float64
	Quality     string
	Path        string
	URL         string
	FileName    string
	StreamURL   string
}

// Parses the xml and returns what we need from the xml
func (e *Episode) GetEpisodeInfo(quality string, cookies []*http.Cookie) error {
	e.Quality = quality // Sets the quality to the passed quality string

	// Gets the HTML of the episode page
	head := http.Header{}
	head.Add("referer", "http://www.crunchyroll.com/"+strings.Split(e.Path, "/")[1])
	resp, err := anirip.GetHTTPResponse("GET", e.URL, nil, head, cookies)
	if err != nil {
		return err
	}

	// Creates the goquery document that will be used to scrape for episode metadata
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return anirip.Error{Message: "There was an error while reading the episode doc", Err: err}
	}

	// Formdata to indicate the source page
	formData := url.Values{
		"current_page": {e.URL},
	}

	// Querystring for getting the crunchyroll standard config
	queryString := url.Values{
		"req":           {"RpcApiVideoPlayer_GetStandardConfig"},
		"media_id":      {strconv.Itoa(e.ID)},
		"video_format":  {getVideoFormat(e.Quality)},
		"video_quality": {getVideoQuality(e.Quality)},
		"auto_play":     {"1"},
		"aff":           {"crunchyroll-website"},
		"show_pop_out_controls":   {"1"},
		"pop_out_disable_message": {""},
		"click_through":           {"0"},
	}

	// Performs the HTTP Request that will get the XML
	head = http.Header{}
	head.Add("Host", "www.crunchyroll.com")
	head.Add("Origin", "http://static.ak.crunchyroll.com")
	head.Add("Content-type", "application/x-www-form-urlencoded")
	head.Add("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.f3770232.swf")
	head.Add("X-Requested-With", "ShockwaveFlash/22.0.0.192")
	resp, err = anirip.GetHTTPResponse("POST", "http://www.crunchyroll.com/xml/?"+queryString.Encode(),
		bytes.NewBufferString(formData.Encode()), head, cookies)
	if err != nil {
		return err
	}

	// Gets the xml string from the recieved xml response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return anirip.Error{Message: "There was an error reading the xml response", Err: err}
	}
	xmlString := string(respBody)

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return anirip.Error{Message: "This video is not available in your region", Err: err}
		}
	}

	// Same type of xml parsing to get the file
	eFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		eFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return anirip.Error{Message: "No hosts were found for the episode", Err: err}
	}

	e.Title = strings.Replace(strings.Replace(doc.Find("#showmedia_about_name").First().Text(), "“", "", -1), "”", "", -1)
	e.FileName = anirip.CleanFileName(e.FileName + e.Title) // Updates filename with title that we just scraped
	e.StreamURL = strings.Replace(eFile, "amp;", "", -1)
	return nil
}

// Downloads entire FLV episodes to our temp directory
func (e *Episode) DownloadEpisode(quality, tempDir string, cookies []*http.Cookie) error {
	// Attempts to dump the FLV of the episode to file / will retry up to 5 times
	err := e.dumpEpisodeFLV(tempDir)
	if err != nil {
		return err
	}

	// Finally renames the dumped FLV to an MKV
	if err := anirip.Rename(tempDir+string(os.PathSeparator)+"incomplete.episode.mkv", tempDir+string(os.PathSeparator)+"episode.mkv", 10); err != nil {
		return err
	}
	return nil
}

// Gets the filename of the episode for referencing outside of this lib
func (e *Episode) GetFileName() string {
	return e.FileName
}

// Calls rtmpdump.exe to dump the episode and names it
func (e *Episode) dumpEpisodeFLV(tempDir string) error {
	// Remove stale temp file to avoid conflcts with CLI
	os.Remove(tempDir + string(os.PathSeparator) + "incomplete.episode.mkv")

	// Executes the command which will be used to dump the episode
	cmd := exec.Command(anirip.FindAbsoluteBinary("ffmpeg"),
		"-i", e.StreamURL,
		"-c", "copy",
		"incomplete.episode.mkv")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return anirip.Error{Message: "There was an error while starting the rtmpdump command...", Err: err}
	}

	return nil
}

// Figures out what the format of the video should be based on crunchyroll xml
func getVideoFormat(quality string) string {
	switch format := strings.ToLower(quality); {
	case strings.Contains(format, "android"):
		return "107"
	case strings.Contains(format, "360"):
		return "106"
	case strings.Contains(format, "480"):
		return "106"
	case strings.Contains(format, "720"):
		return "106"
	case strings.Contains(format, "1080"):
		return "108"
	case strings.Contains(format, "default"):
		return "0"
	default:
		return "0"
	}
}

// Figures out what the resolution/quality should be based on crunchyroll xml
func getVideoQuality(quality string) string {
	switch resolution := strings.ToLower(quality); {
	case strings.Contains(resolution, "android"):
		return "71"
	case strings.Contains(resolution, "360"):
		return "60"
	case strings.Contains(resolution, "480"):
		return "61"
	case strings.Contains(resolution, "720"):
		return "62"
	case strings.Contains(resolution, "1080"):
		return "80"
	case strings.Contains(resolution, "default"):
		return "0"
	default:
		return "0"
	}
}

package crunchyroll

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/ANIRip/anirip"
)

type EpisodeMetaData struct {
	NoShow           string `json:"noShow"`
	Class            string `json:"class"`
	MediaID          string `json:"media_id"`
	CollectionID     string `json:"collection_id"`
	SeriesID         string `json:"series_id"`
	MediaType        string `json:"media_type"`
	EpisodeNumber    string `json:"episode_number"`
	Clip             bool   `json:"clip"`
	URL              string `json:"url"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ScreenshotImage  string `json:"screenshot_image"`
	Available        bool   `json:"available"`
	PremiumAvailable bool   `json:"premium_available"`
	FreeAvailable    bool   `json:"free_available"`
}

// Parses the xml and returns what we need from the xml
func (episode *CrunchyrollEpisode) GetEpisodeInfo(quality string, cookies []*http.Cookie) error {
	episode.Quality = quality // Sets the quality to the passed quality string

	// Gets the HTML of the episode page
	episodeReqHeaders := http.Header{}
	episodeReqHeaders.Add("referer", "http://www.crunchyroll.com/"+strings.Split(episode.Path, "/")[1])
	episodeResponse, err := anirip.GetHTTPResponse("GET",
		episode.URL,
		nil,
		episodeReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Creates the goquery document that will be used to scrape for episode metadata
	episodeDoc, err := goquery.NewDocumentFromResponse(episodeResponse)
	if err != nil {
		return anirip.Error{Message: "There was an error while reading the episode doc", Err: err}
	}

	// Scrapes the episode metadata from the episode page
	episodeMetaDataJSON := episodeDoc.Find("script#liftigniter-metadata").First().Text()

	// Parses the metadata json to a MetaData object
	episodeMetaData := new(EpisodeMetaData)
	if err := json.Unmarshal([]byte(episodeMetaDataJSON), episodeMetaData); err != nil {
		return anirip.Error{Message: "There was an error while parsing episode metadata", Err: err}
	}

	// Formdata to indicate the source page
	formData := url.Values{
		"current_page": {episode.URL},
	}

	// Querystring for getting the crunchyroll standard config
	queryString := url.Values{
		"req":           {"RpcApiVideoPlayer_GetStandardConfig"},
		"media_id":      {strconv.Itoa(episode.ID)},
		"video_format":  {getVideoFormat(episode.Quality)},
		"video_quality": {getVideoQuality(episode.Quality)},
		"auto_play":     {"1"},
		"aff":           {"crunchyroll-website"},
		"show_pop_out_controls":   {"1"},
		"pop_out_disable_message": {""},
		"click_through":           {"0"},
	}

	// Performs the HTTP Request that will get the XML
	standardConfigReqHeaders := http.Header{}
	standardConfigReqHeaders.Add("Host", "www.crunchyroll.com")
	standardConfigReqHeaders.Add("Origin", "http://static.ak.crunchyroll.com")
	standardConfigReqHeaders.Add("Content-type", "application/x-www-form-urlencoded")
	standardConfigReqHeaders.Add("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
	standardConfigReqHeaders.Add("X-Requested-With", "ShockwaveFlash/19.0.0.245")
	standardConfigResponse, err := anirip.GetHTTPResponse("POST",
		"http://www.crunchyroll.com/xml/?"+queryString.Encode(),
		bytes.NewBufferString(formData.Encode()),
		standardConfigReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Gets the xml string from the recieved xml response body
	standardConfigResponseBody, err := ioutil.ReadAll(standardConfigResponse.Body)
	if err != nil {
		return anirip.Error{Message: "There was an error reading the xml response", Err: err}
	}
	xmlString := string(standardConfigResponseBody)

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return anirip.Error{Message: "This video is not available in your region", Err: err}
		}
	}

	// Next performs some really basic parsing of the host url
	xmlHostURL := ""
	if strings.Contains(xmlString, "<host>") && strings.Contains(xmlString, "</host>") {
		xmlHostURL = strings.SplitN(strings.SplitN(xmlString, "<host>", 2)[1], "</host>", 2)[0]
	} else {
		return anirip.Error{Message: "No hosts were found for the episode", Err: err}
	}

	// Same type of xml parsing to get the file
	episodeFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		episodeFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return anirip.Error{Message: "No hosts were found for the episode", Err: err}
	}

	// Parses the URL in order to break out the two urls required for dumping
	url, err := url.Parse(xmlHostURL)
	if err != nil {
		return anirip.Error{Message: "There was an error parsing episode information", Err: err}
	}

	// Sets the RTMP info recieved before returning
	episode.Title = episodeMetaData.Name
	episode.FileName = anirip.CleanFileName(episode.FileName + episode.Title) // Updates filename with title that we just scraped
	episode.MediaInfo = RTMPInfo{
		File:   episodeFile,
		URLOne: url.Scheme + "://" + url.Host + url.Path,
		URLTwo: strings.Trim(url.RequestURI(), "/"),
	}
	return nil
}

// Downloads entire FLV episodes to our temp directory
func (episode *CrunchyrollEpisode) DownloadEpisode(quality, engineDir, tempDir string, cookies []*http.Cookie) error {
	// Attempts to dump the FLV of the episode to file
	err := episode.dumpEpisodeFLV(engineDir, tempDir)
	if err != nil {
		return err
	}

	// Finally renames the dumped FLV to an MKV
	if err := anirip.Rename(tempDir+"\\"+episode.FileName+".flv", tempDir+"\\"+episode.FileName+".mkv", 10); err != nil {
		return err
	}
	return nil
}

// Gets the filename of the episode for referencing outside of this lib
func (episode *CrunchyrollEpisode) GetFileName() string {
	return episode.FileName
}

// Calls rtmpdump.exe to dump the episode and names it
func (episode *CrunchyrollEpisode) dumpEpisodeFLV(engineDir, tempDir string) error {
	// Gets the path of our rtmp dump exe
	path, err := filepath.Abs(engineDir + "\\rtmpdump.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find rtmpdump.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Creates the command which we will use to dump the episode
	cmd := exec.Command(path,
		"-r", episode.MediaInfo.URLOne,
		"-a", episode.MediaInfo.URLTwo,
		"-f", "WIN 19,0,0,245",
		"-W", "http://static.ak.crunchyroll.com/versioned_assets/ChromelessPlayerApp.6282d5bd.swf",
		"-m", "10",
		"-p", episode.URL,
		"-y", episode.MediaInfo.File,
		"-o", episode.FileName+".flv")
	cmd.Dir = tempDir // Sets working directory to temp so our fragments end up there

	// Append retry param if the file already exists
	_, err = filepath.Abs(tempDir + "\\" + episode.FileName + ".flv")
	if err == nil {
		cmd.Args = append(cmd.Args, "-e")
	}

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		// Recursively recalls dempEpisodeFLV if we get an unfinished download
		episode.dumpEpisodeFLV(engineDir, tempDir)
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
	case strings.Contains(format, "highest"):
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
	case strings.Contains(resolution, "highest"):
		return "0"
	default:
		return "0"
	}
}

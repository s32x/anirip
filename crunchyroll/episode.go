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

	"github.com/sdwolfe32/ANIRip/anirip"
)

type RTMPInfo struct {
	URLOne string
	URLTwo string
	File   string
}

// Downloads entire FLV episodes to our temp directory
func (episode *CrunchyrollEpisode) DownloadEpisode(quality string, cookies []*http.Cookie) error {
	// First attempts to get the XML attributes for the requested episode
	rtmpInfo := new(RTMPInfo)
	err := episode.getEpisodeRTMPInfo(quality, rtmpInfo, cookies)
	if err != nil {
		return err
	}

	// Checks to see if the episode already exists, in which case we return os.stat error
	_, err = os.Stat("temp\\" + episode.FileName + ".mkv")
	if err == nil {
		return anirip.Error{Message: "This video has already been downloaded...", Err: err}
	}

	// Attempts to dump the FLV of the episode to file
	err = episode.dumpEpisodeFLV(rtmpInfo)
	if err != nil {
		return err
	}

	// Finally renames the dumped FLV to an MKV
	os.Rename("temp\\"+episode.FileName+".flv", "temp\\"+episode.FileName+".mkv")
	return nil
}

// Gets the filename of the episode for referencing outside of this lib
func (episode *CrunchyrollEpisode) GetFileName() string {
	return episode.FileName
}

// Parses the xml and returns what we need from the xml
func (episode *CrunchyrollEpisode) getEpisodeRTMPInfo(quality string, rtmpInfo *RTMPInfo, cookies []*http.Cookie) error {
	// First gets the XML of the episode video
	episode.Quality = quality // Sets the quality to the passed quality string

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
	rtmpInfo.File = episodeFile
	rtmpInfo.URLOne = url.Scheme + "://" + url.Host + url.Path
	rtmpInfo.URLTwo = strings.Trim(url.RequestURI(), "/")
	return nil
}

// Calls rtmpdump.exe to dump the episode and names it
func (episode *CrunchyrollEpisode) dumpEpisodeFLV(rtmpInfo *RTMPInfo) error {
	// Gets the path of our rtmp dump exe
	path, err := exec.LookPath("engine\\rtmpdump.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find rtmpdump.exe in \\engine\\ directory", Err: err}
	}

	// Creates the command which we will use to dump the episode
	cmd := exec.Command(path,
		"-r", rtmpInfo.URLOne,
		"-a", rtmpInfo.URLTwo,
		"-f", "WIN 19,0,0,245",
		"-W", "http://static.ak.crunchyroll.com/versioned_assets/ChromelessPlayerApp.6282d5bd.swf",
		"-m", "10",
		"-p", episode.URL,
		"-y", rtmpInfo.File,
		"-o", "temp\\"+episode.FileName+".flv")

	// Append retry param if the file already exists
	_, err = exec.LookPath("temp\\" + episode.FileName + ".flv")
	if err == nil {
		cmd.Args = append(cmd.Args, "-e")
	}

	// Executes the dump command and gets the episode
	err = cmd.Start()
	if err != nil {
		return anirip.Error{Message: "There was an error trying to execute our dumper", Err: err}
	}

	// Waits for the download to complete
	err = cmd.Wait()
	if err != nil {
		// Recursively recalls dempEpisodeFLV if we get an unfinished download
		episode.dumpEpisodeFLV(rtmpInfo)
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

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type RTMPInfo struct {
	URLOne string
	URLTwo string
	File   string
}

// Downloads entire FLV episodes to our temp directory
func (episode *Episode) DownloadEpisode(quality string, cookies []*http.Cookie) error {
	// First attempts to get the XML attributes for the requested episode
	rtmpInfo := new(RTMPInfo)
	err := episode.getEpisodeRTMPInfo(quality, rtmpInfo, cookies)
	if err != nil {
		return err
	}

	// Attempts to dump the FLV of the episode to file
	err = episode.dumpEpisodeFLV(rtmpInfo)
	if err != nil {
		return err
	}
	return nil
}

// Parses the xml and returns what we need from the xml
func (episode *Episode) getEpisodeRTMPInfo(quality string, rtmpInfo *RTMPInfo, cookies []*http.Cookie) error {
	// First gets the XML of the episode video
	episode.Quality = quality // Sets the quality to the passed quality string
	xmlString, err := getXML("RpcApiVideoPlayer_GetStandardConfig", episode, cookies)
	if err != nil {
		return err
	}

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return Error{"This video is not available in your region", err}
		}
	}

	// Next performs some really basic parsing of the host url
	xmlHostURL := ""
	if strings.Contains(xmlString, "<host>") && strings.Contains(xmlString, "</host>") {
		xmlHostURL = strings.SplitN(strings.SplitN(xmlString, "<host>", 2)[1], "</host>", 2)[0]
	} else {
		return Error{"No hosts were found for the episode", err}
	}

	// Same type of xml parsing to get the file
	episodeFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		episodeFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return Error{"No hosts were found for the episode", err}
	}

	// Parses the URL in order to break out the two urls required for dumping
	url, err := url.Parse(xmlHostURL)
	if err != nil {
		return Error{"There was an error parsing episode information", err}
	}

	// Sets the RTMP info recieved before returning
	rtmpInfo.File = episodeFile
	rtmpInfo.URLOne = url.Scheme + "://" + url.Host + url.Path
	rtmpInfo.URLTwo = strings.Trim(url.RequestURI(), "/")
	return nil
}

// Calls rtmpdump.exe to dump the episode and names it
func (episode *Episode) dumpEpisodeFLV(rtmpInfo *RTMPInfo) error {
	// Gets the path of our rtmp dump exe
	path, err := exec.LookPath("engine\\rtmpdump.exe")
	if err != nil {
		return Error{"Unable to find rtmpdump.exe in \\engine\\ directory", err}
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
		return Error{"There was an error trying to execute our dumper", err}
	}

	// Waits for the download to complete
	err = cmd.Wait()
	if err != nil {
		// Recursively recalls dempEpisodeFLV if we get an unfinished download
		episode.dumpEpisodeFLV(rtmpInfo)
	}
	return nil
}

// Gets XML data for the requested request type and episode
func getXML(req string, episode *Episode, cookies []*http.Cookie) (string, error) {
	xmlURL := "http://www.crunchyroll.com/xml/?"

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
			"video_format":  {getVideoFormat(episode.Quality)},
			"video_quality": {getVideoQuality(episode.Quality)},
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
			"video_format":         {getVideoFormat(episode.Quality)},
			"video_encode_quality": {getVideoQuality(episode.Quality)},
		}
	}

	// Constructs a client and request that will get the xml we're asking for
	client := &http.Client{}
	xmlReq, err := http.NewRequest("POST", xmlURL+queryString.Encode(), bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", Error{"There was an error creating our getXML request", err}
	}
	xmlReq.Header.Add("Host", "www.crunchyroll.com")
	xmlReq.Header.Add("Origin", "http://static.ak.crunchyroll.com")
	xmlReq.Header.Add("Content-type", "application/x-www-form-urlencoded")
	xmlReq.Header.Add("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
	xmlReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	xmlReq.Header.Add("X-Requested-With", "ShockwaveFlash/19.0.0.245")
	for c := range cookies {
		xmlReq.AddCookie(cookies[c])
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

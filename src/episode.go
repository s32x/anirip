package main

import (
	"net/http"
	"net/url"
	"os/exec"
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
	err := episode.getEpisodeRTMPInfo(rtmpInfo, cookies)
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
func (episode *Episode) getEpisodeRTMPInfo(rtmpInfo *RTMPInfo, cookies []*http.Cookie) error {
	// First gets the XML of the episode video
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

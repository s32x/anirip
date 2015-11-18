package main

import (
	"fmt"
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
func downloadEpisode(episodeFileName string, episode Episode, params SessionParameters) error {
	// First attempts to get the XML attributes for the requested episode
	episodeRTMPInfo, err := getRMTPInfo(episode, params)
	if err != nil {
		return err
	}

	// Attempts to dump the FLV of the episode to file
	err = dumpEpisodeFLV(episodeRTMPInfo, episode, episodeFileName, params)
	if err != nil {
		return err
	}
	return nil
}

// Parses the xml and returns what we need from the xml
func getRMTPInfo(episode Episode, params SessionParameters) (RTMPInfo, error) {
	// First gets the XML of the episode video
	xmlString, err := getXML("RpcApiVideoPlayer_GetStandardConfig", episode, params)
	if err != nil {
		return RTMPInfo{}, err
	}

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return RTMPInfo{}, CRError{"This video is not available in your region", err}
		}
	}

	// Next performs some really basic parsing of the host url
	xmlHostURL := ""
	if strings.Contains(xmlString, "<host>") && strings.Contains(xmlString, "</host>") {
		xmlHostURL = strings.SplitN(strings.SplitN(xmlString, "<host>", 2)[1], "</host>", 2)[0]
	} else {
		return RTMPInfo{}, CRError{"No hosts were found for the episode", err}
	}

	// Same type of xml parsing to get the file
	episodeFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		episodeFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return RTMPInfo{}, CRError{"No hosts were found for the episode", err}
	}

	// Parses the URL in order to break out the two urls required for dumping
	url, err := url.Parse(xmlHostURL)
	if err != nil {
		return RTMPInfo{}, CRError{"There was an error parsing episode information", err}
	}

	// Finds the urls with our URL object and returns them
	return RTMPInfo{
		URLOne: url.Scheme + "://" + url.Host + url.Path,
		URLTwo: strings.Trim(url.RequestURI(), "/"),
		File:   episodeFile,
	}, nil
}

// Calls rtmpdump.exe to dump the episode and names it
func dumpEpisodeFLV(rtmp RTMPInfo, episode Episode, fileName string, params SessionParameters) error {
	// Gets the path of our rtmp dump exe
	path, err := exec.LookPath("engine\\rtmpdump.exe")
	if err != nil {
		return CRError{"Unable to find rtmpdump.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to dump the episode
	cmd := exec.Command(path,
		"-r", rtmp.URLOne,
		"-a", rtmp.URLTwo,
		"-f", "WIN 19,0,0,245",
		"-W", "http://static.ak.crunchyroll.com/versioned_assets/ChromelessPlayerApp.6282d5bd.swf",
		"-m", "10",
		"-p", episode.URL,
		"-y", rtmp.File,
		"-o", "temp\\"+fileName+".flv")

	// Append retry param if the file already exists
	_, err = exec.LookPath("temp\\" + fileName + ".flv")
	if err == nil {
		cmd.Args = append(cmd.Args, "-e")
	}

	// Executes the dump command and gets the episode
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error trying to execute our dumper", err}
	}
	fmt.Printf("Downloading " + fileName + ".flv to \\temp\\ directory...\n")
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("There was an error while downloading... Resuming...\n")
		// Recursively recalls downloadEpisode if we get an unfinished download
		downloadEpisode(fileName, episode, params)
	}
	fmt.Printf("Downloaded " + fileName + ".flv successfully!\n")
	return nil
}

func splitEpisodeFLV(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\flvextract.exe")
	if err != nil {
		return CRError{"Unable to find flvextract.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-v", "-a", "-t", "-o", "temp\\"+fileName+".flv")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error while executing our extracter", err}
	}
	fmt.Printf("Attempting to split " + fileName + ".flv...\n")
	err = cmd.Wait()
	if err != nil {
		return CRError{"There was an error while extracting", err}
	}
	fmt.Printf("Split " + fileName + ".flv successfully!\n")
	return nil
}

func mergeEpisodeMKV(fileName string) error {
	// TODO check if files exist before attempting final merge
	path, err := exec.LookPath("engine\\mkvmerge.exe")
	if err != nil {
		return CRError{"Unable to find mkvmerge.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path,
		"-o", "temp\\"+fileName+".mkv",
		"--language", "0:eng",
		"temp\\"+fileName+".ass",
		"temp\\"+fileName+".264",
		"--aac-is-sbr", "0",
		"temp\\"+fileName+".aac")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error while executing our merger", err}
	}
	fmt.Printf("Attempting to merge " + fileName + ".mkv...\n")
	err = cmd.Wait()
	if err != nil {
		return CRError{"There was an error while merging", err}
	}
	fmt.Printf("Merged " + fileName + ".mkv successfully!\n")
	return nil
}

package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type RTMPInfo struct {
	URLOne string
	URLTwo string
	File   string
}

func getEpisode(episode *Episode, params *SessionParameters) error {
	// Download episodes to file
	err := downloadEpisode(episode, params)
	if err != nil {
		return err
	}

	// Download subtitles to file
	err = downloadSubtitle(episode, params)
	if err != nil {
		return err
	}

	// Splits up the FLV file so we can handle all peices with mergemkv
	err = splitEpisodeFLV(episode)
	if err != nil {
		return err
	}

	// Merges all the files together to create a single solid MKV
	err = mergeEpisodeMKV(episode)
	if err != nil {
		return err
	}

	// Attempts to remove all the files we're done with
	os.Remove("temp\\" + episode.FileName + ".ass")
	os.Remove("temp\\" + episode.FileName + ".264")
	os.Remove("temp\\" + episode.FileName + ".txt")
	os.Remove("temp\\" + episode.FileName + ".aac")
	os.Remove("temp\\" + episode.FileName + ".flv")
	return nil
}

// Downloads entire FLV episodes to our temp directory
func downloadEpisode(episode *Episode, params *SessionParameters) error {
	// First attempts to get the XML attributes for the requested episode
	err := populateRTMPInfo(episode, params)
	if err != nil {
		return err
	}

	// Attempts to dump the FLV of the episode to file
	err = dumpEpisodeFLV(episode, params)
	if err != nil {
		return err
	}
	return nil
}

// Parses the xml and returns what we need from the xml
func populateRTMPInfo(episode *Episode, params *SessionParameters) error {
	// First gets the XML of the episode video
	xmlString, err := getXML("RpcApiVideoPlayer_GetStandardConfig", episode, params)
	if err != nil {
		return err
	}

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return CRError{"This video is not available in your region", err}
		}
	}

	// Next performs some really basic parsing of the host url
	xmlHostURL := ""
	if strings.Contains(xmlString, "<host>") && strings.Contains(xmlString, "</host>") {
		xmlHostURL = strings.SplitN(strings.SplitN(xmlString, "<host>", 2)[1], "</host>", 2)[0]
	} else {
		return CRError{"No hosts were found for the episode", err}
	}

	// Same type of xml parsing to get the file
	episodeFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		episodeFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return CRError{"No hosts were found for the episode", err}
	}

	// Parses the URL in order to break out the two urls required for dumping
	url, err := url.Parse(xmlHostURL)
	if err != nil {
		return CRError{"There was an error parsing episode information", err}
	}

	// Sets the RTMP info recieved before returning
	episode.RTMPInfo.URLOne = url.Scheme + "://" + url.Host + url.Path
	episode.RTMPInfo.URLTwo = strings.Trim(url.RequestURI(), "/")
	episode.RTMPInfo.File = episodeFile

	return nil
}

// Calls rtmpdump.exe to dump the episode and names it
func dumpEpisodeFLV(episode *Episode, params *SessionParameters) error {
	// Gets the path of our rtmp dump exe
	path, err := exec.LookPath("engine\\rtmpdump.exe")
	if err != nil {
		return CRError{"Unable to find rtmpdump.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to dump the episode
	cmd := exec.Command(path,
		"-r", episode.RTMPInfo.URLOne,
		"-a", episode.RTMPInfo.URLTwo,
		"-f", "WIN 19,0,0,245",
		"-W", "http://static.ak.crunchyroll.com/versioned_assets/ChromelessPlayerApp.6282d5bd.swf",
		"-m", "10",
		"-p", episode.URL,
		"-y", episode.RTMPInfo.File,
		"-o", "temp\\"+episode.FileName+".flv")

	// Append retry param if the file already exists
	_, err = exec.LookPath("temp\\" + episode.FileName + ".flv")
	if err == nil {
		cmd.Args = append(cmd.Args, "-e")
	}

	// Executes the dump command and gets the episode
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error trying to execute our dumper", err}
	}
	fmt.Printf("Downloading " + episode.FileName + ".flv to \\temp\\ directory...\n")
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("There was an error while downloading... Resuming...\n")
		// Recursively recalls downloadEpisode if we get an unfinished download
		downloadEpisode(episode, params)
	}
	fmt.Printf("Downloaded " + episode.FileName + ".flv successfully!\n")
	return nil
}

// Splits the episode into multiple media files that we will later merge together
func splitEpisodeFLV(episode *Episode) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\flvextract.exe")
	if err != nil {
		return CRError{"Unable to find flvextract.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-v", "-a", "-t", "-o", "temp\\"+episode.FileName+".flv")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error while executing our extracter", err}
	}
	fmt.Printf("Attempting to split " + episode.FileName + ".flv...\n")
	err = cmd.Wait()
	if err != nil {
		return CRError{"There was an error while extracting", err}
	}
	fmt.Printf("Split " + episode.FileName + ".flv successfully!\n")
	return nil
}

// Merges all media files including subs
func mergeEpisodeMKV(episode *Episode) error {
	// TODO check if files exist before attempting final merge
	path, err := exec.LookPath("engine\\mkvmerge.exe")
	if err != nil {
		return CRError{"Unable to find mkvmerge.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path,
		"-o", "temp\\"+episode.FileName+".mkv",
		"--language", "0:eng",
		"temp\\"+episode.FileName+".ass",
		"temp\\"+episode.FileName+".264",
		"--aac-is-sbr", "0",
		"temp\\"+episode.FileName+".aac")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return CRError{"There was an error while executing our merger", err}
	}
	fmt.Printf("Attempting to merge " + episode.FileName + ".mkv...\n")
	err = cmd.Wait()
	if err != nil {
		return CRError{"There was an error while merging", err}
	}
	fmt.Printf("Merged " + episode.FileName + ".mkv successfully!\n")
	return nil
}

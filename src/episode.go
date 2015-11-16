package crunchyrip

import (
	"fmt"
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

// Parses the xml and returns what we need from the xml
func getRMTPInfo(episode Episode, cookies []*http.Cookie) (RTMPInfo, error) {
	// First gets the XML of the episode video
	xmlString, err := getXML("RpcApiVideoPlayer_GetStandardConfig", episode, cookies)
	if err != nil {
		return RTMPInfo{}, err
	}

	// Checks for an unsupported region first
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			fmt.Println(">> This video is not available in your region...")
			return RTMPInfo{}, nil
		}
	}

	// Next performs some really basic parsing of the host url
	xmlHostURL := ""
	if strings.Contains(xmlString, "<host>") && strings.Contains(xmlString, "</host>") {
		xmlHostURL = strings.SplitN(strings.SplitN(xmlString, "<host>", 2)[1], "</host>", 2)[0]
	} else {
		fmt.Println(">> No hosts was found for the episode")
		return RTMPInfo{}, nil
	}

	// Same type of xml parsing to get the file
	episodeFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		episodeFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		fmt.Println(">> No hosts was found for the episode")
		return RTMPInfo{}, nil
	}

	// Parses the URL in order to break out the two urls required for dumping
	url, err := url.Parse(xmlHostURL)
	if err != nil {
		return RTMPInfo{}, err
	}

	// Finds the urls with our URL object and returns them
	return RTMPInfo{
		URLOne: url.Scheme + "://" + url.Host + url.Path,
		URLTwo: strings.Trim(url.RequestURI(), "/"),
		File:   episodeFile,
	}, nil
}

// Calls rtmpdump.exe to dump the episode and names it
func dumpEpisodeFLV(rtmp RTMPInfo, episodeURL string, fileName string) error {
	// Gets the path of our rtmp dump exe
	path, err := exec.LookPath("engine\\rtmpdump.exe")
	if err != nil {
		fmt.Printf(">> Unable to find rtmpdump.exe in /engine/ directory...\n")
		return err
	}

	// Creates the command which we will use to dump the episode
	cmd := exec.Command(path,
		"-r", rtmp.URLOne,
		"-a", rtmp.URLTwo,
		"-f", "WIN 11,8,800,50",
		"-m", "15",
		"-p", episodeURL,
		"-y", rtmp.File,
		"-o", "temp\\"+fileName+".flv")

	// Executes the dump command and gets the episode
	err = cmd.Start()
	if err != nil {
		fmt.Printf(">>> There was an error while executing rtmpdump.exe\n")
		return err
	}
	fmt.Printf("Downloading " + fileName + ".flv to /temp/ directory...\n")
	err = cmd.Wait()
	if err != nil {
		fmt.Println(">>> There was an error while downloading : ", err)
		return err
	}
	fmt.Printf("Downloaded " + fileName + ".flv successfully!\n")

	// TODO check the existance of the episode flv downloaded
	return nil
}

func splitEpisodeFLV(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\flvextract.exe")
	if err != nil {
		fmt.Printf(">> Unable to find flvextract.exe in /engine/ directory...\n")
		return err
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-v", "-a", "-t", "-o", "temp\\"+fileName+".flv")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		fmt.Printf(">>> There was an error while executing flvextract.exe")
		return err
	}
	fmt.Printf("Attempting to split " + fileName + ".flv...\n")
	err = cmd.Wait()
	if err != nil {
		fmt.Println(">>> There was an error while downloading : ", err)
		return err
	}
	fmt.Printf("Split " + fileName + ".flv successfully!\n")
	return nil
}

func mergeEpisodeMKV(fileName string) error {
	// TODO check if files exist before attempting final merge
	path, err := exec.LookPath("engine\\mkvmerge.exe")
	if err != nil {
		fmt.Printf(">> Unable to find mkvmerge.exe in /engine/ directory...\n")
		return err
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-o", "temp\\"+fileName+".mkv", "temp\\"+fileName+".ass", "temp\\"+fileName+".264", "--aac-is-sbr", "0", "temp\\"+fileName+".aac")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		fmt.Printf(">>> There was an error while executing mkvmerge.exe")
		return err
	}
	fmt.Printf("Attempting to merge " + fileName + ".mkv...\n")
	err = cmd.Wait()
	if err != nil {
		fmt.Println(">>> There was an error while merging : ", err)
		return err
	}
	fmt.Printf("Merged " + fileName + ".mkv successfully!\n")
	return nil
}

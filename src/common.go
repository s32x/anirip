package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

// Splits the episode into multiple media files that we will later merge together
func Split(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\flvextract.exe")
	if err != nil {
		return Error{"Unable to find flvextract.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-v", "-a", "-t", "-o", "temp\\"+fileName+".flv")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return Error{"There was an error while executing our extracter", err}
	}
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while extracting", err}
	}
	return nil
}

// Merges all media files including subs
func Merge(fileName string) error {
	// TODO check if files exist before attempting final merge
	path, err := exec.LookPath("engine\\mkvmerge.exe")
	if err != nil {
		return Error{"Unable to find mkvmerge.exe in \\engine\\ directory", err}
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
		return Error{"There was an error while executing our merger", err}
	}

	// Waits for the merge to complete
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while merging", err}
	}

	_, err = exec.LookPath("temp\\" + fileName + ".mkv")
	if err != nil {
		return Error{"Merged MKV was not found after merger", err}
	}
	// Erases all old media files that we no longer need
	os.Remove("temp\\" + fileName + ".ass")
	os.Remove("temp\\" + fileName + ".264")
	os.Remove("temp\\" + fileName + ".txt")
	os.Remove("temp\\" + fileName + ".aac")
	os.Remove("temp\\" + fileName + ".flv")
	return nil
}

// Cleans up the mkv, optimizing it for playback as well as old remaining files
func Clean(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\mkclean.exe")
	if err != nil {
		return Error{"Unable to find mkclean.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to clean our mkv to "video.clean.mkv"
	cmd := exec.Command(path, "--optimize", "temp\\"+fileName+".mkv")

	// Executes the cleaning and waits for a response
	err = cmd.Start()
	if err != nil {
		return Error{"There was an error while executing our mkv optimizer", err}
	}
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while optimizing our mkv", err}
	}

	// Deletes the old, un-needed dirty mkv file
	os.Remove("temp\\" + fileName + ".mkv")
	os.Rename("temp\\clean."+fileName+".mkv", "temp\\"+fileName+".mkv")
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

// Constructs an episode file name and returns the file name cleaned
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
	fileName := strings.Title(showTitle) + " - S" + seasonNumberString + "E" + episodeNumberString + " - " + description
	return cleanFileName(fileName)
}

// Cleans the new file/folder name so there won't be any write issues
func cleanFileName(fileName string) string {
	newFileName := fileName // Strips out any illegal characters and returns our new file name
	for _, illegalChar := range []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"} {
		newFileName = strings.Replace(newFileName, illegalChar, " ", -1)
	}
	return newFileName
}

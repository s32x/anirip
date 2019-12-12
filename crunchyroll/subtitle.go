package crunchyroll /* import "s32x.com/anirip/crunchyroll" */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"regexp"

	"s32x.com/anirip/common"
)

// Note from turtle: I haven't found a good use for the streams yet but I'll keep looking into it
type streamStruct struct {
	Format 		string `json:"format"`
	AudioLang 	string `json:"audio_lang"`
	SubLang 	string `json:"hardsub_lang"`
	urlThingy 	string `json:"url"`
	Resolution 	string `json:"resolution"`
}

type subtitleStruct struct {
	Language 	string `json:"language"`
	DownloadURL string `json:"url"`
	Title 		string `json:"title"`
	Format 		string `json:"format"`
}

type configStruct struct {
	Streams 	[]streamStruct `json:"streams"`
	Subtitles 	[]subtitleStruct `json:"subtitles"`
}

// DownloadSubtitles entirely downloads subtitles to our temp directory
func (episode *Episode) DownloadSubtitles(client *common.HTTPClient, language string, tempDir string) (string, error) {
	subOutput := tempDir + string(os.PathSeparator) + "subtitles.episode.ass"
	// Remove stale temp file to avoid conflicts in writing
	os.Remove(subOutput)

	// This turns "en-US" into "enUS", which is Crunchyroll's subtitle format
	isoCode := strings.Split(language, "-")[0]
	language = strings.ReplaceAll(language, "-", "")

	// Fetch html page for the episode
	res, err := client.Get(episode.URL, nil)
	if err != nil {
		return "", fmt.Errorf("getting episode page: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading episode response: %w", err)
	}

	jsonResult := regexp.MustCompile("vilos.config.media = (.*);\n").FindSubmatch(body)

	if jsonResult == nil {
		return "", fmt.Errorf("finding vilos config")
	}

	var subStruct configStruct
	if err := json.Unmarshal(jsonResult[1], &subStruct); err != nil {
		return "", fmt.Errorf("unmarshaling json: %w", err)
	}
	// Verify that there are actually subtitles
	if len(subStruct.Subtitles) == 0 {
		return "", fmt.Errorf("no subtitles found (?)")
	}

	// Determine the best subtitle (I know it currently defaults to enUS but it can be fixed later)
	chosenSubtitle := subStruct.Subtitles[0]
	for _, sub := range subStruct.Subtitles {
		if sub.Language == language {
			chosenSubtitle = sub
			break
		}
	}

	// Fetch the download page for the chosen subtitle (the page that's returned is the decrypted subtitles in ass format)
	subResp, err := client.Get(chosenSubtitle.DownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("getting download url: %w", err)
	}

	// Read the subtitles and output them to the subtitles.episode.ass file in the temp directory
	defer subResp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(subResp.Body)

	if err := ioutil.WriteFile(subOutput, buf.Bytes(), os.ModePerm); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return isoCode, nil
}

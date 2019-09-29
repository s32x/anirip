package crunchyroll /* import "s32x.com/anirip/crunchyroll" */

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

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
	// Remove stale temp file to avoid conflicts in func
	os.Remove(tempDir + string(os.PathSeparator) + "subtitles.episode.ass")

	// Subtitle language (The only two I know fo are enUS and jaJP)
	subLang := "enUS"

	// Fetch html page for the episode
	var body []byte
	if res, err := client.Get(episode.URL, nil); err == nil {
		defer res.Body.Close()
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return "", err
		}
	} else {
		return "", err
	}

	// Find the vilos config table and split the area after it. 
	// Then, trim away the "vilos.config.media = " to produce a json table in string form
	stringBody := string(body)
	mediaConfigIndex := strings.Index(stringBody, "vilos.config.media")
	newResult := strings.SplitAfter(stringBody[mediaConfigIndex:], "}]}")
	jsonReady := strings.TrimPrefix(newResult[0], "vilos.config.media = ")

	// Parse json string to configStruct
	var subStruct configStruct
	if err := json.Unmarshal([]byte(jsonReady), &subStruct); err != nil {
		return "", err
	}

	// Verify that there are actually subtitles
	if len(subStruct.Subtitles) == 0 {
		return "", errors.New("No subtitle files found(?)")
	}

	// Determine the best subtitle (I know it currently defaults to enUS but it can be fixed later)
	chosenSubtitle := subStruct.Subtitles[0]
	for _, sub := range subStruct.Subtitles {
		if sub.Language == subLang {
			chosenSubtitle = sub
			break
		}
	}

	// Fetch the download page for the chosen subtitle (the page that's returned is the decrypted subtitles in ass format)
	subResp, err := client.Get(chosenSubtitle.DownloadURL, nil)
	if err != nil {
		return "", err
	}

	// Read the subtitles and output them to the subtitles.episode.ass file in the temp directory
	defer subResp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(subResp.Body)

	if err := ioutil.WriteFile(tempDir + string(os.PathSeparator) + "subtitles.episode.ass", buf.Bytes(), 0777); err != nil {
		return "", err
	}
	return subLang, nil
}

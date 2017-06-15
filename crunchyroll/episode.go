package crunchyroll

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/anirip/anirip"
)

var (
	formats = map[string]string{
		"android": "107",
		"360":     "106",
		"480":     "106",
		"720":     "106",
		"1080":    "108",
		"default": "0",
	}
	qualities = map[string]string{
		"android": "71",
		"360":     "60",
		"480":     "61",
		"720":     "62",
		"1080":    "80",
		"default": "0",
	}
)

// Episode holds all episode metadata needed for downloading
type Episode struct {
	ID          int
	SubtitleID  int
	Title       string
	Description string
	Number      float64
	Quality     string
	Path        string
	URL         string
	Filename    string
	StreamURL   string
}

// GetEpisodeInfo retrieves and populates the metadata on the Episode
func (e *Episode) GetEpisodeInfo(client *anirip.HTTPClient, quality string) error {
	e.Quality = quality // Sets the quality to the passed quality string

	// Gets the HTML of the episode page
	// client.Header.Add("Referer", "http://www.crunchyroll.com/"+strings.Split(e.Path, "/")[1])
	resp, err := client.Get(e.URL, nil)
	if err != nil {
		return anirip.NewError("There was an error requesting the episode doc", err)
	}

	// Creates the document that will be used to scrape for episode metadata
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return anirip.NewError("There was an error reading the episode doc", err)
	}

	// Request querystring
	queryString := url.Values{
		"req":           {"RpcApiVideoPlayer_GetStandardConfig"},
		"media_id":      {strconv.Itoa(e.ID)},
		"video_format":  {getMapping(e.Quality, formats)},
		"video_quality": {getMapping(e.Quality, qualities)},
		"auto_play":     {"1"},
		"aff":           {"crunchyroll-website"},
		"show_pop_out_controls":   {"1"},
		"pop_out_disable_message": {""},
		"click_through":           {"0"},
	}.Encode()

	// Request body
	reqBody := bytes.NewBufferString(url.Values{"current_page": {e.URL}}.Encode())

	// Request header
	header := http.Header{}
	header.Add("Host", "www.crunchyroll.com")
	header.Add("Origin", "http://static.ak.crunchyroll.com")
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	header.Set("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.f3770232.swf")
	header.Add("X-Requested-With", "ShockwaveFlash/22.0.0.192")
	resp, err = client.Post("http://www.crunchyroll.com/xml/?"+queryString, header, reqBody)
	if err != nil {
		return anirip.NewError("There was an error retrieving the manifest", err)
	}

	// Gets the xml string from the recieved xml response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return anirip.NewError("There was an error reading the xml response", err)
	}

	// Checks for an unsupported region first
	// TODO Use REGEX to extract xml
	xmlString := string(body)
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return anirip.NewError("This video is not available in your region", err)
		}
	}

	// Same type of xml parsing to get the file
	// TODO Use REGEX to extract efile
	eFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		eFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return anirip.NewError("No hosts were found for the episode", err)
	}

	e.Title = strings.Replace(strings.Replace(doc.Find("#showmedia_about_name").First().Text(), "“", "", -1), "”", "", -1)
	e.Filename = anirip.CleanFilename(e.Filename + e.Title)
	e.StreamURL = strings.Replace(eFile, "amp;", "", -1)
	return nil
}

// Download downloads entire episode to our temp directory
func (e *Episode) Download(vp anirip.VideoProcessor) error {
	return vp.DumpHLS(e.StreamURL)
}

// GetFilename returns the Episodes filename
func (e *Episode) GetFilename() string {
	return e.Filename
}

// getMapping out what the format or resolution of the video should be based on
// crunchyroll xml
func getMapping(quality string, m map[string]string) string {
	a := strings.ToLower(quality)
	for k, v := range m {
		if strings.Contains(a, k) {
			return v
		}
	}
	return "0"
}

package daisuki

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sdwolfe32/ANIRip/anirip"
)

type TT struct {
	XMLLS string `xml:"xmlns,attr"`
	Head  Head   `xml:"head"`
	Body  Body   `xml:"body"`
}

type Head struct {
	Styling Styling `xml:"styling"`
}

type Styling struct {
	Styles []Style `xml:"style"`
}

type Style struct {
	ID          int    `xml:"id,attr"`
	TextOutline string `xml:"textOutline,attr"`
	Color       string `xml:"color,attr"`
}

type Body struct {
	Subtitles []Subtitle `xml:"div"`
}

type Subtitle struct {
	Language string  `xml:"lang,attr"`
	Events   []Event `xml:"p"`
}

type Event struct {
	Begin string `xml:"begin,attr"`
	End   string `xml:"end,attr"`
	Style int    `xml:"style,attr"`
	Text  string `xml:",chardata"`
}

// Entirely downloads subtitles to our temp directory
func (episode *DaisukiEpisode) DownloadSubtitles(language string, offset int, tempDir string, cookies []*http.Cookie) error {
	// Since we already have the subtitle info lets just go and download the subs
	// If we get back a subtitle that was nil (no ID), there are no subs available
	if episode.SubtitleInfo.TTMLUrl == "" {
		return nil
	}

	// Reaches out to the xml page and gets the subtitles
	subtitles := new(TT)
	if err := episode.getSubtitles(subtitles, cookies); err != nil {
		return err
	}

	// Dumps our final subtitle string into an ass file for merging later on
	if err := episode.dumpSubtitleASS(language, offset, subtitles, tempDir); err != nil {
		return err
	}
	return nil
}

// Gets the subtitles xml from daisuki, parses and popuulates XMTT param
func (episode *DaisukiEpisode) getSubtitles(subtitles *TT, cookies []*http.Cookie) error {
	// Gets the current time and sets up a referrer for our subtitle request
	nowMillis := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)

	// Performs the HTTP Request that will get the XML of the subtitles
	subReqHeaders := http.Header{}
	subReqHeaders.Add("referrer", episode.URL)
	subtitleResp, err := anirip.GetHTTPResponse("GET",
		episode.SubtitleInfo.TTMLUrl+"?cashPath="+nowMillis,
		nil,
		subReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Reads the bytes from the recieved subtitle response body
	subtitleXML, err := ioutil.ReadAll(subtitleResp.Body)
	if err != nil {
		return anirip.Error{Message: "There was an error reading the search response", Err: err}
	}

	// Parses the xml into our subtitles object
	if err = xml.Unmarshal(subtitleXML, subtitles); err != nil {
		return anirip.Error{Message: "There was an error while reading subtitle information", Err: nil}
	}
	return nil
}

// Writes formatted ASS subtitles to file
func (episode *DaisukiEpisode) dumpSubtitleASS(language string, offset int, subtitles *TT, tempDir string) error {
	// Attempts to format the subtitles for ASS
	formattedSubtitles, err := formatSubtitles(language, offset, subtitles)
	if err != nil || formattedSubtitles == "" {
		return err
	}

	// Writes the ASS subtitles to a file in our temp folder (with utf-8-sig encoding)
	subtitlesBytes := append([]byte{0xef, 0xbb, 0xbf}, []byte(formattedSubtitles)...)
	err = ioutil.WriteFile(tempDir+"\\"+episode.FileName+".ass", subtitlesBytes, 0644)
	if err != nil {
		return anirip.Error{Message: "There was an error while writing the subtitles to file", Err: err}
	}
	return nil
}

// Formats the subs while calculating subtitle offset shifts
func formatSubtitles(language string, offset int, subtitles *TT) (string, error) {
	// Sets up default ASS sub info
	header := "[Script Info]\nTitle: Default Aegisub file\nScriptType: v4.00+\nWrapStyle: 0\nPlayResX: 656\nPlayResY: 368\n\n"
	styles := "[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n"
	events := "\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

	// Appends styles TODO (ignoring color)
	for _, style := range subtitles.Head.Styling.Styles {
		styles = styles + "Style: " +
			strconv.Itoa(style.ID) + "," +
			"Trebuchet MS" + "," +
			"24" + "," +
			"&H00FFFFFF" + "," +
			"&H000000FF" + "," +
			"&H00000000" + "," +
			"&H00000000" + "," +
			"0" + "," +
			"0" + "," +
			"0" + "," +
			"0" + "," +
			"100" + "," +
			"100" + "," +
			"0" + "," +
			"0" + "," +
			"1" + "," +
			"2" + "," +
			"0" + "," +
			"2" + "," +
			"0040" + "," +
			"0040" + "," +
			"0018" + "," +
			"0" + "\n"
	}

	// Appends all subtitle captions where the language matches what we want
	for _, subs := range subtitles.Body.Subtitles {
		if strings.ToLower(subs.Language) == strings.ToLower(language) {
			for _, event := range subs.Events {
				beginTime, _ := shiftTime(event.Begin, offset)
				endTime, _ := shiftTime(event.End, offset)
				events = events + "Dialogue: 0," +
					beginTime + "," +
					endTime + "," +
					strconv.Itoa(event.Style) + "," +
					"" + "," + // Name of the person doing the talking
					"0000" + "," +
					"0000" + "," +
					"0000" + "," +
					"" + "," + // Event fx
					strings.Replace(strings.TrimSpace(strings.Replace(event.Text, "\t", "", -1)), "\n", `\N`, -1) + "\n" // Removes all tabs, then removes leading/trailing whitespace, then replaces /n with /N for ASS formatting
			}
		}
	}

	// Returns the full subtitles as an ASS string representation
	return header + styles + events, nil
}

// Shifts the subtitle time to account for the passed millisecond sub offset
func shiftTime(subTime string, offset int) (string, error) {
	// Sets the parsing format to accept a time like this
	assFormat := "15:04:05.999999"

	// Parses the passed subtitle time
	tm, err := time.Parse(assFormat, subTime)
	if err != nil {
		return "", anirip.Error{Message: "There was an error parsing subtitle time", Err: err}
	}

	// Parses the offset to a duration that will be subtracted from the parsed sub time
	offsetDuration, err := time.ParseDuration("-" + strconv.Itoa(offset) + "ms")
	if err != nil {
		return "", anirip.Error{Message: "There was an error parsing subtitle time", Err: err}
	}
	tm = tm.Add(offsetDuration)

	// returns the new shifted time
	return tm.Format("15:04:05.00"), nil
}

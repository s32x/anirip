package crunchyroll

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sdwolfe32/anirip/anirip"
)

type SubListResults struct {
	Subtitles []Subtitle `xml:"subtitle"`
}

type Subtitle struct {
	ID      int     `xml:"id,attr"`
	Link    string  `xml:"link,attr"`
	Title   string  `xml:"title,attr"`
	Default int     `xml:"default,attr"`
	Delay   float64 `xml:"delay,attr"`
	Iv      string  `xml:"iv"`
	Data    string  `xml:"data"`
}

type SubtitleScript struct {
	ID        int            `xml:"id,attr"`
	Title     string         `xml:"title,attr"`
	PlayResX  int            `xml:"play_res_x,attr"`
	PlayResY  int            `xml:"play_res_y,attr"`
	LangCode  string         `xml:"lang_code,attr"`
	Lang      string         `xml:"lang_string,attr"`
	Created   string         `xml:"created,attr"`
	Progress  string         `xml:"progress_string,attr"`
	Status    string         `xml:"status_string,attr"`
	WrapStyle int            `xml:"wrap_style,attr"`
	Styles    []ScriptStyles `xml:"styles"`
	Events    []ScriptEvents `xml:"events"`
}

type ScriptStyles struct {
	Styles []Style `xml:"style"`
}

type Style struct {
	ID             int    `xml:"id,attr"`
	Name           string `xml:"name,attr"`
	FontName       string `xml:"font_name,attr"`
	FontSize       int    `xml:"font_size,attr"`
	PrimaryColor   string `xml:"primary_colour,attr"`
	SecondaryColor string `xml:"secondary_colour,attr"`
	OutlineColor   string `xml:"outline_colour,attr"`
	BackColor      string `xml:"back_colour,attr"`
	Bold           int    `xml:"bold,attr"`
	Italic         int    `xml:"italic,attr"`
	Underline      int    `xml:"underline,attr"`
	Strikeout      int    `xml:"strikeout,attr"`
	ScaleX         int    `xml:"scale_x,attr"`
	ScaleY         int    `xml:"scale_y,attr"`
	Spacing        int    `xml:"spacing,attr"`
	Angle          int    `xml:"angle,attr"`
	BorderStyle    int    `xml:"border_style,attr"`
	Outline        int    `xml:"outline,attr"`
	Shadow         int    `xml:"shadow,attr"`
	Alignment      int    `xml:"alignment,attr"`
	MarginLeft     string `xml:"margin_l,attr"`
	MarginRight    string `xml:"margin_r,attr"`
	MarginVert     string `xml:"margin_v,attr"`
	Encoding       int    `xml:"encoding,attr"`
}

type ScriptEvents struct {
	Events []Event `xml:"event"`
}

type Event struct {
	Event       []ScriptEvents `xml:"events"`
	ID          int            `xml:"id,attr"`
	Start       string         `xml:"start,attr"`
	End         string         `xml:"end,attr"`
	Style       string         `xml:"style,attr"`
	Name        string         `xml:"name,attr"`
	MarginLeft  string         `xml:"margin_l,attr"`
	MarginRight string         `xml:"margin_r,attr"`
	MarginVert  string         `xml:"margin_v,attr"`
	Effect      string         `xml:"effect,attr"`
	Text        string         `xml:"text,attr"`
}

// Entirely downloads subtitles to our temp directory
// IGNORING offset for now (no reason to trim cr subs)
func (episode *Episode) DownloadSubtitles(client *anirip.HTTPClient, language string, tempDir string) (string, error) {
	// Remove stale temp file to avoid conflicts in func
	os.Remove(tempDir + string(os.PathSeparator) + "subtitles.episode.ass")

	// Populates the subtitle info for the episode
	subtitles := new(Subtitle)
	subtitleLang, err := episode.getSubtitleInfo(client, subtitles, language)
	if err != nil {
		return "", anirip.NewError("Failed to retrieve subtitle manifest", err)
	}

	// If we get back a subtitle that was nil (no ID), there are no subs available
	if episode.SubtitleID == 0 {
		return "", nil
	}

	// Places the new subtitle object with JUST INFO into the episode and gets the sub data
	if err = episode.getSubtitleData(client, subtitles); err != nil {
		return "", err
	}

	// Dumps our final subtitle string into an ass file for merging later on
	if err = episode.dumpSubtitleASS(subtitles, tempDir); err != nil {
		return "", err
	}

	// Defaulting to english for now...
	return subtitleLang, nil
}

func (episode *Episode) getSubtitleInfo(client *anirip.HTTPClient, subtitles *Subtitle, language string) (string, error) {
	// Querystring to ask for the subtitles info
	queryString := url.Values{
		"req":                  {"RpcApiSubtitle_GetListing"},
		"media_id":             {strconv.Itoa(episode.ID)},
		"video_format":         {getMapping(episode.Quality, formats)},
		"video_encode_quality": {getMapping(episode.Quality, qualities)},
	}.Encode()

	// Formdata to indicate the source page
	reqBody := bytes.NewBufferString(url.Values{
		"current_page": {episode.URL},
	}.Encode())

	// Performs the HTTP Request that will get the XML
	client.Header.Add("host", "www.crunchyroll.com")
	client.Header.Add("origin", "http://static.ak.crunchyroll.com")
	client.Header.Add("content-type", "application/x-www-form-urlencoded")
	client.Header.Add("referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
	client.Header.Add("x-requested-with", "ShockwaveFlash/19.0.0.245")
	resp, err := client.Post("http://www.crunchyroll.com/xml/?"+queryString, reqBody)
	if err != nil {
		return "", err
	}

	var subListResults SubListResults
	if err := xml.NewDecoder(resp.Body).Decode(&subListResults); err != nil {
		return "", err
	}

	// Finds the subtitle ID of the language we want
	for i := 0; i < len(subListResults.Subtitles); i++ {
		if strings.Contains(subListResults.Subtitles[i].Title, language) {
			subtitles = &subListResults.Subtitles[i]
			episode.SubtitleID = subtitles.ID
			return "eng", nil
		}
	}

	// If we cant find the requested language default to English
	for i := 0; i < len(subListResults.Subtitles); i++ {
		if strings.Contains(subListResults.Subtitles[i].Title, "English") {
			subtitles = &subListResults.Subtitles[i]
			episode.SubtitleID = subtitles.ID
			return "eng", nil
		}
	}

	// Again, if there are no subs found after a succesfull parse, they are either hardcoded or dubbed
	return "", nil
}

// Assigns the subtitle to the passed episode and attempts to get the xml subs for this episode
func (episode *Episode) getSubtitleData(client *anirip.HTTPClient, subtitles *Subtitle) error {
	// Querystring to ask for the subtitles data
	queryString := url.Values{
		"req":                {"RpcApiSubtitle_GetXml"},
		"subtitle_script_id": {strconv.Itoa(episode.SubtitleID)},
	}.Encode()

	// Formdata to indicate the source page
	body := bytes.NewBufferString(url.Values{
		"current_page": {episode.URL},
	}.Encode())

	// Performs the HTTP Request that will get the XML
	client.Header.Add("host", "www.crunchyroll.com")
	client.Header.Add("origin", "http://static.ak.crunchyroll.com")
	client.Header.Add("content-type", "application/x-www-form-urlencoded")
	client.Header.Add("referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
	client.Header.Add("x-requested-with", "ShockwaveFlash/19.0.0.245")
	resp, err := client.Post("http://www.crunchyroll.com/xml/?"+queryString, body)
	if err != nil {
		return err
	}
	return xml.NewDecoder(resp.Body).Decode(&subtitles)
}

// Dumps the crunchyroll subtitles to file to be muxed into MKV
func (episode *Episode) dumpSubtitleASS(subtitles *Subtitle, tempDir string) error {
	// Attempts to decrypt the compressed subtitles we recieved
	decryptedSubtitles, err := decryptSubtitles(subtitles)
	if err != nil {
		return err
	}

	// Attempts to format the subtitles for ASS
	formattedSubtitles, err := formatSubtitles(decryptedSubtitles)
	if err != nil {
		return err
	}

	// Writes the ASS subtitles to a file in our temp folder (with utf-8-sig encoding)
	subtitlesBytes := append([]byte{0xef, 0xbb, 0xbf}, []byte(formattedSubtitles)...)
	err = ioutil.WriteFile(tempDir+string(os.PathSeparator)+"subtitles.episode.ass", subtitlesBytes, 0777)
	if err != nil {
		return err
	}
	return nil
}

// Decrypts the titles
func decryptSubtitles(subtitle *Subtitle) (string, error) {
	// Generates the key that will be used to decrypt our subtitles
	key := generateKey(subtitle.ID)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Gets the bytes of both our iv and subtitle data
	iv, err := base64.StdEncoding.DecodeString(subtitle.Iv)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(subtitle.Data)
	if err != nil {
		return "", err
	}

	// Decrypts our subtitles back into our data byte array
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// Decompresses the subtitles which we've decrypted
	reader := bytes.NewReader(data)
	var subOutput bytes.Buffer
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return "", err
	}
	io.Copy(&subOutput, zlibReader)
	zlibReader.Close()

	// Returns the string output of the reader (subtitles string)
	return subOutput.String(), nil
}

func formatSubtitles(subString string) (string, error) {
	subScript := SubtitleScript{}

	// Parses the xml into our results object
	err := xml.Unmarshal([]byte(subString), &subScript)
	if err != nil {
		return "", err
	}

	// Discarding language for now in order to set to default playback subtitle (subScript.Title)
	header := "[Script Info]\nTitle: Default Aegisub file\nScriptType: v4.00+\nWrapStyle: " + strconv.Itoa(subScript.WrapStyle) + "\nPlayResX: 656\nPlayResY: 368\n\n"
	styles := "[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n"
	events := "\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

	styleArray := subScript.Styles[0].Styles
	eventArray := subScript.Events[0].Events

	for _, style := range styleArray {
		styles = styles + "Style: " +
			style.Name + "," +
			style.FontName + "," +
			strconv.Itoa(style.FontSize) + "," +
			style.PrimaryColor + "," +
			style.SecondaryColor + "," +
			style.OutlineColor + "," +
			style.BackColor + "," +
			strconv.Itoa(style.Bold) + "," +
			strconv.Itoa(style.Italic) + "," +
			strconv.Itoa(style.Underline) + "," +
			strconv.Itoa(style.Strikeout) + "," +
			strconv.Itoa(style.ScaleX) + "," +
			strconv.Itoa(style.ScaleY) + "," +
			strconv.Itoa(style.Spacing) + "," +
			strconv.Itoa(style.Angle) + "," +
			strconv.Itoa(style.BorderStyle) + "," +
			strconv.Itoa(style.Outline) + "," +
			strconv.Itoa(style.Shadow) + "," +
			strconv.Itoa(style.Alignment) + "," +
			style.MarginLeft + "," +
			style.MarginRight + "," +
			style.MarginVert + "," +
			strconv.Itoa(style.Encoding) + "\n"
	}

	for _, event := range eventArray {
		events = events + "Dialogue: 0," +
			event.Start + "," +
			event.End + "," +
			event.Style + "," +
			event.Name + "," +
			event.MarginLeft + "," +
			event.MarginRight + "," +
			event.MarginVert + "," +
			event.Effect + "," +
			event.Text + "\n"
	}

	return header + styles + events, nil
}

func generateKey(subtitleID int) []byte {
	// Does some dank maths to calculate the location of waldo
	eq1 := int(math.Floor((math.Sqrt(6.9) * math.Pow(2, 25)))) ^ subtitleID
	eq2 := int(math.Floor(math.Sqrt(6.9) * math.Pow(2, 25)))
	eq3 := uint32((subtitleID ^ eq2) ^ (subtitleID^eq2)>>3 ^ eq1*32)

	// Creates a 160-Bit SHA1 hash
	hashData := []byte(createString([]int{20, 97, 1, 2}) + fmt.Sprint(eq3))
	shortHashArray := sha1.Sum(hashData)

	// Transforms shortHashArray into 256bit in case a 256bit key is requested
	longHashArray := [32]byte{}
	for i := range shortHashArray {
		longHashArray[i] = shortHashArray[i]
	}

	// Finally turns our longhash into a standard byte array for conversion to string
	finalHashArray := []byte{}
	for i := range longHashArray {
		finalHashArray = append(finalHashArray, longHashArray[i])
	}
	return finalHashArray
}

func createString(args []int) string {
	i := 0
	argArray := []int{args[2], args[3]}
	for i < args[0] {
		argArray = append(argArray, argArray[len(argArray)-1]+argArray[len(argArray)-2])
		i = i + 1
	}
	finalString := ""
	for _, arg := range argArray[2:] {
		finalString += string(arg%args[1] + 33)
	}
	return finalString
}

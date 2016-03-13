package daisuki

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/ANIRip/anirip"
)

type ApiData struct {
	SS_ID     string `json:"ss_id,omitempty"`
	MV_ID     string `json:"mv_id,omitempty"`
	Device_CD string `json:"device_cd,omitempty"`
	SS1_PRM   string `json:"ss1_prm,omitempty"`
	SS2_PRM   string `json:"ss2_prm,omitempty"`
	SS3_PRM   string `json:"ss3_prm,omitempty"`
}

type InitResponse struct {
	Rcd string `json:"rcd"`
	Rtn string `json:"rtn"`
}

type MetaData struct {
	Ssrcd         string   `json:"ssrcd"`
	InitID        string   `json:"init_id"`
	PlayURL       string   `json:"play_url"`
	SttmvURL      string   `json:"sttmv_url"`
	VlEnableF     string   `json:"vl_enable_f"`
	VlURL         string   `json:"vl_url"`
	VlIntervalSec string   `json:"vl_interval_sec"`
	VlTimeoutSec  string   `json:"vl_timeout_sec"`
	VlErrlimitCnt string   `json:"vl_errlimit_cnt"`
	BwLabel       []string `json:"bw_label"`
	AdqueMsec     []string `json:"adque_msec"`
	CaptionURL    string   `json:"caption_url"`
	CaptionLang   []struct {
		English string `json:"English"`
	} `json:"caption_lang"`
	CopyrightStr string      `json:"copyright_str"`
	PreimgURL    interface{} `json:"preimg_url"`
	AutoplayFlag string      `json:"autoplay_flag"`
	SeriesStr    string      `json:"series_str"`
	TitleStr     string      `json:"title_str"`
	SsExt        struct {
		Series  string `json:"series"`
		Product string `json:"product"`
		Fov     string `json:"fov"`
		Vast    string `json:"vast"`
		Prev    struct {
			MovieProductID string `json:"movie_product_id"`
		} `json:"prev"`
		Next struct {
			MovieProductID string `json:"movie_product_id"`
		} `json:"next"`
	} `json:"ss_ext"`
}

// Parses the xml and returns what we need from the xml
func (episode *DaisukiEpisode) GetEpisodeInfo(quality string, cookies []*http.Cookie) error {
	episode.Quality = quality // Sets the quality to the passed quality string

	// Gets the HTML of the episode page
	episodeReqHeaders := http.Header{}
	episodeReqHeaders.Add("referer", "http://www.daisuki.net/us/en/anime/detail."+strings.Split(episode.Path, ".")[1]+".html")
	episodeResponse, err := anirip.GetHTTPResponse("GET",
		episode.URL,
		nil,
		episodeReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Creates the goquery document that will be used to scrape for episode info
	episodeDoc, err := goquery.NewDocumentFromResponse(episodeResponse)
	if err != nil {
		return anirip.Error{Message: "There was an error while reading the episode doc", Err: err}
	}

	flashVars := map[string]string{}
	episodeDoc.Find("div#movieFlash").Each(func(i int, flash *goquery.Selection) {
		// Finds the movie/large episode and adds it to our map
		flashVarString := flash.Find("script").First().Text()
		flashVarString = strings.SplitN(flashVarString, "}", 2)[0]
		flashVarString = strings.SplitN(flashVarString, "{", 2)[1]
		flashVarString = strings.Replace(flashVarString, "'", "\"", -1)
		flashVarString = strings.Replace(flashVarString, " ", "", -1)
		flashVarString = strings.Replace(flashVarString, "\n", "", -1)
		flashVarsArray := strings.Split(flashVarString, "\"")
		newFlashVarsArray := []string{}
		for f := 1; f < len(flashVarsArray)-1; f++ {
			if !strings.Contains(flashVarsArray[f], ":") && !strings.Contains(flashVarsArray[f], ",") {
				newFlashVarsArray = append(newFlashVarsArray, flashVarsArray[f])
			}
		}
		// Declares and fills our map with all the key,values needed from flashvars
		var e = 0
		for e < len(newFlashVarsArray) {
			flashVars[newFlashVarsArray[e]] = newFlashVarsArray[e+1]
			e = e + 2
		}
	})

	// Check for required fields in flashVars map
	if _, ok := flashVars["s"]; !ok {
		return anirip.Error{Message: "'s' was missing from flashvars", Err: err}
	}
	if _, ok := flashVars["country"]; !ok {
		return anirip.Error{Message: "'country' was missing from flashvars", Err: err}
	}
	if _, ok := flashVars["init"]; !ok {
		return anirip.Error{Message: "'init' was missing from flashvars", Err: err}
	}

	// Gets the current time which we will use in our flashvars request
	nowMillis := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)

	// Performs the HTTP Request that will get the country code
	countryReqHeaders := http.Header{}
	countryReqHeaders.Add("referer", "https://www.crunchyroll.com/login")
	countryReqHeaders.Add("content-type", "application/x-www-form-urlencoded")
	countryResponse, err := anirip.GetHTTPResponse("GET",
		"http://www.daisuki.net"+flashVars["country"]+"?cashPath="+nowMillis,
		nil,
		countryReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Reads the body and extracts the country code which will be used in later requests
	body, err := ioutil.ReadAll(countryResponse.Body)
	if err != nil {
		return anirip.Error{Message: "There was an error getting the country code", Err: err}
	}
	countryCode := strings.Split(strings.Split(string(body), "<country_code>")[1], "</country_code>")[0]

	// Generates a new random 256bit key
	key := make([]byte, 32)
	if _, err = io.ReadFull(crand.Reader, key); err != nil {
		return anirip.Error{Message: "There was an error generating 256 bit key", Err: err}
	}

	api := new(ApiData)
	if val, ok := flashVars["ss_id"]; ok {
		api.SS_ID = val
	}
	if val, ok := flashVars["mv_id"]; ok {
		api.MV_ID = val
	}
	if val, ok := flashVars["device_cd"]; ok {
		api.Device_CD = val
	}
	if val, ok := flashVars["ss1_prm"]; ok {
		api.SS1_PRM = val
	}
	if val, ok := flashVars["ss2_prm"]; ok {
		api.SS2_PRM = val
	}
	if val, ok := flashVars["ss3_prm"]; ok {
		api.SS3_PRM = val
	}
	plainApiJSON, err := json.Marshal(api)
	if err != nil {
		return anirip.Error{Message: "There was an error marshalling api json", Err: err}
	}

	// Pads plainApiJSON to equal full BlockSize if not full block multiple
	if len(plainApiJSON)%aes.BlockSize != 0 {
		padding := aes.BlockSize - (len(plainApiJSON) % aes.BlockSize)
		paddedJSON := make([]byte, len(plainApiJSON)+padding)
		for p, b := range plainApiJSON {
			paddedJSON[p] = b
		}
		plainApiJSON = paddedJSON
	}

	// Creates a new aes cipher using the key we generated
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return anirip.Error{Message: "There was an error generating aes cipherblock", Err: err}
	}

	cipherApiJSON := make([]byte, len(plainApiJSON))
	iv := make([]byte, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(cipherBlock, iv)
	mode.CryptBlocks(cipherApiJSON, plainApiJSON)

	// Key uused to re-encrypt our request data to daisuki
	var pemPublicKey = "-----BEGIN PUBLIC KEY-----\n" +
		"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDFUkwl6OFLNms3VJQL7rb5bLfi\n" +
		"/u8Lkyx2WaDFw78XPWAkZMLfc9aTtROuBv8b6PNnUpqzC/lpxWQFIhgfKgxR6lRq\n" +
		"4SDT2NkIWV5O/3ZbOJzeCAoe9/G7+wdBHMVo23O39SHO3ycMv74N28KbGsnQ8tj0\n" +
		"NZCYyv/ubQeRUCAHfQIDAQAB\n" +
		"-----END PUBLIC KEY-----"

	pemBlock, _ := pem.Decode([]byte(pemPublicKey))
	pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return anirip.Error{Message: "There was an error x509 parsing our pem public key", Err: err}
	}
	pubKey := pub.(*rsa.PublicKey)
	encodedKey, err := rsa.EncryptPKCS1v15(crand.Reader, pubKey, key)
	if err != nil {
		return anirip.Error{Message: "There was an error encrypting our generated key", Err: err}
	}

	// Constructs url params used in encrypted init request
	queryParams := url.Values{
		"s": {flashVars["s"]},
		"c": {countryCode},
		"e": {strings.Replace(episode.URL, ".", "%2", -1)},
		"d": {base64.StdEncoding.EncodeToString(cipherApiJSON)},
		"a": {base64.StdEncoding.EncodeToString(encodedKey)},
	}

	// Executes init request
	bgnInitReqHeaders := http.Header{}
	bgnInitReqHeaders.Add("Content-Type", "application/x-www-form-urlencoded")
	bgnInitReqHeaders.Add("X-Requested-With", "ShockwaveFlash/20.0.0.306")
	bgnInitResponse, err := anirip.GetHTTPResponse("GET",
		"http://www.daisuki.net"+flashVars["init"]+"?"+queryParams.Encode(),
		nil,
		bgnInitReqHeaders,
		cookies)
	if err != nil {
		return err
	}

	// Reads the body of our init requests response
	body, err = ioutil.ReadAll(bgnInitResponse.Body)
	if err != nil {
		return anirip.Error{Message: "There was an error reading init response body", Err: err}
	}

	// Parses our json init response body
	initBody := new(InitResponse)
	if err = json.Unmarshal(body, initBody); err != nil {
		return anirip.Error{Message: "There was an error unmarshalling init response body", Err: err}
	}

	// Attempts to decrypt the encrypted data recieved from InitResponse
	inData, err := base64.StdEncoding.DecodeString(initBody.Rtn)
	if err != nil {
		return anirip.Error{Message: "Unable to base64 decode init return", Err: err}
	}

	// Pads inData to equal full BlockSize if not full block multiple
	if len(inData)%aes.BlockSize != 0 {
		padding := aes.BlockSize - (len(inData) % aes.BlockSize)
		paddedJSON := make([]byte, len(inData)+padding)
		for p, b := range inData {
			paddedJSON[p] = b
		}
		inData = paddedJSON
	}
	outData := make([]byte, len(inData))
	mode = cipher.NewCBCDecrypter(cipherBlock, iv)
	mode.CryptBlocks(outData, inData)

	// Finds the last non-zero byte of outData
	end := len(outData)
	for outData[end-1] == 0 {
		end--
	}

	// If the end of the array isn't the length of outData resize
	if end != len(outData) {
		outData = outData[:end]
	}

	// Finally unmarshalls metadata json after decryption
	metaData := new(MetaData)
	if err = json.Unmarshal(outData, metaData); err != nil {
		return anirip.Error{Message: "There was an error unmarshalling daisuki metadata", Err: err}
	}

	// Stores all the info we needed for getting the episodes info
	episode.Title = strings.SplitN(metaData.TitleStr, " ", 2)[1]
	episode.FileName = anirip.CleanFileName(episode.FileName + episode.Title) // Updates filename with title that we just scraped
	episode.SubtitleInfo = TTMLInfo{
		TTMLUrl: metaData.CaptionURL,
	}
	episode.MediaInfo = HDSInfo{
		ManifestURL: metaData.PlayURL,
	}
	return nil
}

// Downloads entire FLV episodes to our temp directory
func (episode *DaisukiEpisode) DownloadEpisode(quality, engineDir string, tempDir string, cookies []*http.Cookie) error {
	// Attempts to dump the FLV of the episode to file
	err := episode.dumpEpisodeFLV(quality, engineDir, tempDir)
	if err != nil {
		return err
	}

	// Finally renames the dumped FLV to an MKV
	if err := anirip.Rename(tempDir+"\\incomplete.episode.flv", tempDir+"\\episode.mkv", 10); err != nil {
		return err
	}
	return nil
}

// Gets the filename of the episode for referencing outside of this lib
func (episode *DaisukiEpisode) GetFileName() string {
	return episode.FileName
}

// Calls on AdobeHDS.php to dump the episode and name it
func (episode *DaisukiEpisode) dumpEpisodeFLV(quality string, engineDir, tempDir string) error {
	// Remove stale temp file to avoid conflcts with CLI
	os.Remove(tempDir + "\\incomplete.episode.flv")

	episode.Quality = quality // Sets the quality to the passed quality string

	// Gets the path of php and our adobeHDS php fil
	phpPath, err := filepath.Abs(engineDir + "\\php\\php.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find php.exe in \\" + engineDir + "\\php\\ directory", Err: err}
	}
	adobeHDSPath, err := filepath.Abs(engineDir + "\\AdobeHDS.php")
	if err != nil {
		return anirip.Error{Message: "Unable to find adobeHDS in \\" + engineDir + "\\ directory", Err: err}
	}

	// Executes the dump command and gets the episode
	cmd := exec.Command(phpPath, adobeHDSPath,
		"--manifest", episode.MediaInfo.ManifestURL+"&g="+generateGUID(12)+"&hdcore=3.2.0",
		"--outfile", "incomplete.episode",
		"--quality", "high",
		"--referrer", episode.URL,
		"--rename", "--delete")
	cmd.Dir = tempDir // Sets working directory to temp so our fragments end up there
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Executes the command
	if err = cmd.Run(); err != nil {
		// Recursively recalls dempEpisodeFLV if we get an unfinished download
		episode.dumpEpisodeFLV(quality, engineDir, tempDir)
	}

	return nil
}

// Generates a random GUID of n length for use in our Manifest URL
func generateGUID(n int) string {
	letterBytes := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	return string(b)
}

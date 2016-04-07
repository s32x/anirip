//go:generate goversioninfo -icon=icon.ico

package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/sdwolfe32/ANIRip/anirip"
	"github.com/sdwolfe32/ANIRip/crunchyroll"
	"github.com/sdwolfe32/ANIRip/daisuki"
)

var (
	tempDir = os.TempDir() + string(os.PathSeparator) + "ANIRip"
)

const (
	daisukiIntroLength = 5040
	aniplexIntroLength = 6747
	sunriseIntroLength = 8227
)

func main() {
	username := ""
	password := ""
	language := "English"
	quality := "1080p"
	trim := ""
	daisukiIntroTrim := false
	aniplexIntroTrim := false
	sunriseIntroTrim := false

	app := cli.NewApp()
	app.Name = "ANIRip"
	app.Author = "Viz_"
	app.Email = "vizicore@gmail.com"
	app.Version = "v1.3.1(4/4/2016)"
	app.Usage = "Crunchyroll/Daisuki show ripper CLI"
	color.Cyan(app.Name + " " + app.Version + " - by " + app.Author + " <" + app.Email + ">\n")
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "lang, l",
			Value:       "english",
			Usage:       "desired subtitle language",
			Destination: &language,
		},
		cli.StringFlag{
			Name:        "quality, q",
			Value:       "1080p",
			Usage:       "desired video quality",
			Destination: &quality,
		},
		cli.StringFlag{
			Name:        "trim, t",
			Value:       "daisuki,aniplex,sunrise",
			Usage:       "desired intros to be trimmed off of final video",
			Destination: &quality,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "login",
			Aliases: []string{"l"},
			Usage:   "creates and stores cookies for a stream provider",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "user, u",
					Value:       "myusername",
					Usage:       "premium username used to access video stream",
					Destination: &username,
				},
				cli.StringFlag{
					Name:        "pass, p",
					Value:       "mypassword",
					Usage:       "premium password used to access video stream",
					Destination: &password,
				},
			},
			Action: func(c *cli.Context) {
				// Gets the provider name from the cli argument
				provider := ""
				if c.NArg() > 0 {
					provider = c.Args()[0]
				} else {
					color.Red("[ANIRip] No provider given...")
					return
				}

				// Creates session with cookies to store in file
				var session anirip.Session
				if strings.Contains(provider, "crunchyroll") {
					color.Cyan("[ANIRip] Logging to CrunchyRoll as " + username + "...")
					session = new(crunchyroll.CrunchyrollSession)
				} else if strings.Contains(provider, "daisuki") {
					color.Cyan("[ANIRip] Logging to Daisuki as " + username + "...")
					session = new(daisuki.DaisukiSession)
				} else {
					color.Red("[ANIRip] The given provider is not supported.")
					return
				}

				// Performs the login procedure, storing the login information to file
				if err := session.Login(username, password, tempDir); err != nil {
					color.Red("[ANIRip] " + err.Error())
					return
				}
				color.Green("[ANIRip] Successfully logged in... Cookies saved to " + tempDir)
			},
		},
		{
			Name:    "clear",
			Aliases: []string{"c"},
			Usage:   "erases the temporary directory used for cookies and temp files",
			Action: func(c *cli.Context) {
				// Attempts to erase the temporary directory
				if err := os.RemoveAll(tempDir); err != nil {
					color.Red("[ANIRip] There was an error erasing the temporary directory : " + err.Error())
					return
				}
				color.Green("[ANIRip] Successfully erased the temporary directory " + tempDir)
			},
		},
	}
	app.Action = func(c *cli.Context) {
		showURL := ""
		if c.NArg() > 0 {
			showURL = c.Args()[0]
		} else {
			color.Red("[ANIRip] No show URL provided.")
			return
		}

		// Parses the URL so we can accurately judge the provider based on the host
		url, err := url.Parse(showURL)
		if err != nil {
			color.Red("[ANIRip] There was an error parsing the URL you entered.\n")
			return
		}

		// Creates the authentication & show objects for the provider we're ripping from
		var session anirip.Session
		var show anirip.Show
		if strings.Contains(strings.ToLower(url.Host), "crunchyroll") {
			show = new(crunchyroll.CrunchyrollShow)
			session = new(crunchyroll.CrunchyrollSession)
		} else if strings.Contains(strings.ToLower(url.Host), "daisuki") {
			show = new(daisuki.DaisukiShow)
			session = new(daisuki.DaisukiSession)
		} else {
			color.Red("[ANIRip] The URL provided is not supported.")
			return
		}

		// Performs the generic login procedure
		if err = session.Login(username, password, tempDir); err != nil {
			color.Red("[ANIRip] " + err.Error())
			return
		}

		// Attempts to scrape the shows metadata/information
		color.White("[ANIRip] Getting a list of episodes for the show...")
		if err = show.ScrapeEpisodes(showURL, session.GetCookies()); err != nil {
			color.Red("[ANIRip] " + err.Error())
			return
		}

		// Sets the boolean values for what intros we would like to trim
		if strings.Contains(strings.ToLower(trim), "daisuki") {
			daisukiIntroTrim = true
		}
		if strings.Contains(strings.ToLower(trim), "aniplex") {
			daisukiIntroTrim = true
		}
		if strings.Contains(strings.ToLower(trim), "sunrise") {
			daisukiIntroTrim = true
		}

		seasonMap := map[int]string{
			0:  "Specials",
			1:  "Season One",
			2:  "Season Two",
			3:  "Season Three",
			4:  "Season Four",
			5:  "Season Five",
			6:  "Season Six",
			7:  "Season Seven",
			8:  "Season Eight",
			9:  "Season Nine",
			10: "Season Ten",
		}

		os.Mkdir(show.GetTitle(), 0777)
		for _, season := range show.GetSeasons() {
			os.Mkdir(show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()], 0777)
			for _, episode := range season.GetEpisodes() {
				color.White("[ANIRip] Getting Episode Info...\n")
				if err = episode.GetEpisodeInfo(quality, session.GetCookies()); err != nil {
					color.Red("[ANIRip] " + err.Error())
					continue
				}

				// Checks to see if the episode already exists, in which case we continue to the next
				_, err = os.Stat(show.GetTitle() + string(os.PathSeparator) + seasonMap[season.GetNumber()] + string(os.PathSeparator) + episode.GetFileName() + ".mkv")
				if err == nil {
					color.Green("[ANIRip] " + episode.GetFileName() + ".mkv has already been downloaded successfully..." + "\n")
					continue
				}

				subOffset := 0
				color.Cyan("[ANIRip] Downloading " + episode.GetFileName() + "\n")
				// Downloads full MKV video from stream provider
				color.White("[ANIRip] Downloading video...\n")
				if err := episode.DownloadEpisode(quality, tempDir, session.GetCookies()); err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
				if daisukiIntroTrim {
					subOffset = subOffset + daisukiIntroLength
					color.White("[ANIRip] Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
					if err := trimMKV(daisukiIntroLength, tempDir); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
				if aniplexIntroTrim {
					subOffset = subOffset + aniplexIntroLength
					color.White("[ANIRip] Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
					if err := trimMKV(aniplexIntroLength, tempDir); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
				if sunriseIntroTrim {
					subOffset = subOffset + sunriseIntroLength
					color.White("[ANIRip] Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
					if err := trimMKV(sunriseIntroLength, tempDir); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Downloads the subtitles to .ass format and
				// offsets their times by the passed provided interval
				color.White("[ANIRip] Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
				subtitleLang, err := episode.DownloadSubtitles(language, subOffset, tempDir, session.GetCookies())
				if err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Attempts to merge the downloaded subtitles into the video strea
				color.White("[ANIRip] Merging subtitles into mkv container...\n")
				if err := mergeSubtitles("jpn", subtitleLang, tempDir); err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Moves the episode to the appropriate season sub-directory
				if err := anirip.Rename(tempDir+string(os.PathSeparator)+"episode.mkv",
					show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()]+string(os.PathSeparator)+episode.GetFileName()+".mkv", 10); err != nil {
					color.Red(err.Error() + "\n\n")
				}
				color.Green("[ANIRip] Downloading and merging completed successfully.\n")
			}
		}
		color.Cyan("[ANIRip] Completed processing episodes for " + show.GetTitle() + "\n")
	}
	app.Run(os.Args)
}

func init() {
	// Verifies the existance of an ANIRip folder in our temp directory
	_, err := os.Stat(tempDir)
	if err != nil {
		os.Mkdir(tempDir, 0777)
	}

	// Checks for the existance of our AdobeHDS script which we will get if we don't have it
	_, err = os.Stat(tempDir + string(os.PathSeparator) + "AdobeHDS.php")
	if err != nil {
		adobeHDSResp, err := anirip.GetHTTPResponse("GET", "https://raw.githubusercontent.com/K-S-V/Scripts/master/AdobeHDS.php", nil, nil, nil)
		if err != nil {
			color.Red("[ANIRip] There was an error retrieving AdobeHDS.php script from GitHub...")
			return
		}
		defer adobeHDSResp.Body.Close()
		adobeHDSBody, err := ioutil.ReadAll(adobeHDSResp.Body)
		if err != nil {
			color.Red("[ANIRip] There was an error reading the AdobeHDS.php body...")
			return
		}
		err = ioutil.WriteFile(tempDir+string(os.PathSeparator)+"AdobeHDS.php", adobeHDSBody, 0777)
		if err != nil {
			color.Red("[ANIRip] There was an error writing AdobeHDS.php to file...")
			return
		}
	}
}

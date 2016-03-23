//go:generate goversioninfo -icon=icon.ico

package main

import (
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

const (
	daisukiIntroLength = 5040
	aniplexIntroLength = 6747
	sunriseIntroLength = 8227
)

func main() {
	showURL := ""
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
	app.Author = "Steven Wolfe"
	app.Email = "steven@swolfe.me"
	app.Version = "v1.3.0(3/22/2016)"
	app.Usage = "Crunchyroll/Daisuki show ripper CLI"
	app.Flags = []cli.Flag{
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
	app.Action = func(c *cli.Context) {
		color.Cyan(app.Name + " " + app.Version + " - by " + app.Author + " <" + app.Email + ">\n")
		if c.NArg() > 0 {
			showURL = c.Args()[0]
		} else {
			color.Red("[ANIRip] No show URL provided.")
			return
		}

		// Parses the URL so we can accurately judge the provider based on the host
		color.White("[ANIRip] Parsing the URL passed...")
		url, err := url.Parse(showURL)
		if err != nil {
			color.Red("[ANIRip] There was an error parsing the URL you entered.\n")
			return
		}

		// Creates the authentication & show objects for the provider we're ripping from
		var session anirip.Session
		var show anirip.Show
		if strings.Contains(strings.ToLower(url.Host), "crunchyroll") {
			color.White("[ANIRip] Logging into Crunchyroll as " + username + "...")
			show = new(crunchyroll.CrunchyrollShow)
			session = new(crunchyroll.CrunchyrollSession)
		} else if strings.Contains(strings.ToLower(url.Host), "daisuki") {
			color.White("[ANIRip] Logging into Daisuki as " + username + "...")
			show = new(daisuki.DaisukiShow)
			session = new(daisuki.DaisukiSession)
		} else {
			color.Red("[ANIRip] The URL provided is not supported.")
			return
		}

		// Performs the generic login procedure
		if err = session.Login(username, password); err != nil {
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

		for _, season := range show.GetSeasons() {
			for _, episode := range season.GetEpisodes() {
				color.White("[ANIRip] Getting Episode Info...\n")
				if err = episode.GetEpisodeInfo(quality, session.GetCookies()); err != nil {
					color.Red("[ANIRip] " + err.Error())
					continue
				}

				// Checks to see if the episode already exists, in which case we continue to the next
				_, err = os.Stat(episode.GetFileName() + ".mkv")
				if err == nil {
					color.Green("[ANIRip] " + episode.GetFileName() + ".mkv has already been downloaded successfully..." + "\n")
					continue
				}

				subOffset := 0
				color.Cyan("[ANIRip] Downloading " + episode.GetFileName() + "\n")
				// Downloads full MKV video from stream provider
				color.White("[ANIRip] Downloading video...\n")
				if err := episode.DownloadEpisode(quality, session.GetCookies()); err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
				if daisukiIntroTrim {
					subOffset = subOffset + daisukiIntroLength
					color.White("[ANIRip] Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
					if err := trimMKV(daisukiIntroLength); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
				if aniplexIntroTrim {
					subOffset = subOffset + aniplexIntroLength
					color.White("[ANIRip] Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
					if err := trimMKV(aniplexIntroLength); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
				if sunriseIntroTrim {
					subOffset = subOffset + sunriseIntroLength
					color.White("[ANIRip] Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
					if err := trimMKV(sunriseIntroLength); err != nil {
						color.Red("[ANIRip] " + err.Error() + "\n")
						continue
					}
				}

				// Downloads the subtitles to .ass format and
				// offsets their times by the passed provided interval
				color.White("[ANIRip] Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
				subtitleLang, err := episode.DownloadSubtitles(language, subOffset, session.GetCookies())
				if err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Attempts to merge the downloaded subtitles into the video strea
				color.White("[ANIRip] Merging subtitles into mkv container...\n")
				if err := mergeSubtitles("jpn", subtitleLang); err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Cleans and optimizes the final MKV
				color.White("[ANIRip] Cleaning and optimizing mkv...\n")
				if err := cleanMKV(); err != nil {
					color.Red("[ANIRip] " + err.Error() + "\n")
					continue
				}

				// Moves the episode to the appropriate directory
				if err := anirip.Rename("episode.mkv", episode.GetFileName()+".mkv", 10); err != nil {
					color.Red(err.Error() + "\n\n")
				}
				color.Green("[ANIRip] Downloading and merging completed successfully.\n")
			}
		}
		color.Cyan("[ANIRip] Completed processing episodes for " + show.GetTitle() + "\n")
	}
	app.Run(os.Args)
}

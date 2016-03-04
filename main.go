//go:generate goversioninfo -icon=icon.ico

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sdwolfe32/ANIRip/anirip"
	"github.com/sdwolfe32/ANIRip/crunchyroll"
	"github.com/sdwolfe32/ANIRip/daisuki"
)

const (
	daisukiIntroLength = 5040
	aniplexIntroLength = 6747
	sunriseIntroLength = 8227
)

var (
	session          anirip.Session
	show             anirip.Show
	showURL          = ""
	username         = ""
	password         = ""
	daisukiIntroTrim = "n"
	aniplexIntroTrim = "n"
	sunriseIntroTrim = "n"
)

func login(session anirip.Session, username, password string) {
	fmt.Printf("\n")
	if err := session.Login(username, password); err != nil {
		fmt.Println("Please login to the above provider.")
		getStandardUserInput("Username : ", &username)
		getStandardUserInput("Password : ", &password)
		login(session, username, password)
	}
	return
}

func main() {
	// Intro header
	fmt.Printf("-------------------------------------------------------------\n")
	fmt.Printf("-------------------- ANIRip v1.1 by Viz_ --------------------\n")
	fmt.Printf("-------------------------------------------------------------\n\n")
	fmt.Printf("- Currently supports Crunchyroll & Daisuki show URLs\n\n")

	// Asks the user for their desired show URL
	getStandardUserInput("Enter a Show URL : ", &showURL)

	// url, err := url.Parse(showURL)
	// if err != nil {
	// 	return anirip.Error{Message: "There was an error parsing episode information", Err: err}
	// }

	// Creates the authentication & show objects for the provider we're ripping from
	if strings.Contains(strings.ToLower(showURL), "crunchyroll") {
		show = new(crunchyroll.CrunchyrollShow)
		session = new(crunchyroll.CrunchyrollSession)
	} else if strings.Contains(strings.ToLower(showURL), "daisuki") {
		show = new(daisuki.DaisukiShow)
		session = new(daisuki.DaisukiSession)
	}

	// Performs the generic login procedure
	login(session, username, password)

	// Attempts to scrape the show and episode info
	if err := show.ScrapeEpisodes(showURL, session.GetCookies()); err != nil {
		fmt.Println(err)
		pause()
		return
	}

	// Asks if the user wants to trim any intro clips off of the final file IF provider is daisuki
	if strings.Contains(strings.ToLower(showURL), "daisuki") {
		getStandardUserInput("Do you want to trim a Daisuki intro? [Y/N] : ", &daisukiIntroTrim)
		getStandardUserInput("Do you want to trim an Aniplex intro? [Y/N] : ", &aniplexIntroTrim)
		getStandardUserInput("Do you want to trim a Sunrise intro? [Y/N] : ", &sunriseIntroTrim)
	}

	fmt.Printf("\n")
	for _, season := range show.GetSeasons() {
		for _, episode := range season.GetEpisodes() {
			subOffset := 0
			fmt.Printf("Downloading " + episode.GetFileName() + "\n\n")

			// Downloads full MKV video from stream provider
			fmt.Printf("Downloading video...\n")
			if err := episode.DownloadEpisode("1080p", session.GetCookies()); err != nil {
				fmt.Printf(err.Error() + "\n\n")
				continue
			}

			// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
			if strings.ToLower(daisukiIntroTrim) == "y" {
				subOffset = subOffset + daisukiIntroLength
				fmt.Printf("Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
				if err := trimMKV(episode.GetFileName(), daisukiIntroLength, 6200); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}
			}

			// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
			if strings.ToLower(aniplexIntroTrim) == "y" {
				subOffset = subOffset + aniplexIntroLength
				fmt.Printf("Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
				if err := trimMKV(episode.GetFileName(), aniplexIntroLength, 9000); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}
			}

			// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
			if strings.ToLower(sunriseIntroTrim) == "y" {
				subOffset = subOffset + sunriseIntroLength
				fmt.Printf("Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
				if err := trimMKV(episode.GetFileName(), sunriseIntroLength, 10000); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}
			}

			// Downloads the subtitles to .ass format and
			// offsets their times by the passed provided interval
			fmt.Printf("Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
			if err := episode.DownloadSubtitles("English", subOffset, session.GetCookies()); err != nil {
				fmt.Printf(err.Error() + "\n\n")
				continue
			}

			// Attempts to merge the downloaded subtitles into the video stream
			fmt.Printf("Merging subtitles into mkv container...\n")
			if err := mergeSubtitles(episode.GetFileName(), "jpn", "eng"); err != nil {
				fmt.Printf(err.Error() + "\n\n")
				continue
			}

			// Cleans and optimizes the final MKV
			fmt.Printf("Cleaning and optimizing mkv...\n")
			if err := cleanMKV(episode.GetFileName()); err != nil {
				fmt.Printf(err.Error() + "\n\n")
				continue
			}
			fmt.Printf("Downloading and merging completed successfully.\n\n")
		}
	}
	fmt.Printf("Completed processing episodes for " + show.GetTitle() + "\n\n")
	pause()
}

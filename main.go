//go:generate goversioninfo -icon=icon.ico

package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sdwolfe32/ANIRip/anirip"
	"github.com/sdwolfe32/ANIRip/crunchyroll"
	"github.com/sdwolfe32/ANIRip/daisuki"
)

const (
	daisukiIntroLength = 5040
	aniplexIntroLength = 6747
	sunriseIntroLength = 8227
	engineDir          = "engine"   // Where our clis are required to be
	cookieDir          = "cookies"  // Where we will store the cookies
	tempDir            = "temp"     // Where our temporary video/audio/subtitle streams will be held
	outputDir          = "finished" // Where our finalized show directories will be held
)

func main() {
	// Asks the user for their desired show URLs delimited by spaces
	showURLs := ""
	getStandardUserInput("Enter a list of Show URLs (delimited by spaces) : ", &showURLs)
	showURLArray := strings.Split(showURLs, " ")

	// Iterates over every showURL found in the array
	for _, showURL := range showURLArray {
		// Parses the URL so we can accurately judge the provider based on the host
		url, err := url.Parse(showURL)
		if err != nil {
			color.Red("There was an error parsing the URL you entered.\n")
			main()
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
		}

		// Performs the generic login procedure
		username := ""
		password := ""
		login(session, username, password, cookieDir)

		// Attempts to scrape the shows metadata/information
		if err := show.ScrapeEpisodes(showURL, session.GetCookies()); err != nil {
			color.Red(err.Error() + "\n\n")
			pause()
			return
		}

		// Asks if the user wants to trim any intro clips off of the final file IF provider is daisuki
		daisukiIntroTrim := "n"
		aniplexIntroTrim := "n"
		sunriseIntroTrim := "n"
		if strings.Contains(strings.ToLower(showURL), "daisuki") {
			getStandardUserInput("Do you want to trim a Daisuki intro? [Y/N] : ", &daisukiIntroTrim)
			getStandardUserInput("Do you want to trim an Aniplex intro? [Y/N] : ", &aniplexIntroTrim)
			getStandardUserInput("Do you want to trim a Sunrise intro? [Y/N] : ", &sunriseIntroTrim)
			fmt.Printf("\n")
		}

		// Sets up an array that we will use for season directory naming
		seasonNameArray := []string{
			"Specials",   // Season num 0
			"Season One", // Season num 1...
			"Season Two",
			"Season Three",
			"Season Four",
			"Season Five",
			"Season Six",
			"Season Seven",
			"Season Eight",
			"Season Nine",
			"Season Ten",
		}

		// Creates a folder in the temporary directory that will store the seasons
		os.Mkdir(tempDir+"\\"+show.GetTitle(), 0777)
		for _, season := range show.GetSeasons() {
			// Creates a folder in the show directory that will store the episodes
			os.Mkdir(tempDir+"\\"+show.GetTitle()+"\\"+seasonNameArray[season.GetNumber()], 0777)
			for _, episode := range season.GetEpisodes() {
				color.Cyan("> Getting Episode Info...\n")
				if err := episode.GetEpisodeInfo("1080p", session.GetCookies()); err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
					continue
				}

				// Checks to see if the episode already exists, in which case we continue to the next
				_, err = os.Stat(tempDir + "\\" + show.GetTitle() + "\\" + seasonNameArray[season.GetNumber()] + "\\" + episode.GetFileName() + ".mkv")
				if err == nil {
					color.Cyan(episode.GetFileName() + ".mkv has already been downloaded successfully..." + "\n\n")
					continue
				}

				subOffset := 0
				color.Cyan("> Downloading " + episode.GetFileName() + "\n\n")
				// Downloads full MKV video from stream provider
				color.Green("> Downloading video...\n")
				if err := episode.DownloadEpisode("1080p", engineDir, tempDir, session.GetCookies()); err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
					continue
				}

				// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
				if strings.ToLower(daisukiIntroTrim) == "y" {
					subOffset = subOffset + daisukiIntroLength
					color.Green("> Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
					if err := trimMKV(daisukiIntroLength, engineDir, tempDir); err != nil {
						color.Red(err.Error() + "\n\n")
						pause()
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
				if strings.ToLower(aniplexIntroTrim) == "y" {
					subOffset = subOffset + aniplexIntroLength
					color.Green("> Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
					if err := trimMKV(aniplexIntroLength, engineDir, tempDir); err != nil {
						color.Red(err.Error() + "\n\n")
						pause()
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
				if strings.ToLower(sunriseIntroTrim) == "y" {
					subOffset = subOffset + sunriseIntroLength
					color.Green("> Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
					if err := trimMKV(sunriseIntroLength, engineDir, tempDir); err != nil {
						color.Red(err.Error() + "\n\n")
						pause()
						continue
					}
				}

				// Downloads the subtitles to .ass format and
				// offsets their times by the passed provided interval
				color.Green("> Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
				subtitleLang, err := episode.DownloadSubtitles("English", subOffset, tempDir, session.GetCookies())
				if err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
					continue
				}

				// Attempts to merge the downloaded subtitles into the video strea
				color.Green("> Merging subtitles into mkv container...\n")
				if err := mergeSubtitles("jpn", subtitleLang, engineDir, tempDir); err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
					continue
				}

				// Cleans and optimizes the final MKV
				color.Green("> Cleaning and optimizing mkv...\n")
				if err := cleanMKV(engineDir, tempDir); err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
					continue
				}

				// Moves the episode to the appropriate directory
				if err := anirip.Rename(tempDir+"\\episode.mkv",
					tempDir+"\\"+show.GetTitle()+"\\"+seasonNameArray[season.GetNumber()]+"\\"+episode.GetFileName()+".mkv", 10); err != nil {
					color.Red(err.Error() + "\n\n")
					pause()
				}
				color.Green("> Downloading and merging completed successfully.\n\n")
			}
		}
		if err := anirip.Rename(tempDir+"\\"+show.GetTitle(), outputDir+"\\"+show.GetTitle(), 10); err != nil {
			color.Red(err.Error() + "\n\n")
			pause()
		}
		color.Cyan("> Completed processing episodes for " + show.GetTitle() + "\n\n")
	}
	pause()
}

func init() {
	// Intro header
	color.Green("-------------------------------------------------------------\n")
	color.Green("------------------- ANIRip v1.2.0 by Viz_ -------------------\n")
	color.Green("-------------------------------------------------------------\n\n")
	color.Green("- Currently supports Crunchyroll & Daisuki show URLs\n\n")

	// Checks for the existance of our required output folders and creates them if they don't
	_, err := os.Stat(outputDir)
	if err != nil {
		os.Mkdir(outputDir, 0777)
	}

	_, err = os.Stat(tempDir)
	if err != nil {
		os.Mkdir(tempDir, 0777)
	}

	_, err = os.Stat(cookieDir)
	if err != nil {
		os.Mkdir(cookieDir, 0777)
	}
}

func login(session anirip.Session, username, password, cookieDir string) {
	fmt.Printf("\n")
	if err := session.Login(username, password, cookieDir); err != nil {
		color.Green("Please login to the above provider.")
		getStandardUserInput("Username : ", &username)
		getStandardUserInput("Password : ", &password)
		login(session, username, password, cookieDir)
	}
	return
}

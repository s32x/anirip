//go:generate goversioninfo -icon=icon.ico

package main

import (
	"fmt"
	"net/url"
	"os"
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
	cookieDir          = "cookies"
	engineDir          = "engine"
	tempDir            = "temp"
	outputDir          = "output"
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
			fmt.Println("There was an error parsing the URL you entered.")
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
			fmt.Println(err)
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
		}

		fmt.Printf("\n")

		// Sets up an array that we will use for season directory naming
		seasonNameArray := []string{"Specials", // Season num 0
			"Season One", // Season num 1...
			"Season Two",
			"Season Three",
			"Season Four",
			"Season Five",
			"Season Six",
			"Season Seven",
			"Season Eight",
			"Season Nine",
			"Season Ten"}

		// Creates a folder in the temporary directory that will store the seasons
		os.Mkdir(tempDir+"\\"+show.GetTitle(), 0777)
		for _, season := range show.GetSeasons() {
			// Creates a folder in the show directory that will store the episodes
			os.Mkdir(tempDir+"\\"+show.GetTitle()+"\\"+seasonNameArray[season.GetNumber()], 0777)
			for _, episode := range season.GetEpisodes() {
				fmt.Printf("Getting Episode Info...\n")
				if err := episode.GetEpisodeInfo("1080p", session.GetCookies()); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Checks to see if the episode already exists, in which case we continue to the next
				_, err = os.Stat(tempDir + "\\" + show.GetTitle() + "\\" + seasonNameArray[season.GetNumber()] + "\\" + episode.GetFileName() + ".mkv")
				if err == nil {
					fmt.Printf(episode.GetFileName() + ".mkv has already been downloaded successfully..." + "\n\n")
					continue
				}

				subOffset := 0
				fmt.Printf("Downloading " + episode.GetFileName() + "\n\n")
				// Downloads full MKV video from stream provider
				fmt.Printf("Downloading video...\n")
				if err := episode.DownloadEpisode("1080p", engineDir, tempDir, session.GetCookies()); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
				if strings.ToLower(daisukiIntroTrim) == "y" {
					subOffset = subOffset + daisukiIntroLength
					fmt.Printf("Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
					if err := trimMKV(episode.GetFileName(), daisukiIntroLength, 6200, engineDir, tempDir); err != nil {
						fmt.Printf(err.Error() + "\n\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
				if strings.ToLower(aniplexIntroTrim) == "y" {
					subOffset = subOffset + aniplexIntroLength
					fmt.Printf("Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
					if err := trimMKV(episode.GetFileName(), aniplexIntroLength, 9000, engineDir, tempDir); err != nil {
						fmt.Printf(err.Error() + "\n\n")
						continue
					}
				}

				// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
				if strings.ToLower(sunriseIntroTrim) == "y" {
					subOffset = subOffset + sunriseIntroLength
					fmt.Printf("Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
					if err := trimMKV(episode.GetFileName(), sunriseIntroLength, 10000, engineDir, tempDir); err != nil {
						fmt.Printf(err.Error() + "\n\n")
						continue
					}
				}

				// Downloads the subtitles to .ass format and
				// offsets their times by the passed provided interval
				fmt.Printf("Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
				subtitleLang, err := episode.DownloadSubtitles("English", subOffset, tempDir, session.GetCookies())
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Attempts to merge the downloaded subtitles into the video stream
				fmt.Printf("Merging subtitles into mkv container...\n")
				if err := mergeSubtitles(episode.GetFileName(), "jpn", subtitleLang, engineDir, tempDir); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Cleans and optimizes the final MKV
				fmt.Printf("Cleaning and optimizing mkv...\n")
				if err := cleanMKV(episode.GetFileName(), engineDir, tempDir); err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Moves the episode to the appropriate directory
				if err := anirip.Rename(tempDir+"\\"+episode.GetFileName()+".mkv",
					tempDir+"\\"+show.GetTitle()+"\\"+seasonNameArray[season.GetNumber()]+"\\"+episode.GetFileName()+".mkv", 10); err != nil {
					fmt.Printf(err.Error() + "\n\n")
				}
				fmt.Printf("Downloading and merging completed successfully.\n\n")
			}
		}
		if err := anirip.Rename(tempDir+"\\"+show.GetTitle(), outputDir+"\\"+show.GetTitle(), 10); err != nil {
			fmt.Printf(err.Error() + "\n\n")
		}
		fmt.Printf("Completed processing episodes for " + show.GetTitle() + "\n\n")
	}
	pause()
}

func init() {
	// Intro header
	fmt.Printf("-------------------------------------------------------------\n")
	fmt.Printf("------------------- ANIRip v1.1.2 by Viz_ -------------------\n")
	fmt.Printf("-------------------------------------------------------------\n\n")
	fmt.Printf("- Currently supports Crunchyroll & Daisuki show URLs\n\n")

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
		fmt.Println("Please login to the above provider.")
		getStandardUserInput("Username : ", &username)
		getStandardUserInput("Password : ", &password)
		login(session, username, password, cookieDir)
	}
	return
}

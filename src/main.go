package main

import (
	"fmt"
	"strings"
)

var (
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36"
	cookiesFile = "cookies.txt" // store cookies in "cookies" file
)

func main() {
	// Displays the usual title banner for the application
	fmt.Printf("\n-------------------------------------------------------------------------\n")
	fmt.Printf("----------------======== CrunchyRip v0.3 by Viz_ ========----------------\n")
	fmt.Printf("-------------------------------------------------------------------------\n\n")

	params := SessionParameters{}
	// Ask the user first if they would like their session to use cookies
	getStandardUserInput("Would you like to log into your Crunchyroll account [Y/N]? ", &params.AccountStatus)
	if strings.ToLower(params.AccountStatus) == "y" {
		err := authenticate(&params.Cookies)
		if err != nil {
			fmt.Printf(err.Error() + "\n\n")
		}
	}

	for {
		// Handles getting all show information after asking the user
		err := getShow(&params.Show)
		if err != nil || len(params.Show.Seasons) == 0 {
			fmt.Printf(err.Error() + "\n\n")
			continue
		}

		params.Preferences.DesiredSeasons = "all"
		params.Preferences.DesiredLanguage = "all"
		params.Preferences.DesiredLanguage = "English"
		params.Preferences.DesiredQuality = "1080p"

		fmt.Printf("\n")

		// Our loop that performs the downloading of all our episodes
		for _, season := range params.Show.Seasons {
			for _, episode := range season.Episodes {
				// Sets the next episode filename based on where we are in the loop
				err := setEpisodeFileName(params.Show.Title, season.Number, &episode)
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Perfroms all episode/subtitle downloading, splitting and merging
				err = getEpisode(&episode, &params)
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}
			}
			// Move episodes into Season folder
		}
		// Move seasons into Show folder
		// Move show folder into output folder
	}
}

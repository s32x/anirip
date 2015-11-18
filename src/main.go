package main

import (
	"fmt"
	"strconv"
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
		err := login(&params.Cookies)
		if err != nil {
			fmt.Printf(err.Error() + "\n\n")
		}
	}

	for {
		// First we ask the user for what show they would like to download
		getStandardUserInput("Enter a show name : ", &params.SearchTerm)

		// First we get the showname/path of the show we would like to download
		show, err := searchShowPath(params.SearchTerm)
		if err != nil || show.URL == "" {
			fmt.Printf(err.Error() + "\n\n")
			continue
		}
		fmt.Println("\nDetermined a valid show name of : --- " + show.Title + " ---")

		// Gets the episodes for the show recieved
		show, err = getEpisodes(show)
		if err != nil || len(show.Seasons) == 0 || len(show.Seasons[0].Episodes) == 0 {
			fmt.Printf(err.Error() + "\n\n")
			continue
		}

		// Attempts to access and print the titles of all seasons recieved
		fmt.Printf("Below is a list of seasons found ...\n\n")
		for i := 0; i < len(show.Seasons); i++ {
			fmt.Printf("\tSeason " + strconv.Itoa(show.Seasons[i].Number) + " - " + show.Seasons[i].Title + " (" + strconv.Itoa(show.Seasons[i].Length) + " Episodes)\n")
		}

		// Gets the users desired settings
		// fmt.Printf("\n")
		// GetStandardUserInput("Enter the seasons you wish to download [EX:'2,3,6' or 'all']: ", &params.DesiredSeasons)
		// GetStandardUserInput("Enter the episodes you wish to download [EX:'2,3,6' or 'all']: ", &params.DesiredLanguage)
		// GetStandardUserInput("Enter a subtitle language ['NONE' for no subs]: ", &params.DesiredLanguage)
		// GetStandardUserInput("Enter your desired video quality [EX:'1080p']: ", &params.DesiredQuality)
		params.DesiredSeasons = "all"
		params.DesiredLanguage = "all"
		params.DesiredLanguage = "English"
		params.DesiredQuality = "1080p"

		fmt.Printf("\n")
		for _, season := range show.Seasons {
			for _, episode := range season.Episodes {
				episodeFileName := getEpisodeFileName(show.Title, strconv.Itoa(season.Number), strconv.FormatFloat(episode.Number, 'f', -1, 64), episode.Description)

				// Download episodes to file
				err = downloadEpisode(episodeFileName, episode, params)
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Download subtitles to file
				err = downloadSubtitle(episodeFileName, episode, params)
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}

				// Merge episode with subtitle and clean up temp directory
				err = splitMergeAndClean(episodeFileName, episode)
				if err != nil {
					fmt.Printf(err.Error() + "\n\n")
					continue
				}
				fmt.Printf("\n\n")
			}
			// Move episodes into Season folder
		}
		// Move seasons into Show folder

		// Move show folder into output folder
	}
}

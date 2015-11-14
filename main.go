package main

import (
	"fmt"
	"strconv"
)

var (
	showSearchTerm      = ""
	showdesiredSeasons  = ""
	showDesiredQuality  = ""
	showDesiredLanguage = ""
	hostURL             = "http://www.crunchyroll.com"
	searchURL           = hostURL + "/search?from=&q="
)

func main() {
	// Displays the usual title banner for the application
	fmt.Printf("\n-------------------------------------------------------------------------\n")
	fmt.Printf("----------------======== CrunchyRip v0.1 by Viz_ ========----------------\n")
	fmt.Printf("-------------------------------------------------------------------------\n\n")

	//TODO Form a VPN connection and output the IP address what will be used for the connection

	// Ask the user first if they would like their session to use cookies
	accountStatus := ""
	GetStandardUserInput("Would you like to log into your Crunchyroll account [Y/N]?", &accountStatus)
	if accountStatus == "Y" || accountStatus == "y" {
		loginStatus, err := login()
		if !loginStatus || err != nil {
			fmt.Println(">>> There was an issue while attempting to log in : ", err)
		}
	}

	for {
		// First we as the user for what show they would like to rip
		GetStandardUserInput("Enter a show name : ", &showSearchTerm)

		// First we get the showname/path of the show we would like to download
		show, err := searchShowPath(showSearchTerm)
		if err != nil || show.URL == "" {
			fmt.Println(">>> Unable to get a show name/path via search results. \n", err)
			continue
		}
		fmt.Println("\nDetermined a valid show name of : --- " + show.Title + " ---")

		// Gets the episodes for the show recieved
		show, err = getEpisodes(show)
		if err != nil || len(show.Seasons) == 0 || len(show.Seasons[0].Episodes) == 0 {
			fmt.Println(">>> Unable to get any episodes for the show specified. \n", err)
			continue
		}

		// Attempts to access and print the titles of all seasons recieved
		fmt.Println("Below is a list of seasons found ...\n")
		for i := 0; i < len(show.Seasons); i++ {
			fmt.Println("\tSeason " + strconv.Itoa(show.Seasons[i].Number) + " - " + show.Seasons[i].Title + " (" + strconv.Itoa(show.Seasons[i].Length) + " Episodes)")
		}

		// Gets the users desired settings
		GetStandardUserInput("\nEnter the seasons you wish to download : ", &showdesiredSeasons)
		GetStandardUserInput("\nEnter a subtitle language ('NONE' for no subs) : ", &showDesiredLanguage)
		GetStandardUserInput("\nEnter your desired video quality : ", &showDesiredQuality)

		//TODO RTMP Dumps each episode in a seperate goroutine...
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"
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
	fmt.Println("\n-------------------------------------------------------------------------")
	fmt.Println("----------------======== CrunchyRip v0.1 by Viz_ ========----------------")
	fmt.Println("-------------------------------------------------------------------------\n")

	//TODO Form a VPN connection and output the IP address what will be used for the connection

	// Asks if the user has an account and if they would login to get premium content
	accountStatus := ""
	getStandardUserInput("Do you have a CrunchyRoll Premium account [Y/N]?", &accountStatus)

	username := ""
	password := ""
	fmt.Println("First, Please login to your CrunchyRoll account in order to access your Premium content.\n")
	getStandardUserInput("Username : ", &username)
	getStandardUserInput("Password : ", &password)
	login(username, password)

	for {
		// First we as the user for what show they would like to rip
		getStandardUserInput("Enter a show name : ", &showSearchTerm)

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
		getStandardUserInput("\nEnter the seasons you wish to download : ", &showdesiredSeasons)
		getStandardUserInput("\nEnter a subtitle language ('NONE' for no subs) : ", &showDesiredLanguage)
		getStandardUserInput("\nEnter your desired video quality : ", &showDesiredQuality)

		//TODO RTMP Dumps each episode in a seperate goroutine...
	}
}

// Gets user input from the user and unmarshalls it into the input
func getStandardUserInput(prefixText string, input *string) error {
	fmt.Printf(prefixText)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		*input = scanner.Text()
		break
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

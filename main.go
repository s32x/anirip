package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var (
	showSearchTerm = ""
	hostURL        = "http://www.crunchyroll.com"
	searchURL      = hostURL + "/search?from=&q="
)

func main() {
	for {
		//TODO Form a VPN connection and output the IP address what will be used for the connection

		//TODO Request the users login credentials and build a cookie for required auth requests

		// First we as the user for what show they would like to rip
		fmt.Println("\n")
		getStandardUserInput("Please enter a show name : ", &showSearchTerm)

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
		fmt.Println("\nBelow is a list of all seasons found ...\n")
		for i := 0; i < len(show.Seasons); i++ {
			fmt.Println("\tSeason " + strconv.Itoa(show.Seasons[i].Number) + " - " + show.Seasons[i].Title + " (" + strconv.Itoa(len(show.Seasons[i].Episodes)) + " Episodes)")
		}

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

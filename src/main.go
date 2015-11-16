package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var (
	showSearchTerm      = ""
	showdesiredSeasons  = ""
	showDesiredQuality  = ""
	showDesiredLanguage = ""
	userAgent           = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36"
	cookiesFile         = "cookies.txt" // store cookies in "cookies" file
	userCookies         = []*http.Cookie{}
)

func main() {
	// Displays the usual title banner for the application
	fmt.Printf("\n-------------------------------------------------------------------------\n")
	fmt.Printf("----------------======== CrunchyRip v0.1 by Viz_ ========----------------\n")
	fmt.Printf("-------------------------------------------------------------------------\n\n")

	//TODO Form a VPN connection and output the IP address what will be used for the connection

	// Ask the user first if they would like their session to use cookies
	accountStatus := ""
	GetStandardUserInput("Would you like to log into your Crunchyroll account [Y/N]? ", &accountStatus)
	if accountStatus == "Y" || accountStatus == "y" {
		cookies, err := login()
		userCookies = cookies
		if len(userCookies) == 0 {
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
		fmt.Printf("Below is a list of seasons found ...\n\n")
		for i := 0; i < len(show.Seasons); i++ {
			fmt.Println("\tSeason " + strconv.Itoa(show.Seasons[i].Number) + " - " + show.Seasons[i].Title + " (" + strconv.Itoa(show.Seasons[i].Length) + " Episodes)")
		}

		// Gets the users desired settings
		GetStandardUserInput("\nEnter the seasons you wish to download : ", &showdesiredSeasons)
		GetStandardUserInput("\nEnter a subtitle language ('NONE' for no subs) : ", &showDesiredLanguage)
		GetStandardUserInput("\nEnter your desired video quality : ", &showDesiredQuality)

		for _, season := range show.Seasons {
			for _, episode := range season.Episodes {
				episodeFileName := cleanFileName(show.Title + " - " + "S" + strconv.Itoa(season.Number) + "E" + strings.Split(strconv.FormatFloat(episode.Number, 'E', -1, 64), "E")[0] + " - " + episode.Description)
				downloadEpisode(episode, userCookies, episodeFileName)
				downloadSubtitle(showDesiredLanguage, episode, userCookies, episodeFileName)
				splitMergeAndClean(episodeFileName, episode)
			}
		}
	}
}

func downloadEpisode(episode Episode, userCookies []*http.Cookie, episodeFileName string) error {
	// First attempts to get the XML attributes for the requested episode
	episodeRTMPInfo, err := getRMTPInfo(episode, userCookies)
	if err != nil {
		fmt.Println(">>> There was an issue while getting XML for "+episode.Title, err)
		return err
	}

	// Attempts to dump the FLV of the episode to file
	err = dumpEpisodeFLV(episodeRTMPInfo, episode.URL, episodeFileName)
	if err != nil {
		fmt.Println(">>> There was an issue while trying to rip the FLV for "+episode.Title, err)
		return err
	}
	return nil
}

func downloadSubtitle(showDesiredLanguage string, episode Episode, userCookies []*http.Cookie, episodeFileName string) error {
	// Populates the subtitle info for the episode
	subtitle, err := getSubtitleInfo(showDesiredLanguage, episode, userCookies)
	if err != nil {
		fmt.Println(">>> There was an error while getting Subtitle for "+episode.Title+" :", err)
		return err
	}

	// Places the new subtitle object with JUST INFO into the episode and gets the sub data
	subtitle, err = getSubtitleData(subtitle, episode, userCookies)
	if err != nil {
		fmt.Println(">>> There was an error while getting subtitle XML for "+episode.Title+" :", err)
		return err
	}

	// Dumps our final subtitle string into an ass file for merging later on
	err = dumpSubtitleASS(subtitle, episode, episodeFileName)
	if err != nil {
		fmt.Println(">>> There was an error while creating subtitles for "+episode.Title+" :", err)
		return err
	}
	return nil
}

func splitMergeAndClean(episodeFileName string, episode Episode) error {
	// Splits up the FLV file so we can handle all peices with mergemkv
	err := splitEpisodeFLV(episodeFileName)
	if err != nil {
		fmt.Println(">>> There was an issue while trying to split the FLV for "+episode.Title, err)
		return err
	}

	// Merges all the files together to create a single solid MKV
	err = mergeEpisodeMKV(episodeFileName)
	if err != nil {
		fmt.Println(">>> There was an issue while trying to merge the MKV for "+episode.Title, err)
		return err
	}
	return nil
}

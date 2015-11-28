//go:generate goversioninfo -icon=icon.ico

package main

import "fmt"

func main() {
	session := new(Session)
	// Intro header
	fmt.Printf("------------------------------------------------------------\n")
	fmt.Printf("------------------ CrunchyRip 1.0 by Viz_ ------------------\n")
	fmt.Printf("------------------------------------------------------------\n")

	username := ""
	password := ""
	fmt.Printf("Please login to Crunchyroll below...\n")
	getStandardUserInput("Username : ", &username)
	getStandardUserInput("Password : ", &password)

	err := session.Login(username, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	show := new(Show)

	desiredShow := ""
	getStandardUserInput("Enter a Show name : ", &desiredShow)
	show.FindShow(desiredShow, session.Cookies)

	// Gets all episodes associated with the show found
	show.GetEpisodes(session.Cookies)
	// TODO Create new show folder
	for _, season := range show.Seasons {
		// TODO Create new season folder
		for _, episode := range season.Episodes {
			fmt.Printf(episode.FileName + "\n")

			fmt.Printf("Downloading video...\n")
			err = episode.DownloadEpisode("1080p", session.Cookies)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Downloading subtitles...\n")
			err = episode.DownloadSubtitles("English", session.Cookies)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Splitting video...\n")
			err = Split(episode.FileName)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Merging video...\n")
			err = Merge(episode.FileName)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Cleaning video...\n")
			err = Clean(episode.FileName)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Downloading and merging completed successfully.\n\n")
			// TODO Move shows
		}
	}
}

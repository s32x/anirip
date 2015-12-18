//go:generate goversioninfo -icon=icon.ico

package main

import "fmt"

func main() {
	session := new(Session)

	// Intro header
	fmt.Printf("------------------------------------------------------------\n")
	fmt.Printf("------------------ CrunchyRip 1.1 by Viz_ ------------------\n")
	fmt.Printf("------------------------------------------------------------\n")

	// Attempts to log the user in / request login credentials
	if err := session.Login(); err != nil {
		fmt.Println(err)
		pause()
		return
	}

	// Asks the user for their desired show and builds out a crunchyroll show object
	show := new(Show)
	desiredShow := ""
	getStandardUserInput("Enter a Show name : ", &desiredShow)

	// Finds the show and gets all the episodes associated with it
	show.FindShow(desiredShow, session.Cookies)
	show.GetEpisodes(session.Cookies)

	// TODO Create new show folder
	for _, season := range show.Seasons {
		// TODO Create new season folder
		for _, episode := range season.Episodes {
			fmt.Printf(episode.FileName + "\n")

			fmt.Printf("Downloading video...\n")
			if err := episode.DownloadEpisode("1080p", session.Cookies); err != nil {
				fmt.Println(err)
				pause()
				break
			}

			fmt.Printf("Downloading subtitles...\n")
			if err := episode.DownloadSubtitles("English", session.Cookies); err != nil {
				fmt.Println(err)
				pause()
				break
			}

			fmt.Printf("Splitting video...\n")
			if err := Split(episode.FileName); err != nil {
				fmt.Println(err)
				pause()
				break
			}

			fmt.Printf("Merging video...\n")
			if err := Merge(episode.FileName); err != nil {
				fmt.Println(err)
				pause()
				break
			}

			fmt.Printf("Cleaning video...\n")
			if err := Clean(episode.FileName); err != nil {
				fmt.Println(err)
				pause()
				break
			}
			fmt.Printf("Downloading and merging completed successfully.\n\n")
			// TODO Move shows
		}
	}
	fmt.Printf("Completed processing episodes for " + show.Title + "\n\n")
	pause()
}

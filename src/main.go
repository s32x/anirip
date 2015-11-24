package main

import "fmt"

func main() {
	session := new(Session)
	_ = session.Login("", "")

	show := new(Show)
	show.FindShow("the world is still beautiful", session.Cookies)
	show.GetEpisodes(session.Cookies)
	// Creates new show folder
	for _, season := range show.Seasons {
		// Creates new season folder
		for _, episode := range season.Episodes {
			fmt.Printf(episode.FileName + "\n")
			fmt.Printf("Downloading video...\n")
			episode.DownloadEpisode("1080p", session.Cookies)
			fmt.Printf("Downloading subtitles...\n")
			episode.DownloadSubtitles("English", session.Cookies)
			fmt.Printf("Splitting video...\n")
			Split(episode.FileName)
			fmt.Printf("Merging video...\n")
			Merge(episode.FileName)
			fmt.Printf("Cleaning video...\n")
			Clean(episode.FileName)
			fmt.Printf("Downloading and merging completed successfully.\n\n")
			// Moves show
		}
	}
}

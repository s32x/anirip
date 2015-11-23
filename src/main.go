package main

import "fmt"

func main() {
	session := new(Session)
	_ = session.Login("", "")

	show := new(Show)
	show.FindShow("sword art")
	show.GetEpisodes()
	for _, season := range show.Seasons {
		for _, episode := range season.Episodes {
			fmt.Printf("Downloading episode - " + episode.FileName + "\n")
			episode.DownloadEpisode("1080p", session.Cookies)
			fmt.Printf("Downloading subtitles...\n")
			episode.DownloadSubtitles("English", session.Cookies)
			fmt.Printf("Splitting video...\n")
			episode.Split()
			fmt.Printf("Merging video...\n")
			episode.Merge()
			fmt.Printf("Downloading and merging completed successfully.\n\n")
		}
	}
}

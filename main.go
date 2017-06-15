package main

import (
	"github.com/sdwolfe32/anirip/anirip"
	"github.com/sdwolfe32/anirip/crunchyroll"
	"os"
)

var (
	tempDir   = os.TempDir() + string(os.PathSeparator) + "anirip"
	seasonMap = map[int]string{
		0:  "Specials",
		1:  "Season One",
		2:  "Season Two",
		3:  "Season Three",
		4:  "Season Four",
		5:  "Season Five",
		6:  "Season Six",
		7:  "Season Seven",
		8:  "Season Eight",
		9:  "Season Nine",
		10: "Season Ten",
	}
)

func main() {
	l := anirip.NewLogger() // Generate a new logger
	l.Cyan("v1.5.0(6/14/2017) - by Steven Wolfe <steven@swolfe.me>")
	args := os.Args

	// If the user isn't using the cli correctly give them an example of how
	if len(os.Args) != 4 {
		l.Warn("CLI usage : anirip username password http://www.crunchyroll.com/miss-kobayashis-dragon-maid")
		return
	}
	download(l, args[3], args[1], args[2], "1080", "eng")
}

func download(l *anirip.Logger, showURL, user, pass, quality, subLang string) {
	// Verifies the existance of an anirip folder in our temp directory
	_, err := os.Stat(tempDir)
	if err != nil {
		l.Info("Generating new temporary directory")
		os.Mkdir(tempDir, 0777)
	}

	// Generate the HTTP client that will be used through whole lifecycle
	client, err := anirip.NewHTTPClient()
	if err != nil {
		l.Error(err)
		return
	}

	// Logs the user in and stores their session data in the clients jar
	l.Info("Logging into Crunchyroll as %s...	", user)
	if err = crunchyroll.Login(client, user, pass); err != nil {
		l.Error(err)
		return
	}

	// Scrapes all show metadata for the show requested
	var show anirip.Show
	show = new(crunchyroll.Show)
	l.Info("Scraping show metadata for %s", user)
	if err = show.Scrape(client, showURL); err != nil {
		l.Error(err)
		return
	}

	// Creates a new video processor that wil perform all video processing operations
	vp := anirip.NewVideoProcessor(tempDir)

	// Creates a new show directory which will store all seasons
	os.Mkdir(show.GetTitle(), 0777)
	for _, season := range show.GetSeasons() {

		// Creates a new season directory that will store all episodes
		os.Mkdir(show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()], 0777)
		for _, episode := range season.GetEpisodes() {

			// Retrieves more fine grained episode metadata
			l.Info("Retrieving Episode Info...")
			if err = episode.GetEpisodeInfo(client, "1080"); err != nil {
				l.Error(err)
				continue
			}

			// Checks to see if the episode already exists, in which case we continue to the next
			_, err = os.Stat(show.GetTitle() + string(os.PathSeparator) + seasonMap[season.GetNumber()] +
				string(os.PathSeparator) + episode.GetFilename() + ".mkv")
			if err == nil {
				l.Success("%s.mkv has already been downloaded successfully!", episode.GetFilename())
				continue
			}

			l.Cyan("Downloading %s", episode.GetFilename())

			// Downloads full MKV video from stream provider
			l.Info("Downloading video...")
			if err = episode.Download(vp); err != nil {
				l.Error(err)
				continue
			}

			// Downloads the subtitles to .ass format
			l.Info("Downloading subtitles...")
			subLang, err = episode.DownloadSubtitles(client, subLang, tempDir)
			if err != nil {
				l.Error(err)
				continue
			}

			// Attempts to merge the downloaded subtitles into the video stream
			l.Info("Merging subtitles into MKV container...")
			if err := vp.MergeSubtitles("jpn", subLang); err != nil {
				l.Error(err)
				continue
			}

			// Cleans the MKVs metadata for better reading by clients
			l.Info("Cleaning MKV...")
			if err := vp.CleanMKV(); err != nil {
				l.Error(err)
				continue
			}

			// Moves the episode to the appropriate season sub-directory
			if err := anirip.Rename(tempDir+string(os.PathSeparator)+"episode.mkv",
				show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()]+
					string(os.PathSeparator)+episode.GetFilename()+".mkv", 10); err != nil {
				l.Error(err)
			}
			l.Success("Downloading and merging completed successfully!")
		}
	}
	l.Cyan("Completed downloading episodes form %s", show.GetTitle())
	l.Info("Cleaning up temporary directory...")
	os.RemoveAll(tempDir)
}

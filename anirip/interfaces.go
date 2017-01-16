package anirip

import (
	"net/http"
	"os"
	"strconv"

	"github.com/fatih/color"
)

type Session interface {
	Login(string, string, string) error
	GetCookies() []*http.Cookie
}

type Show interface {
	Scrape(string, []*http.Cookie) error
	GetTitle() string
	GetSeasons() Seasons
}

type Seasons []Season

type Season interface {
	GetNumber() int
	GetEpisodes() Episodes
}

type Episodes []Episode

type Episode interface {
	GetEpisodeInfo(string, []*http.Cookie) error
	DownloadEpisode(string, string, []*http.Cookie) error
	DownloadSubtitles(string, int, string, []*http.Cookie) (string, error)
	GetFileName() string
}

func Download(show *Show) error {
	seasonMap := map[int]string{
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

	// Creates a new show directory which will store all seasons
	os.Mkdir(show.GetTitle(), 0777)
	for _, season := range show.GetSeasons() {

		// Creates a new season directory that will store all episodes
		os.Mkdir(show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()], 0777)
		for _, episode := range season.GetEpisodes() {

			color.White("[anirip] Retrieving Episode Info...\n")
			if err = episode.GetEpisodeInfo(quality, session.GetCookies()); err != nil {
				color.Red("[anirip] " + err.Error())
				continue
			}

			// Checks to see if the episode already exists, in which case we continue to the next
			_, err = os.Stat(show.GetTitle() + string(os.PathSeparator) + seasonMap[season.GetNumber()] + string(os.PathSeparator) + episode.GetFileName() + ".mkv")
			if err == nil {
				color.Green("[anirip] " + episode.GetFileName() + ".mkv has already been downloaded successfully..." + "\n")
				continue
			}

			subOffset := 0
			color.Cyan("[anirip] Downloading " + episode.GetFileName() + "\n")
			// Downloads full MKV video from stream provider
			color.White("[anirip] Downloading video...\n")
			if err := episode.DownloadEpisode(quality, tempDir, session.GetCookies()); err != nil {
				color.Red("[anirip] " + err.Error() + "\n")
				continue
			}

			// Trims down the downloaded MKV if the user wants to trim a Daisuki intro
			if daisukiIntroTrim {
				subOffset = subOffset + daisukiIntroLength
				color.White("[anirip] Trimming off Daisuki Intro - " + strconv.Itoa(daisukiIntroLength) + "ms\n")
				if err := trimMKV(daisukiIntroLength, tempDir); err != nil {
					color.Red("[anirip] " + err.Error() + "\n")
					continue
				}
			}

			// Trims down the downloaded MKV if the user wants to trim an Aniplex intro
			if aniplexIntroTrim {
				subOffset = subOffset + aniplexIntroLength
				color.White("[anirip] Trimming off Aniplex Intro - " + strconv.Itoa(aniplexIntroLength) + "ms\n")
				if err := trimMKV(aniplexIntroLength, tempDir); err != nil {
					color.Red("[anirip] " + err.Error() + "\n")
					continue
				}
			}

			// Trims down the downloaded MKV if the user wants to trim a Sunrise intro
			if sunriseIntroTrim {
				subOffset = subOffset + sunriseIntroLength
				color.White("[anirip] Trimming off Sunrise Intro - " + strconv.Itoa(sunriseIntroLength) + "ms\n")
				if err := trimMKV(sunriseIntroLength, tempDir); err != nil {
					color.Red("[anirip] " + err.Error() + "\n")
					continue
				}
			}

			// Downloads the subtitles to .ass format and
			// offsets their times by the passed provided interval
			color.White("[anirip] Downloading subtitles with a total offset of " + strconv.Itoa(subOffset) + "ms...\n")
			subtitleLang, err := episode.DownloadSubtitles(language, subOffset, tempDir, session.GetCookies())
			if err != nil {
				color.Red("[anirip] " + err.Error() + "\n")
				continue
			}

			// Attempts to merge the downloaded subtitles into the video strea
			color.White("[anirip] Merging subtitles into mkv container...\n")
			if err := mergeSubtitles("jpn", subtitleLang, tempDir); err != nil {
				color.Red("[anirip] " + err.Error() + "\n")
				continue
			}

			// Cleans the MKVs metadata for better reading by clients
			color.White("[anirip] Cleaning MKV...\n")
			if err := cleanMKV(tempDir); err != nil {
				color.Red("[anirip] " + err.Error() + "\n")
				continue
			}

			// Moves the episode to the appropriate season sub-directory
			if err := Rename(tempDir+string(os.PathSeparator)+"episode.mkv",
				show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()]+string(os.PathSeparator)+episode.GetFileName()+".mkv", 10); err != nil {
				color.Red(err.Error() + "\n\n")
			}
			color.Green("[anirip] Downloading and merging completed successfully.\n")
		}
	}
	color.Cyan("[anirip] Completed processing episodes for " + show.GetTitle() + "\n")
}

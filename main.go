package main /* import "s32x.com/anirip" */

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli"
	"s32x.com/anirip/common"
	"s32x.com/anirip/common/log"
	"s32x.com/anirip/crunchyroll"
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
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	app := cli.NewApp()
	app.Name = "anirip"
	app.Version = "1.5.2(12/8/2018)"
	app.Author = "Steven Wolfe"
	app.Email = "steven@swolfe.me"
	app.Usage = "anirip username password http://www.crunchyroll.com/miss-kobayashis-dragon-maid"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "language, l",
			Value: "en-US",
			Usage: "language code for the subtitles (not all are supported) ex: en-US, ja-JP",
		},
		cli.StringFlag{
			Name:  "quality, q",
			Value: "720",
			Usage: "quality of video to download ex: 1080, 720, 480, 360, android",
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Cyan("v%s - by %s <%s>", app.Version, app.Author, app.Email)
		args := c.Args()
		if len(args) != 3 {
			log.Warn("CLI Usage : " + app.Usage)
			return nil
		}

		download(args[2], args[0], args[1], c.String("quality"), c.String("language"))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func download(showURL, user, pass, quality, origsubLang string) {
	// Verifies the existence of an anirip folder in our temp directory
	_, err := os.Stat(tempDir)
	if err != nil {
		log.Info("Generating new temporary directory")
		os.Mkdir(tempDir, os.ModePerm)
	}

	// Generate the HTTP client that will be used through whole lifecycle
	client := common.NewHTTPClient()

	// Logs the user in and stores their session data in the clients jar
	log.Info("Logging into Crunchyroll...")
	if err = crunchyroll.Login(client, user, pass); err != nil {
		log.Error(err)
		return
	}

	// Scrapes all show metadata for the show requested
	var show common.Show
	show = new(crunchyroll.Show)
	log.Info("Scraping show metadata...")
	if err = show.Scrape(client, showURL); err != nil {
		log.Error(err)
		return
	}

	// Creates a new video processor that wil perform all video processing operations
	vp := common.NewVideoProcessor(tempDir)

	// Creates a new show directory which will store all seasons
	os.Mkdir(show.GetTitle(), os.ModePerm)
	for _, season := range show.GetSeasons() {

		// Creates a new season directory that will store all episodes
		os.Mkdir(show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()], os.ModePerm)
		for _, episode := range season.GetEpisodes() {

			// Retrieves more fine grained episode metadata
			log.Info("Retrieving Episode Info...")
			if err = episode.GetEpisodeInfo(client, quality); err != nil {
				log.Error(err)
				continue
			}

			// Checks to see if the episode already exists, in which case we continue to the next
			_, err = os.Stat(show.GetTitle() + string(os.PathSeparator) + seasonMap[season.GetNumber()] +
				string(os.PathSeparator) + episode.GetFilename() + ".mkv")
			if err == nil {
				log.Success("%s.mkv has already been downloaded successfully!", episode.GetFilename())
				continue
			}

			log.Cyan("Downloading %s", episode.GetFilename())

			// Downloads full MKV video from stream provider
			log.Info("Downloading video...")
			if err = episode.Download(vp); err != nil {
				log.Error(err)
				continue
			}

			subLang := origsubLang
			// Downloads the subtitles to .ass format
			log.Info("Downloading subtitles...%s", subLang)
			subLang, err = episode.DownloadSubtitles(client, subLang, tempDir)
			if err != nil {
				log.Error(err)
				continue
			}

			// Attempts to merge the downloaded subtitles into the video stream
			log.Info("Merging subtitles into MKV container...")
			if err := vp.MergeSubtitles("jpn", subLang); err != nil {
				log.Error(err)
				continue
			}

			// Moves the episode to the appropriate season sub-directory
			if err := common.Rename(tempDir+string(os.PathSeparator)+"episode.mkv",
				show.GetTitle()+string(os.PathSeparator)+seasonMap[season.GetNumber()]+
					string(os.PathSeparator)+episode.GetFilename()+".mkv", 10); err != nil {
				log.Error(err)
			}
			log.Success("Downloading and merging completed successfully!")
		}
	}
	log.Cyan("Completed downloading episodes from %s", show.GetTitle())
	log.Info("Cleaning up temporary directory...")
	os.RemoveAll(tempDir)
}

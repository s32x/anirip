package daisuki

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/ANIRip/anirip"
)

type DaisukiShow struct {
	Title   string
	AdID    string
	Path    string
	URL     string
	Seasons []DaisukiSeason
}

type DaisukiSeason struct {
	Title    string
	Number   int
	Length   int
	Episodes []DaisukiEpisode
}

type DaisukiEpisode struct {
	ID           int
	SubtitleID   int
	Title        string
	Description  string
	Number       float64
	Quality      string
	Path         string
	URL          string
	FileName     string
	SubtitleInfo TTMLInfo
	MediaInfo    HDSInfo
}

type TTMLInfo struct {
	TTMLUrl string
}

type HDSInfo struct {
	ManifestURL string
}

// Given a show pointer, appends all the seasons/episodes found for the show
func (show *DaisukiShow) ScrapeEpisodes(showURL string, cookies []*http.Cookie) error {
	// Gets the HTML of the show page
	showResponse, err := anirip.GetHTTPResponse("GET",
		showURL,
		nil,
		nil,
		cookies)
	if err != nil {
		return err
	}

	// Creates a goquery document for scraping episodes
	showDoc, err := goquery.NewDocumentFromResponse(showResponse)
	if err != nil {
		return anirip.Error{Message: "There was an error while accessing the show page", Err: err}
	}

	// Sets Title, AdID and Path in the case where the user passes a show URL
	if show.Title == "" {
		show.Title = showDoc.Find("h1#animeTitle").First().Text() // Scrapes the show title fromt he season body if it wasn't set by a search
	}
	if show.URL == "" {
		show.URL = showURL
	}
	if show.Path == "" {
		show.Path = strings.Replace(show.URL, "http://www.daisuki.net", "", 1) // Removes the host so we have just the path
	}
	if show.AdID == "" {
		show.AdID = strings.Replace(show.URL, "http://www.daisuki.net/us/en/anime/detail.", "", 1) // Removes the leading path
		show.AdID = strings.Replace(show.AdID, ".html", "", 1)                                     // Replaces the .html so we have just the AdID
	}

	// Searches first for the episodes/movies
	episodeMap := make(map[int]string)
	showDoc.Find("div#moviesBlock").Each(func(i int, season *goquery.Selection) {
		// Finds the non-movie/latest-or-first episode and adds it to our map
		season.Find("div#content0.content.clearFix.liquid").Each(func(i2 int, movieEpisode *goquery.Selection) {
			// Gets the episode number from that movie
			episodeThumb, exists := movieEpisode.Find("img").First().Attr("delay")
			episodeNumber, err := strconv.Atoi(movieEpisode.Find("p.episodeNumber").Text())
			if err == nil && exists && episodeThumb != "" {
				episodeMap[episodeNumber] = "/us/en/anime/watch." + strings.Split(episodeThumb, "/")[6] + "." + strings.Split(episodeThumb, "/")[7] + ".html"
			}
		})
		// Finds ALL non-movie/latest-or-first episodes
		season.Find("div#contentList0.contentList.clearFix.liquid div.item").Each(func(i2 int, episodeItem *goquery.Selection) {
			// Grabs episode information that isn't empty and has a url associated with an episode number
			episodeThumb, exists := episodeItem.Find("img").First().Attr("delay")
			episodeNumber, err := strconv.Atoi(episodeItem.Find("p.episodeNumber").Text())
			if err == nil && exists && episodeThumb != "" {
				episodeMap[episodeNumber] = "/us/en/anime/watch." + strings.Split(episodeThumb, "/")[6] + "." + strings.Split(episodeThumb, "/")[7] + ".html"
			}
		})
	})

	// Re-arranges seasons and episodes in the shows object so we have first to last
	show.Seasons = append(show.Seasons, DaisukiSeason{ // appends a new season that we'll append episodes to
		Title:  show.Title,
		Number: 1,
		Length: len(episodeMap),
	})
	for i := 0; i < len(episodeMap); i++ {
		show.Seasons[0].Episodes = append(show.Seasons[0].Episodes, DaisukiEpisode{
			Number: float64(i + 1),
			Path:   episodeMap[i+1],
			URL:    "http://www.daisuki.net" + episodeMap[i+1],
		})
	}

	// Assigns each season a number and episode a filename
	for s, season := range show.Seasons {
		for e, episode := range season.Episodes {
			// Generates a partial file name that we'll later improve on when we get the episode html
			show.Seasons[s].Episodes[e].FileName = anirip.GenerateEpisodeFileName(show.Title, show.Seasons[s].Number, episode.Number, "")
		}
	}
	return nil
}

// Gets the title of the show for referencing outside of this lib
func (show *DaisukiShow) GetTitle() string {
	return show.Title
}

// Re-stores seasons belonging to the show and returns them for iteration
func (show *DaisukiShow) GetSeasons() anirip.Seasons {
	seasons := []anirip.Season{}
	for i := 0; i < len(show.Seasons); i++ {
		seasons = append(seasons, &show.Seasons[i])
	}
	return seasons
}

// Re-stores episodes belonging to the season and returns them for iteration
func (season *DaisukiSeason) GetEpisodes() anirip.Episodes {
	episodes := []anirip.Episode{}
	for i := 0; i < len(season.Episodes); i++ {
		episodes = append(episodes, &season.Episodes[i])
	}
	return episodes
}

// Return the season number that will be used for folder naming
func (season *DaisukiSeason) GetNumber() int {
	return season.Number
}

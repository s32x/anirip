package crunchyroll

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/ANIRip/anirip"
)

type CrunchyrollShow struct {
	Title   string
	AdID    string
	Path    string
	URL     string
	Seasons []CrunchyrollSeason
}

type CrunchyrollSeason struct {
	Title    string
	Number   int
	Length   int
	Episodes []CrunchyrollEpisode
}

type CrunchyrollEpisode struct {
	ID          int
	SubtitleID  int
	Title       string
	Description string
	Number      float64
	Quality     string
	Path        string
	URL         string
	FileName    string
	MediaInfo   RTMPInfo
}

type RTMPInfo struct {
	URLOne string
	URLTwo string
	File   string
}

type ShowMetaData struct {
	Class          string `json:"class"`
	SeriesID       string `json:"series_id"`
	URL            string `json:"url"`
	Name           string `json:"name"`
	MediaType      string `json:"media_type"`
	LandscapeImage string `json:"landscape_image"`
	PortraitImage  string `json:"portrait_image"`
	Description    string `json:"description"`
}

// Given a show pointer, appends all the seasons/episodes found for the show
func (show *CrunchyrollShow) ScrapeEpisodes(showURL string, cookies []*http.Cookie) error {
	// Gets the HTML of the show page
	showResponse, err := anirip.GetHTTPResponse("GET",
		showURL,
		nil,
		http.Header{},
		cookies)
	if err != nil {
		return err
	}

	// Creates a goquery document for scraping
	showDoc, err := goquery.NewDocumentFromResponse(showResponse)
	if err != nil {
		return anirip.Error{Message: "There was an error while accessing the show page", Err: err}
	}

	// Scrapes the show metadata from the show page
	showMetaDataJSON := showDoc.Find("script#liftigniter-metadata").First().Text()

	// Parses the metadata json to a ShowMetaData object
	showMetaData := new(ShowMetaData)
	if err := json.Unmarshal([]byte(showMetaDataJSON), showMetaData); err != nil {
		return anirip.Error{Message: "There was an error while parsing show metadata", Err: err}
	}

	// Sets Title, and Path and URL on our show object
	show.Title = showMetaData.Name
	show.URL = showMetaData.URL
	show.Path = strings.Replace(show.URL, "http://www.crunchyroll.com", "", 1) // Removes the host so we have just the path

	// Searches first for the search div
	showDoc.Find("ul.list-of-seasons.cf").Each(func(i int, seasonList *goquery.Selection) {
		seasonList.Find("li.season").Each(func(i2 int, episodeList *goquery.Selection) {
			// Adds a new season to the show containing all information
			seasonTitle, _ := episodeList.Find("a").First().Attr("title")

			// Adds the title minus any "Episode XX" for shows that only have one season
			show.Seasons = append(show.Seasons, CrunchyrollSeason{
				Title: strings.SplitN(seasonTitle, " Episode ", 2)[0],
			})

			// Within that season finds all episodes
			episodeList.Find("div.wrapper.container-shadow.hover-classes").Each(func(i3 int, episode *goquery.Selection) {
				// Appends all new episode information to newly appended season
				episodeTitle := strings.TrimSpace(strings.Replace(episode.Find("span.series-title.block.ellipsis").First().Text(), "\n", "", 1))
				episodeNumber, _ := strconv.ParseFloat(strings.Replace(episodeTitle, "Episode ", "", 1), 64)
				episodePath, _ := episode.Find("a").First().Attr("href")
				episodeID, _ := strconv.Atoi(episodePath[len(episodePath)-6:])
				show.Seasons[i2].Episodes = append(show.Seasons[i2].Episodes, CrunchyrollEpisode{
					ID:     episodeID,
					Title:  episodeTitle,
					Number: episodeNumber,
					Path:   episodePath,
					URL:    "http://www.crunchyroll.com" + episodePath,
				})
			})
		})
	})

	// Re-arranges seasons and episodes in the shows object so we have first to last
	tempSeasonArray := []CrunchyrollSeason{}
	for i := len(show.Seasons) - 1; i >= 0; i-- {
		// First sort episodes from first to last
		tempEpisodesArray := []CrunchyrollEpisode{}
		for n := len(show.Seasons[i].Episodes) - 1; n >= 0; n-- {
			tempEpisodesArray = append(tempEpisodesArray, show.Seasons[i].Episodes[n])
		}
		// Lets not bother appending anything if there are no episodes in the season
		if len(tempEpisodesArray) > 0 {
			tempSeasonArray = append(tempSeasonArray, CrunchyrollSeason{
				Title:    show.Seasons[i].Title,
				Length:   len(tempEpisodesArray),
				Episodes: tempEpisodesArray,
			})
		}
	}
	show.Seasons = tempSeasonArray

	// Assigns each season a number and episode a filename
	for s, season := range show.Seasons {
		show.Seasons[s].Number = s + 1
		for e, episode := range season.Episodes {
			show.Seasons[s].Episodes[e].FileName = anirip.GenerateEpisodeFileName(show.Title, show.Seasons[s].Number, episode.Number, "")
		}
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)
	return nil
}

// Gets the title of the show for referencing outside of this lib
func (show *CrunchyrollShow) GetTitle() string {
	return show.Title
}

// Re-stores seasons belonging to the show and returns them for iteration
func (show *CrunchyrollShow) GetSeasons() anirip.Seasons {
	seasons := []anirip.Season{}
	for i := 0; i < len(show.Seasons); i++ {
		seasons = append(seasons, &show.Seasons[i])
	}
	return seasons
}

// Re-stores episodes belonging to the season and returns them for iteration
func (season *CrunchyrollSeason) GetEpisodes() anirip.Episodes {
	episodes := []anirip.Episode{}
	for i := 0; i < len(season.Episodes); i++ {
		episodes = append(episodes, &season.Episodes[i])
	}
	return episodes
}

// Return the season number that will be used for folder naming
func (season *CrunchyrollSeason) GetNumber() int {
	return season.Number
}

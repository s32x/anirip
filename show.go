package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Show struct {
	Title   string
	AdID    string
	Path    string
	URL     string
	Seasons []Season
}

type Season struct {
	Title    string
	Number   int
	Length   int
	Episodes []Episode
}

type Episode struct {
	ID          int
	SubtitleID  int
	Title       string
	Description string
	Number      float64
	Quality     string
	Path        string
	URL         string
	FileName    string
}

// Takes the passed show name and es crunchyroll,
// taking the first showname found as the show
func (show *Show) FindShow(searchTerm string, cookies []*http.Cookie) error {

	// Reforms showName string to url param
	encodedSearchTerm := strings.ToLower(strings.Replace(searchTerm, " ", "+", -1))

	// Gets the html of the search page
	searchReq, err := http.NewRequest("GET", "http://www.crunchyroll.com/search?from=&q="+encodedSearchTerm, nil)
	if err != nil {
		return Error{"There was an error creating episodes request", err}
	}
	searchReq.Header.Add("referer", "https://www.crunchyroll.com/")
	searchReq.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	for c := range cookies {
		searchReq.AddCookie(cookies[c])
	}
	searchResp, err := http.DefaultTransport.RoundTrip(searchReq)
	if err != nil {
		return Error{"There was an error performing show search", err}
	}

	// Gets the html of the search page we're looking for
	searchDoc, err := goquery.NewDocumentFromResponse(searchResp)
	if err != nil {
		return Error{"There was an error searching for show", err}
	}

	// Searches first for the search div
	firstSeriesTitle := ""
	firstEpisodeURL := ""
	searchDoc.Find("div#aux_results").Each(func(i int, s *goquery.Selection) {
		firstSeriesTitle = s.Find("span.series").First().Text()
		firstEpisodeURL, _ = s.Find("a").First().Attr("href")
	})
	if firstSeriesTitle == "" || firstEpisodeURL == "" {
		return Error{"There was an issue while getting the first search result", nil}
	}

	// Gets the first result from our parse search and returns the path if its not ""/store/" or "/crunchygay/"
	firstPath := strings.Replace(firstEpisodeURL, "http://www.crunchyroll.com/", "", 1)
	firstShowPath := strings.Split(firstPath, "/")[0]               // Gets only the first path name (ideally a show name)
	if firstShowPath == "store" || firstShowPath == "crunchycast" { // tf is a crunchycast?
		return Error{"Recieved a non-show store/crunchycast result", nil}
	}

	// Sets the title, path, and url of the first result
	show.Title = firstSeriesTitle
	show.Path = firstShowPath
	show.URL = "http://www.crunchyroll.com/" + firstShowPath
	return nil
}

// Given a show pointer, appends all the seasons/episodes found for the show
func (show *Show) GetEpisodes(cookies []*http.Cookie) error {
	// Gets the html of the show page
	episodesReq, err := http.NewRequest("GET", show.URL, nil)
	if err != nil {
		return Error{"There was an error creating episodes request", err}
	}
	episodesReq.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
	for c := range cookies {
		episodesReq.AddCookie(cookies[c])
	}
	episodesResp, err := http.DefaultTransport.RoundTrip(episodesReq)
	if err != nil {
		return Error{"There was an error performing login request", err}
	}
	showDoc, err := goquery.NewDocumentFromResponse(episodesResp)
	if err != nil {
		return Error{"There was an error while accessing the show page", err}
	}

	// Searches first for the search div
	showDoc.Find("ul.list-of-seasons.cf").Each(func(i int, seasonList *goquery.Selection) {
		seasonList.Find("li.season").Each(func(i2 int, episodeList *goquery.Selection) {
			// Adds a new season to the show containing all information
			seasonTitle, _ := episodeList.Find("a").First().Attr("title")

			// Adds the title minus any "Episode XX" for shows that only have one season
			show.Seasons = append(show.Seasons, Season{
				Title: strings.SplitN(seasonTitle, " Episode ", 2)[0],
			})

			// Within that season finds all episodes
			episodeList.Find("div.wrapper.container-shadow.hover-classes").Each(func(i3 int, episode *goquery.Selection) {
				// Appends all new episode information to newly appended season
				episodeTitle := strings.TrimSpace(strings.Replace(episode.Find("span.series-title.block.ellipsis").First().Text(), "\n", "", 1))
				episodeDescription := strings.TrimSpace(episode.Find("p.short-desc").First().Text())
				episodeNumber, _ := strconv.ParseFloat(strings.Replace(episodeTitle, "Episode ", "", 1), 64)
				episodePath, _ := episode.Find("a").First().Attr("href")
				episodeID, _ := strconv.Atoi(episodePath[len(episodePath)-6:])
				show.Seasons[i2].Episodes = append(show.Seasons[i2].Episodes, Episode{
					ID:          episodeID,
					Title:       episodeTitle,
					Description: episodeDescription,
					Number:      episodeNumber,
					Path:        episodePath,
					URL:         "http://www.crunchyroll.com" + episodePath,
				})
			})
		})
	})

	// Re-arranges seasons and episodes in the shows object so we have first to last
	tempSeasonArray := []Season{}
	for i := len(show.Seasons) - 1; i >= 0; i-- {
		// First sort episodes from first to last
		tempEpisodesArray := []Episode{}
		for n := len(show.Seasons[i].Episodes) - 1; n >= 0; n-- {
			tempEpisodesArray = append(tempEpisodesArray, show.Seasons[i].Episodes[n])
		}
		// Lets not bother appending anything if there are no episodes in the season
		if len(tempEpisodesArray) > 0 {
			tempSeasonArray = append(tempSeasonArray, Season{
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
			show.Seasons[s].Episodes[e].FileName = generateEpisodeFileName(show.Title, show.Seasons[s].Number, episode.Number, episode.Description)
		}
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)
	return nil
}

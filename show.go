package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Show struct {
	Title   string
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
	Title       string
	Description string
	Number      float64
	URL         string
}

// Takes the passed show name and searches crunchyroll,
// taking the first showname found as the show
func searchShowPath(showName string) (Show, error) {

	// Reforms showName string to url param
	encodedShowName := strings.ToLower(strings.Replace(showName, " ", "+", -1))

	// Attempts to grab the contents of the search URL
	fmt.Println("\nAttempting to determine the URL for passed show : " + showName)
	fmt.Println(">> Accessing search URL : " + searchURL + encodedShowName)

	// Gets the html of the search page we're looking for
	searchDoc, err := goquery.NewDocument(searchURL + encodedShowName)
	if err != nil {
		return Show{}, err
	}

	// Searches first for the search div
	firstSeriesTitle := ""
	firstEpisodeURL := ""
	searchDoc.Find("div#aux_results").Each(func(i int, s *goquery.Selection) {
		firstSeriesTitle = s.Find("span.series").First().Text()
		firstEpisodeURL, _ = s.Find("a").First().Attr("href")
	})
	if firstSeriesTitle == "" || firstEpisodeURL == "" {
		fmt.Println(">>> There was an issue while getting the first search result.")
		return Show{}, nil
	}

	// Gets the first result from our parse search and returns the path if its not ""/store/" or "/crunchygay/"
	firstPath := strings.Replace(firstEpisodeURL, hostURL+"/", "", 1)
	firstShowPath := strings.Split(firstPath, "/")[0]               // Gets only the first path name (ideally a show name)
	if firstShowPath == "store" || firstShowPath == "crunchycast" { // tf is a crunchycast?
		return Show{}, nil
	}

	// Packs up all assumed show information and returns it
	return Show{
		Title: firstSeriesTitle, // Series name recieved from javascript json
		Path:  firstShowPath,    // Show path retrieved from a href url
		URL:   hostURL + "/" + firstShowPath,
	}, nil
}

// Given a show object, returns an array of urls of
// all the episodes associated with that show
func getEpisodes(show Show) (Show, error) {
	// Attempts to grab the contents of the show page
	fmt.Println("\n>> Accessing show page URL : " + show.URL)

	// Gets the html of the show page we previously got
	showDoc, err := goquery.NewDocument(show.URL)
	if err != nil {
		return Show{}, err
	}

	// Searches first for the search div
	showDoc.Find("ul.list-of-seasons.cf").Each(func(i int, seasonList *goquery.Selection) {
		seasonList.Find("li.season").Each(func(i2 int, episodeList *goquery.Selection) {
			// Adds a new season to the show containing all information
			seasonTitle, _ := episodeList.Find("a").First().Attr("title")
			show.Seasons = append(show.Seasons, Season{
				// Adds the title minus any "Episode XX" for shows that only have one season
				Title: strings.SplitN(seasonTitle, " Episode ", 2)[0],
			})
			episodeList.Find("div.wrapper.container-shadow.hover-classes").Each(func(i3 int, episode *goquery.Selection) {
				// Appends all new episode information to newly appended season
				episodeTitle := strings.TrimSpace(strings.Replace(episode.Find("span.series-title.block.ellipsis").First().Text(), "\n", "", 1))
				episodeDescription := strings.TrimSpace(episode.Find("p.short-desc").First().Text())
				episodeNumber, _ := strconv.ParseFloat(strings.Replace(episodeTitle, "Episode ", "", 1), 64)
				episodeURL, _ := episode.Find("a").First().Attr("href")
				episodeID, _ := strconv.Atoi(episodeURL[len(episodeURL)-6:])
				show.Seasons[i2].Episodes = append(show.Seasons[i2].Episodes, Episode{
					ID:          episodeID,
					Title:       episodeTitle,
					Description: episodeDescription,
					Number:      episodeNumber,
					URL:         hostURL + episodeURL,
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

	// Lastly assigns each season a number based on their order in our sorted array
	for i := 0; i < len(show.Seasons); i++ {
		show.Seasons[i].Number = i + 1
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)

	// Prints and returns the array of episodes that we we're able to get
	fmt.Println("Discovered a total of " + strconv.Itoa(len(show.Seasons)) + " seasons...")
	return show, nil
}

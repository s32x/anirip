package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Show struct {
	Title      string
	Path       string
	URL        string
	Seasons    []Season
	SearchTerm string
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
	Path        string
	URL         string
	FileName    string
	Subtitle    Subtitle
	RTMPInfo    RTMPInfo
}

func getShow(show *Show) error {
	// First asks the user for the show they are interested in downloading
	getStandardUserInput("Enter a show name: ", &show.SearchTerm)

	// Next we search for the showname/path of the show we would like to download
	err := searchShowPath(show)
	if err != nil || show.URL == "" {
		fmt.Printf(err.Error() + "\n\n")
		return err
	}
	fmt.Println("\nDetermined a valid show name of : --- " + show.Title + " ---")

	// Populates all episodes for the given show
	err = populateEpisodes(show)
	if err != nil || len(show.Seasons) == 0 || len(show.Seasons[0].Episodes) == 0 {
		fmt.Printf(err.Error() + "\n\n")
		return err
	}

	// Attempts to access and print the titles of all seasons recieved
	fmt.Printf("Below is a list of seasons found ...\n\n")
	for i := 0; i < len(show.Seasons); i++ {
		fmt.Printf("\tSeason " + strconv.Itoa(show.Seasons[i].Number) + " - " + show.Seasons[i].Title + " (" + strconv.Itoa(show.Seasons[i].Length) + " Episodes)\n")
	}

	return nil
}

// Takes the passed show name and es crunchyroll,
// taking the first showname found as the show
func searchShowPath(show *Show) error {

	// Reforms showName string to url param
	encodedSearchTerm := strings.ToLower(strings.Replace(show.SearchTerm, " ", "+", -1))

	// Attempts to grab the contents of the search URL
	fmt.Println("\nAttempting to determine the URL for passed show : " + show.SearchTerm)
	fmt.Println(">> Accessing search URL : " + "http://www.crunchyroll.com/search?from=&q=" + encodedSearchTerm)

	// Gets the html of the search page we're looking for
	searchDoc, err := goquery.NewDocument("http://www.crunchyroll.com/search?from=&q=" + encodedSearchTerm)
	if err != nil {
		return CRError{"There was an error searching for show", err}
	}

	// Searches first for the search div
	firstSeriesTitle := ""
	firstEpisodeURL := ""
	searchDoc.Find("div#aux_results").Each(func(i int, s *goquery.Selection) {
		firstSeriesTitle = s.Find("span.series").First().Text()
		firstEpisodeURL, _ = s.Find("a").First().Attr("href")
	})
	if firstSeriesTitle == "" || firstEpisodeURL == "" {
		return CRError{"There was an issue while getting the first search result", nil}
	}

	// Gets the first result from our parse search and returns the path if its not ""/store/" or "/crunchygay/"
	firstPath := strings.Replace(firstEpisodeURL, "http://www.crunchyroll.com/", "", 1)
	firstShowPath := strings.Split(firstPath, "/")[0]               // Gets only the first path name (ideally a show name)
	if firstShowPath == "store" || firstShowPath == "crunchycast" { // tf is a crunchycast?
		return CRError{"Recieved a non-show store/crunchycast result", nil}
	}

	// Sets the title, path, and url of the first result
	show.Title = firstSeriesTitle
	show.Path = firstShowPath
	show.URL = "http://www.crunchyroll.com/" + firstShowPath

	return nil
}

// Given a show pointer, appends all the seasons/episodes found for the show
func populateEpisodes(show *Show) error {
	// Attempts to grab the contents of the show page
	fmt.Println("\n>> Accessing show page URL : " + show.URL)

	// Gets the html of the show page we previously got
	showDoc, err := goquery.NewDocument(show.URL)
	if err != nil {
		return CRError{"There was an error while accessing the show page", err}
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

	// Lastly assigns each season a number based on their order in our sorted array
	for i := 0; i < len(show.Seasons); i++ {
		show.Seasons[i].Number = i + 1
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)

	// Prints and returns the array of episodes that we we're able to get
	fmt.Println("Discovered a total of " + strconv.Itoa(len(show.Seasons)) + " seasons...")
	return nil
}

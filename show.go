package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
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
	Title  string
	Number int
	URL    string
}

// Takes the passed show name and searches crunchyroll,
// taking the first showname found as the show
func searchShowPath(showName string) (Show, error) {

	// Reforms showName string to url param
	encodedShowName := strings.ToLower(strings.Replace(showName, " ", "+", -1))

	// Attempts to grab the contents of the search URL
	fmt.Println("\nAttempting to determine the URL for passed show : " + showName)
	fmt.Println(">> Accessing search URL : " + searchURL + encodedShowName)
	resp, err := http.Get(searchURL + encodedShowName)
	if err != nil {
		return Show{}, err
	}
	defer resp.Body.Close()

	// Breaks HTML string into a reader that we can search for HTML element nodes
	searchBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Show{}, err
	}
	doc, err := html.Parse(strings.NewReader(string(searchBytes)))
	if err != nil {
		return Show{}, err
	}
	var findFirstShow func(*html.Node)

	// Handles searching for the url of the show by the first <a href="http://www.crunchyroll.com..."
	URLArray := []string{}
	seriesTitle := ""
	titleFlag := false

	findFirstShow = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if strings.Contains(a.Val, hostURL) {
						URLArray = append(URLArray, a.Val)
						break
					}
				}
			}
		}
		// Activator for title reading after span class="series"
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, span := range n.Attr {
				if span.Key == "class" && span.Val == "series" {
					titleFlag = true // power up the bass cannon
					break
				}
				if span.Key == "class" && span.Val == "name" {
					titleFlag = false // power down the bass cannon
					break
				}
			}
		}
		// Wait for titleFlag to be tripped by class="series" before setting series title
		if n.Type == html.TextNode && n.Data != "" && titleFlag && seriesTitle == "" {
			seriesTitle = n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findFirstShow(c)
		}
	}
	findFirstShow(doc)

	// Gets the first result from our parse search and returns the path if its not ""/store/" or "/crunchygay/"
	firstPath := strings.Replace(URLArray[0], hostURL+"/", "", 1)
	firstShowPath := strings.Split(firstPath, "/")[0]               // Gets only the first path name (ideally a show name)
	if firstShowPath == "store" || firstShowPath == "crunchycast" { // tf is a crunchycast?
		return Show{}, nil
	}

	// Packs up all assumed show information and returns it
	return Show{
		Title: seriesTitle,   // Series name recieved from javascript json
		Path:  firstShowPath, // Show path retrieved from a href url
		URL:   hostURL + "/" + firstShowPath,
	}, nil
}

// Given a show name, returns an array of urls of
// all the episodes associated with that show
func getEpisodes(show Show) (Show, error) {
	// Attempts to grab the contents of the show page
	fmt.Println("\n>> Accessing show page URL : " + show.URL)
	resp, err := http.Get(show.URL)
	if err != nil {
		return show, err
	}
	defer resp.Body.Close()

	// Parses the HTML in order to search for the list of episodes
	showBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return show, err
	}
	doc, err := html.Parse(strings.NewReader(string(showBytes)))
	if err != nil {
		return show, err
	}
	var findEpisodes func(*html.Node)

	seasonArray := []Season{Season{}}
	episodeArray := []Episode{}

	// Handles searching for the url of the show by the first <a href="http://www.crunchyroll.com..."
	findEpisodes = func(n *html.Node) {
		// If we see an li that resembles a new season list
		if n.Type == html.ElementNode && n.Data == "li" {
			for _, li := range n.Attr {
				if li.Key == "class" && strings.Contains(li.Val, "season ") {
					// Appends a new season to the seasonArray which we will add episodes to
					seasonArray[len(seasonArray)-1].Episodes = episodeArray

					// Clears out the episodeArray and sets up a new empty season for populating the new season
					seasonArray = append(seasonArray, Season{})
					episodeArray = []Episode{}
					break
				}
			}
		}
		// If we see an 'a' that represents a new episode
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "title" && seasonArray[len(seasonArray)-1].Title == "" {
					// Stores the series title so we can use it on the season title
					seasonArray[len(seasonArray)-1].Title = a.Val
				}
				if a.Key == "href" && strings.Contains(a.Val, "/episode-") {
					// Pulls all the episode information from the URL in order to populate a new episode object
					episodePathArray := strings.Split(strings.Split(a.Val, "/")[2], "-")

					// Gets episode name from a.Val
					episodeName := ""
					for i := 2; i < len(episodePathArray)-1; i++ {
						if i != 2 {
							episodeName = episodeName + " "
						}
						episodeName = episodeName + episodePathArray[i]
					}
					episodeName = strings.Title(episodeName)

					// Gets episode number from a.Val
					episodeNumber, err := strconv.Atoi(episodePathArray[1])
					if err != nil {
						episodeNumber = 0
					}

					// Gets episode URL from a.Val
					episodeURL := hostURL + a.Val
					episodeArray = append(episodeArray, Episode{
						Title:  episodeName,
						Number: episodeNumber,
						URL:    episodeURL,
					})
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findEpisodes(c)
		}
	}
	findEpisodes(doc)

	// Adds the final episode array (in most cases season 1) to the seasons array
	seasonArray[len(seasonArray)-1].Episodes = episodeArray

	// Re-arranges seasons array so that first season is first instead of last
	finalSeasonArray := []Season{}
	for i := len(seasonArray) - 1; i >= 0; i-- {
		if len(seasonArray[i].Episodes) > 0 {
			finalSeasonArray = append(finalSeasonArray, seasonArray[i])
		}
	}

	// Re-arranges episode in each season so we have first to last
	for i := 0; i < len(finalSeasonArray); i++ {
		finalEpisodeArray := []Episode{}
		for n := len(finalSeasonArray[i].Episodes) - 1; n >= 0; n-- {
			finalEpisodeArray = append(finalEpisodeArray, finalSeasonArray[i].Episodes[n])
		}
		finalSeasonArray[i].Episodes = finalEpisodeArray
		finalSeasonArray[i].Length = len(finalEpisodeArray)
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)

	// Prints and returns the array of episodes that we we're able to get
	show.Seasons = finalSeasonArray
	fmt.Println("Recieved a total of " + strconv.Itoa(len(show.Seasons)) + " seasons...")
	return show, nil
}

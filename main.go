package main

import (
  "fmt"
  "net/http"
  "golang.org/x/net/html"
  "strings"
  "io/ioutil"
  "strconv"
  "os"
  "bufio"
)

var (
  showSearchTerm = ""
  hostURL = "http://www.crunchyroll.com"
  searchURL = hostURL + "/search?from=&q="
)

func main() {
  // First we as the user for what show they would like to rip
  getStandardUserInput("Please enter a show name : ", &showSearchTerm)

  // First we get the showname/path of the show we would like to download
  showPath, err := searchShowPath(showSearchTerm)
  if err != nil || showPath == "" {
    fmt.Println("Unable to get a show name/path via search results. ", err)
    return
  }
  fmt.Println("Determined a valid show name of : --- " + strings.Title(strings.Replace(showPath, "-", " ", -1)) + " ---")

  // Next we attempt to get an array of episode URLS via the results page
  showURL := hostURL + "/" + showPath
  // Gets a 2-dimentional array of episode URLS
  episodeList, err := getEpisodeList(showURL)
  if err != nil || len(episodeList) == 0 {
    fmt.Println("Unable to get any episode URLS. ", err)
    return
  }
  fmt.Println("Recieved the following list of episodes : ", episodeList)

  //TODO RTMP Dumps each episode in a seperate goroutine...
}

// Gets user input from the user and unmarshalls it into the input
func getStandardUserInput(prefixText string, input *string) error {
  fmt.Printf(prefixText)
  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
  	*input = scanner.Text()
    break
  }
  if err := scanner.Err(); err != nil {
    return err
  }
  return nil
}

// Takes the passed show name and searches crunchyroll,
// taking the first showname found as the show
func searchShowPath(showName string) (string, error) {
  // Reforms showName string to url param
  encodedShowName := strings.ToLower(strings.Replace(showName, " ", "+", -1))

  // Attempts to grab the contents of the search URL
  fmt.Println("Attempting to determine the URL for passed show : " + showName)
  fmt.Println(">> Accessing search URL : " + searchURL + encodedShowName)
  resp, err := http.Get(searchURL + encodedShowName)
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

  // Breaks HTML string into a reader that we can search for HTML element nodes
  searchBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }
  doc, err := html.Parse(strings.NewReader(string(searchBytes)))
  if err != nil {
    return "", err
  }
  var findFirstShow func(*html.Node)

  // Handles searching for the url of the show by the first <a href="http://www.crunchyroll.com..."
  resultsArray := []string{}
  findFirstShow = func(n *html.Node) {
      if n.Type == html.ElementNode && n.Data == "a" {
          for _, a := range n.Attr {
              if a.Key == "href" {
                  if strings.Contains(a.Val, hostURL) {
                    resultsArray = append(resultsArray, a.Val)
                    break
                  }
              }
          }
      }
      for c := n.FirstChild; c != nil; c = c.NextSibling {
          findFirstShow(c)
      }
  }
  findFirstShow(doc)

  // Gets the first result from our parse search and returns the path if its not ""/store/""
  firstPath := strings.Replace(resultsArray[0], hostURL + "/", "", 1)
  firstShowPath := strings.Split(firstPath, "/")[0] // Gets only the first path name (ideally a show name)
  if firstShowPath == "store" || firstShowPath == "crunchycast"{ // tf is a crunchycast?
    return "", nil
  }
  return firstShowPath, nil
}

// Given a show name, returns an array of urls of
// all the episodes associated with that show
func getEpisodeList(showURL string) ([]string, error) {
  // Attempts to grab the contents of the show page
  fmt.Println(">> Accessing show page URL : " + showURL)
  resp, err := http.Get(showURL)
  if err != nil {
    return []string{}, err
  }
  defer resp.Body.Close()

  // Parses the HTML in order to search for the list of episodes
  showBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return []string{}, err
  }
  doc, err := html.Parse(strings.NewReader(string(showBytes)))
  if err != nil {
    return []string{}, err
  }
  var findEpisodes func(*html.Node)

  // Handles searching for the url of the show by the first <a href="http://www.crunchyroll.com..."
  episodeArray := []string{}
  episodeTotal := 0
  findEpisodes = func(n *html.Node) {
      // If we see an li that resembles a new season list
      if n.Type == html.ElementNode && n.Data == "li" {
        for _, li := range n.Attr {
          if li.Key == "class" && strings.Contains(li.Val, "season") {
            // Appends a "SEASONBREAK" string so we can separate seasons in the array
            episodeArray = append(episodeArray, "SEASONBREAK")
            break
          }
        }
      }
      // If we see an 'a' that represents a new episode
      if n.Type == html.ElementNode && n.Data == "a" {
        for _, a := range n.Attr {
          if a.Key == "href" && strings.Contains(a.Val, "/episode-"){
            episodeArray = append(episodeArray, hostURL + a.Val)
            episodeTotal = episodeTotal + 1
            break
          }
        }
      }
      for c := n.FirstChild; c != nil; c = c.NextSibling {
          findEpisodes(c)
      }
  }
  findEpisodes(doc)

  // Re-arranges array so that first episode is first instead of last
  finalEpisodeArray := []string{}
  for i := len(episodeArray) - 1; i >= 0; i-- {
    finalEpisodeArray = append(finalEpisodeArray, episodeArray[i])
  }

  // Prints and returns the array of episodes that we we're able to get
  fmt.Println("Returning a total of " + strconv.Itoa(episodeTotal) + " episodes.")
  return finalEpisodeArray, nil
}

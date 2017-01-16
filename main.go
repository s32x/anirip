package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/sdwolfe32/anirip/anirip"
	"github.com/sdwolfe32/anirip/crunchyroll"
	"gopkg.in/urfave/cli.v1"
)

var (
	tempDir = os.TempDir() + string(os.PathSeparator) + "anirip"
)

func main() {
	username := ""
	password := ""
	language := "English"
	quality := "1080p"

	app := cli.NewApp()
	app.Name = "anirip"
	app.Author = "Steven Wolfe"
	app.Email = "steven@swolfe.me"
	app.Version = "v1.5.0(1/15/2017)"
	app.Usage = "Crunchyroll show ripper CLI"
	color.Cyan(app.Name + " " + app.Version + " - by " + app.Author + " <" + app.Email + ">\n")
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "lang, l",
			Value:       "english",
			Usage:       "desired subtitle language",
			Destination: &language,
		},
		cli.StringFlag{
			Name:        "quality, q",
			Value:       "1080p",
			Usage:       "desired video quality",
			Destination: &quality,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "login",
			Aliases: []string{"l"},
			Usage:   "creates and stores cookies for a stream provider",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "user, u",
					Value:       "myusername",
					Usage:       "premium username used to access video stream",
					Destination: &username,
				},
				cli.StringFlag{
					Name:        "pass, p",
					Value:       "mypassword",
					Usage:       "premium password used to access video stream",
					Destination: &password,
				},
			},
			Action: func(c *cli.Context) error {
				// Creates session with cookies to store in file
				var session anirip.Session
				color.Cyan("[anirip] Logging to CrunchyRoll as " + username + "...")

				// Performs the login procedure, storing the login information to file
				session = new(crunchyroll.Session)
				if err := session.Login(username, password, tempDir); err != nil {
					color.Red("[anirip] " + err.Error())
					return anirip.Error{Message: "Unable to login to Crunchyroll", Err: err}
				}
				color.Green("[anirip] Successfully logged in... Cookies saved to " + tempDir)
				return nil
			},
		},
		{
			Name:    "clear",
			Aliases: []string{"c"},
			Usage:   "erases the temporary directory used for cookies and temp files",
			Action: func(c *cli.Context) error {
				// Attempts to erase the temporary directory
				if err := os.RemoveAll(tempDir); err != nil {
					color.Red("[anirip] There was an error erasing the temporary directory : " + err.Error())
					return anirip.Error{Message: "There was an error erasing the temporary directory", Err: err}
				}
				color.Green("[anirip] Successfully erased the temporary directory " + tempDir)
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.NArg() == 0 {
			color.Red("[anirip] No show URLs provided.")
			return anirip.Error{Message: "No show URLs provided"}
		}

		for _, showURL := range c.Args() {
			// Logs the user in
			session = new(crunchyroll.Session)
			if err = session.Login(username, password, tempDir); err != nil {
				color.Red("[anirip] " + err.Error())
				return anirip.Error{Message: "Unable to login to provider", Err: err}
			}

			// Attempts to scrape the shows metadata/information
			show = new(crunchyroll.Show)
			if err = show.Scrape(showURL, session.GetCookies()); err != nil {
				color.Red("[anirip] " + err.Error())
				return anirip.Error{Message: "Unable to get episodes", Err: err}
			}

			if err = anirip.Download(show); err != nil {
				color.Red("[anirip] " + err.Error())
				return anirip.Error{Message: "Unable to download show", Err: err}
			}
		}
		return nil
	}
	app.Run(os.Args)
}

func init() {
	// Verifies the existance of an anirip folder in our temp directory
	_, err := os.Stat(tempDir)
	if err != nil {
		os.Mkdir(tempDir, 0777)
	}
}

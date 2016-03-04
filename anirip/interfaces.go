package anirip

import "net/http"

type Session interface {
	Login(string, string) error
	GetCookies() []*http.Cookie
}

type Show interface {
	ScrapeEpisodes(string, []*http.Cookie) error
	GetTitle() string
	GetSeasons() Seasons
}

type Seasons []Season

type Season interface {
	GetEpisodes() Episodes
}

type Episodes []Episode

type Episode interface {
	DownloadEpisode(string, []*http.Cookie) error
	DownloadSubtitles(string, int, []*http.Cookie) error
	GetFileName() string
}

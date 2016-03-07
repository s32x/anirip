package anirip

import "net/http"

type Session interface {
	Login(string, string, string) error
	GetCookies() []*http.Cookie
}

type Show interface {
	ScrapeEpisodes(string, []*http.Cookie) error
	GetTitle() string
	GetSeasons() Seasons
}

type Seasons []Season

type Season interface {
	GetNumber() int
	GetEpisodes() Episodes
}

type Episodes []Episode

type Episode interface {
	GetEpisodeInfo(string, []*http.Cookie) error
	DownloadEpisode(string, string, string, []*http.Cookie) error
	DownloadSubtitles(string, int, string, []*http.Cookie) (string, error)
	GetFileName() string
}

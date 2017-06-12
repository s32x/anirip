package anirip

// TODO Consolidate interfaces into Downloader

// Session implements Session functionality
type Session interface {
	Login(HTTPClient, string, string, string) error
	GetCookies() HTTPClient
}

// Show implements Show functionality
type Show interface {
	Scrape(HTTPClient, string) error
	GetTitle() string
	GetSeasons() Seasons
}

// Seasons is an aliased slice of Seasons
type Seasons []Season

// Season implements Season functionality
type Season interface {
	GetNumber() int
	GetEpisodes() Episodes
}

// Episodes is an aliased slice of Episodes
type Episodes []Episode

// Episode implements Episode functionality
type Episode interface {
	GetEpisodeInfo(client HTTPClient, quality string) error
	DownloadEpisode(vp VideoProcessor, quality string) error
	// DownloadSubtitles(client HTTPClient, string, int, string) (string, error)
	GetFilename() string
}

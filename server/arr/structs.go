package arr

import "time"

type QualityProfile struct {
	Name           string `json:"name"`
	UpgradeAllowed bool   `json:"upgradeAllowed"`
	Cutoff         int    `json:"cutoff"`
	Items          []struct {
		Quality struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Resolution int    `json:"resolution"`
		} `json:"quality,omitempty"`
		Items   []any  `json:"items"`
		Allowed bool   `json:"allowed"`
		Name    string `json:"name,omitempty"`
		ID      int    `json:"id,omitempty"`
	} `json:"items"`
	ID int `json:"id"`
}

type RootFolder struct {
	Path            string `json:"path"`
	Accessible      bool   `json:"accessible"`
	FreeSpace       int64  `json:"freeSpace"`
	UnmappedFolders []any  `json:"unmappedFolders"`
	ID              int    `json:"id"`
}

type LanguageProfile struct {
	Name           string `json:"name"`
	UpgradeAllowed bool   `json:"upgradeAllowed"`
	Cutoff         struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"cutoff"`
	Languages []struct {
		Language struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"language"`
		Allowed bool `json:"allowed"`
	} `json:"languages"`
	ID int `json:"id"`
}

type ArrCommandRequest struct {
	Name string `json:"name"`
}

type CommandResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CommandName string `json:"commandName"`
	Message     string `json:"message"`
	Body        struct {
		SendUpdatesToClient bool      `json:"sendUpdatesToClient"`
		LastExecutionTime   time.Time `json:"lastExecutionTime"`
		LastStartTime       time.Time `json:"lastStartTime"`
		Trigger             string    `json:"trigger"`
		SuppressMessages    bool      `json:"suppressMessages"`
		ClientUserAgent     string    `json:"clientUserAgent"`
	} `json:"body"`
	Priority string    `json:"priority"`
	Status   string    `json:"status"`
	Result   string    `json:"result"`
	Queued   time.Time `json:"queued"`
	Started  time.Time `json:"started"`
	Ended    time.Time `json:"ended"`
	Duration struct {
		Ticks int `json:"ticks"`
	} `json:"duration"`
	Exception           string    `json:"exception"`
	Trigger             string    `json:"trigger"`
	ClientUserAgent     string    `json:"clientUserAgent"`
	StateChangeTime     time.Time `json:"stateChangeTime"`
	SendUpdatesToClient bool      `json:"sendUpdatesToClient"`
	UpdateScheduledTask bool      `json:"updateScheduledTask"`
	LastExecutionTime   time.Time `json:"lastExecutionTime"`
}

type QueueDetail struct {
	Languages []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"languages"`
	Quality struct {
		Quality struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Resolution int    `json:"resolution"`
			Modifier   string `json:"modifier"`
		} `json:"quality"`
		Revision struct {
			Version  int  `json:"version"`
			Real     int  `json:"real"`
			IsRepack bool `json:"isRepack"`
		} `json:"revision"`
	} `json:"quality"`
	Size                                float64       `json:"size"`
	Title                               string        `json:"title"`
	SizeLeft                            float64       `json:"sizeleft"`
	TimeLeft                            string        `json:"timeleft"`
	EstimatedCompletionTime             time.Time     `json:"estimatedCompletionTime"`
	Added                               time.Time     `json:"added"`
	Status                              string        `json:"status"`
	TrackedDownloadStatus               string        `json:"trackedDownloadStatus"`
	TrackedDownloadState                string        `json:"trackedDownloadState"`
	StatusMessages                      []interface{} `json:"statusMessages"`
	DownloadID                          string        `json:"downloadId"`
	Protocol                            string        `json:"protocol"`
	DownloadClient                      string        `json:"downloadClient"`
	DownloadClientHasPostImportCategory bool          `json:"downloadClientHasPostImportCategory"`
	Indexer                             string        `json:"indexer"`
	ID                                  int           `json:"id"`
}

type RadarrQueueDetails []struct {
	QueueDetail
	MovieID int `json:"movieId"`
}

type SonarrQueueDetails []struct {
	QueueDetail
	SeriesID       int  `json:"seriesId"`
	EpisodeID      int  `json:"episodeId"`
	SeasonNumber   int  `json:"seasonNumber"`
	EpisodeHasFile bool `json:"episodeHasFile"`
}

// From `GET /movie/{id}` or `GET /series/{id}`.
// Not all fields described here, just the wanted ones.
type MovieSerie struct {
	Title         string    `json:"title"`
	OriginalTitle string    `json:"originalTitle"`
	ID            int       `json:"id"`
	HasFile       bool      `json:"hasFile"`
	IsAvailable   bool      `json:"isAvailable"`
	Added         time.Time `json:"added"`
}

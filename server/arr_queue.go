package main

import (
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/sbondCo/Watcharr/arr"
)

type ArrDetailsResponse struct {
	Progress                int       `json:"progress"`
	EstimatedCompletionTime time.Time `json:"estimatedCompletionTime"`
	Status                  string    `json:"status"`
	TrackedDownloadStatus   string    `json:"trackedDownloadStatus"`
	TrackedDownloadState    string    `json:"trackedDownloadState"`
}

type SonarrDetailsResponseItem struct {
	ArrDetailsResponse
	Title string `json:"title"`
}

type SonarrDetailsResponse struct {
	// Overall progress
	Progress int `json:"progress"`
	// Overall status
	Status string `json:"status"`
	// Each downloading episode
	Items []SonarrDetailsResponseItem `json:"items"`
}

func getRadarrQueueDetails(serverName string, arrId string) (*ArrDetailsResponse, error) {
	server, err := getRadarr(serverName)
	if err != nil {
		slog.Error("getRadarrQueueDetails: Failed to get server", "error", err)
		return &ArrDetailsResponse{}, errors.New("failed to get server")
	}
	radarr := arr.New(arr.RADARR, &server.Host, &server.Key)
	// Run refresh downloads, likely won't be refreshed in time before we run GetQueueDetails below,
	// but if the user calls this again, it should be, which is better than waiting a whole minute for
	// refresh task to run automatically. This exists until a better solution is thought of.
	_, err = radarr.RunCommand("RefreshMonitoredDownloads")
	if err != nil {
		slog.Error("getRadarrQueueDetails: Failed to refresh monitored downloads.", "error", err)
	}
	resp := arr.RadarrQueueDetails{}
	err = radarr.GetQueueDetails(arrId, &resp)
	if err != nil {
		slog.Error("getRadarrQueueDetails: Failed to add content", "error", err)
		return &ArrDetailsResponse{}, errors.New("failed to add content")
	}
	if len(resp) <= 0 {
		slog.Error("getRadarrQueueDetails: No results in response")
		return &ArrDetailsResponse{}, errors.New("no details found")
	}
	adr := ArrDetailsResponse{}
	adr.Progress = int(math.Round((1 - (resp[0].SizeLeft / resp[0].Size)) * 100))
	adr.EstimatedCompletionTime = resp[0].EstimatedCompletionTime
	adr.TrackedDownloadStatus = resp[0].TrackedDownloadStatus
	adr.TrackedDownloadState = resp[0].TrackedDownloadState
	adr.Status = resp[0].Status
	return &adr, nil
}

func getSonarrQueueDetails(serverName string, arrId string) (*SonarrDetailsResponse, error) {
	server, err := getSonarr(serverName)
	if err != nil {
		slog.Error("getSonarrQueueDetails: Failed to get server", "error", err)
		return &SonarrDetailsResponse{}, errors.New("failed to get server")
	}
	sonarr := arr.New(arr.SONARR, &server.Host, &server.Key)
	// Run refresh downloads, likely won't be refreshed in time before we run GetQueueDetails below,
	// but if the user calls this again, it should be, which is better than waiting a whole minute for
	// refresh task to run automatically. This exists until a better solution is thought of.
	_, err = sonarr.RunCommand("RefreshMonitoredDownloads")
	if err != nil {
		slog.Error("getSonarrQueueDetails: Failed to refresh monitored downloads.", "error", err)
	}
	resp := arr.SonarrQueueDetails{}
	err = sonarr.GetQueueDetails(arrId, &resp)
	if err != nil {
		slog.Error("getSonarrQueueDetails: Failed to add content", "error", err)
		return &SonarrDetailsResponse{}, errors.New("failed to add content")
	}
	if len(resp) <= 0 {
		slog.Error("getSonarrQueueDetails: No results in response")
		return &SonarrDetailsResponse{}, errors.New("no details found")
	}
	dr := SonarrDetailsResponse{}
	totalSize := 0.0
	totalSizeLeft := 0.0
	for _, v := range resp {
		dr.Items = append(dr.Items, SonarrDetailsResponseItem{
			Title: "TODO :(", // TODO: Will need to do another request to get all episodes so we can get real episode names from EpisodeID
			ArrDetailsResponse: ArrDetailsResponse{
				Progress:                int(math.Round((1 - (v.SizeLeft / v.Size)) * 100)),
				EstimatedCompletionTime: v.EstimatedCompletionTime,
				TrackedDownloadStatus:   v.TrackedDownloadStatus,
				TrackedDownloadState:    v.TrackedDownloadState,
				Status:                  v.Status,
			},
		})
		totalSize += v.Size
		totalSizeLeft += v.SizeLeft
	}
	dr.Progress = int(math.Round((1 - (totalSizeLeft / totalSize)) * 100))
	// HACK: Statuses are not great, no easy way to get a good status, since there are multiple items to consider.
	// At some point, we could add extra checks (ex. if all items (episodes) are paused, overall status can be set to paused).
	// For now, this is good enough.
	dr.Status = "requested"
	if dr.Progress > 0 {
		dr.Status = "downloading"
	}
	return &dr, nil
}

// Refresh download queues for our sonarr/radarr servers.
// If the queues don't refresh regularly, our queue detail
// calls will just always return the same info.
func refreshArrQueues() {
	slog.Debug("refreshArrQueues: Refreshing queues for all configured arr servers.")
	// We don't care about responses, errors will be logged by the RunCommand func.
	for _, v := range Config.RADARR {
		radarr := arr.New(arr.RADARR, &v.Host, &v.Key)
		radarr.RunCommand("RefreshMonitoredDownloads")
	}
	for _, v := range Config.SONARR {
		sonarr := arr.New(arr.SONARR, &v.Host, &v.Key)
		sonarr.RunCommand("RefreshMonitoredDownloads")
	}
}

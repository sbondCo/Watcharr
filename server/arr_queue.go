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

func getRadarrQueueDetails(serverName string, arrId string) (ArrDetailsResponse, error) {
	server, err := getRadarr(serverName)
	if err != nil {
		slog.Error("createRadarrRequest: Failed to get server", "error", err)
		return ArrDetailsResponse{}, errors.New("failed to get server")
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
		slog.Error("createRadarrRequest: Failed to add content", "error", err)
		return ArrDetailsResponse{}, errors.New("failed to add content")
	}
	if len(resp) <= 0 {
		slog.Error("createRadarrRequest: No results in response")
		return ArrDetailsResponse{}, errors.New("no details found")
	}
	adr := ArrDetailsResponse{}
	adr.Progress = int(math.Round((1 - (resp[0].SizeLeft / resp[0].Size)) * 100))
	adr.EstimatedCompletionTime = resp[0].EstimatedCompletionTime
	adr.TrackedDownloadStatus = resp[0].TrackedDownloadStatus
	adr.TrackedDownloadState = resp[0].TrackedDownloadState
	adr.Status = resp[0].Status
	return adr, nil
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

package main

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/arr"
	"gorm.io/gorm"
)

// Deny an arr request
func denyArrRequest(db *gorm.DB, id uint) error {
	resp := db.Model(&ArrRequest{}).Where("id = ?", id).Update("status", ARR_REQUEST_DENIED)
	if resp.Error != nil {
		slog.Error("denyArrRequest: Failed to update status to denied", "error", resp.Error)
		return errors.New("failed when updating request status")
	}
	return nil
}

// Approve radarr movie
func approveRadarrRequest(db *gorm.DB, reqId uint, ur arr.RadarrRequest) (int, error) {
	_, err := getArrRequest(db, reqId)
	if err != nil {
		slog.Error("approveRadarrRequest: Failed to get request from db", "error", err)
		return 0, errors.New("failed to get request")
	}
	// Get server in request
	server, err := getRadarr(ur.ServerName)
	if err != nil {
		slog.Error("approveRadarrRequest: Failed to get server", "error", err)
		return 0, errors.New("failed to get server")
	}
	radarr := arr.New(arr.RADARR, &server.Host, &server.Key)
	ur.AutomaticSearch = server.AutomaticSearch
	resp, err := radarr.AddContent(radarr.BuildAddMovieBody(ur))
	if err != nil {
		slog.Error("approveRadarrRequest: Failed to add content", "error", err)
		return 0, errors.New("failed to add content")
	}
	dbResp := db.Model(&ArrRequest{}).Where("id = ?", reqId).Update("arr_id", resp["id"]).Update("status", ARR_REQUEST_APPROVED)
	if dbResp.Error != nil {
		slog.Error("approveRadarrRequest: Failed to update request in db", "error", err)
		return 0, errors.New("content was requested, but we failed to update the db")
	}
	arrId, ok := resp["id"].(float64)
	if !ok {
		slog.Error("approveRadarrRequest: Failed to cast arr id as an int", "id", resp["id"])
		return 0, errors.New("content added, but failed to get arr id")
	}
	return int(arrId), nil
}

// Approve sonarr movie
func approveSonarrRequest(db *gorm.DB, reqId uint, ur arr.SonarrRequest) (int, error) {
	_, err := getArrRequest(db, reqId)
	if err != nil {
		slog.Error("approveSonarrRequest: Failed to get request from db", "error", err)
		return 0, errors.New("failed to get request")
	}
	// Get server in request
	server, err := getSonarr(ur.ServerName)
	if err != nil {
		slog.Error("approveSonarrRequest: Failed to get server", "error", err)
		return 0, errors.New("failed to get server")
	}
	sonarr := arr.New(arr.SONARR, &server.Host, &server.Key)
	ur.AutomaticSearch = server.AutomaticSearch
	resp, err := sonarr.AddContent(sonarr.BuildAddShowBody(ur))
	if err != nil {
		slog.Error("approveSonarrRequest: Failed to add content", "error", err)
		return 0, errors.New("failed to add content")
	}
	dbResp := db.Model(&ArrRequest{}).Where("id = ?", reqId).Update("arr_id", resp["id"]).Update("status", ARR_REQUEST_APPROVED)
	if dbResp.Error != nil {
		slog.Error("approveSonarrRequest: Failed to update request in db", "error", err)
		return 0, errors.New("content was requested, but we failed to update the db")
	}
	arrId, ok := resp["id"].(float64)
	if !ok {
		slog.Error("approveSonarrRequest: Failed to cast arr id as an int", "id", resp["id"])
		return 0, errors.New("failed to get arr id")
	}
	return int(arrId), nil
}

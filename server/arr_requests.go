package main

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/arr"
	"gorm.io/gorm"
)

type ArrRequest struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uint      `json:"-" gorm:"not null"`
	User      User      `json:"-"`
	ContentID *int      `json:"-" gorm:"uniqueIndex:sn_to_cid;not null"`
	Content   *Content  `json:"content,omitempty"`
	// Server names are used as an identifier
	ServerName string `json:"serverName" gorm:"uniqueIndex:sn_to_cid;not null"`
	// Sonarr/Radarrs seriesId/movieId
	ArrID int `json:"arrId"`
}

func createArrRequest(db *gorm.DB, userId uint, serverName string, contentType ContentType, tmdbId int) (*ArrRequest, error) {
	content, err := getOrCacheContent(db, contentType, tmdbId)
	if err != nil {
		slog.Error("createArrRequest: getOrCacheContent errored.")
		return &ArrRequest{}, err
	}
	req := ArrRequest{UserID: userId, ServerName: serverName, ContentID: &content.ID}
	resp := db.Create(&req)
	if resp.Error != nil {
		slog.Error("createArrRequest: Failed when inserting request into db.", "error", err)
		return &ArrRequest{}, errors.New("failed when adding request")
	}
	return &req, nil
}

func createSonarrRequest(db *gorm.DB, userId uint, userPerms int, ur arr.SonarrRequest) error {
	server, err := getSonarr(ur.ServerName)
	if err != nil {
		slog.Error("createSonarrRequest: Failed to get server", "error", err)
		return errors.New("failed to get server")
	}
	arrReq, err := createArrRequest(db, userId, ur.ServerName, SHOW, ur.TMDBID)
	if err != nil {
		slog.Error("createSonarrRequest: Failed when creating arr request", "error", err)
		return errors.New("failed when creating request")
	}
	if hasPermission(userPerms, PERM_REQUEST_CONTENT_AUTO_APPROVE) {
		slog.Debug("createSonarrRequest: User has auto approve permission.. sending request to Sonarr.")
		ur.AutomaticSearch = server.AutomaticSearch
		sonarr := arr.New(arr.SONARR, &server.Host, &server.Key)
		resp, err := sonarr.AddContent(sonarr.BuildAddShowBody(ur))
		if err != nil {
			slog.Error("createSonarrRequest: Failed to add content", "error", err)
			return errors.New("failed to add content")
		}
		dbResp := db.Model(&ArrRequest{}).Where("id = ?", arrReq.ID).Update("arr_id", resp["id"])
		if dbResp.Error != nil {
			slog.Error("createSonarrRequest: Failed to update request in db", "error", err)
			return errors.New("content was requested, but we failed to update the db")
		}
	}
	return nil
}

func createRadarrRequest(db *gorm.DB, userId uint, userPerms int, ur arr.RadarrRequest) error {
	server, err := getRadarr(ur.ServerName)
	if err != nil {
		slog.Error("createRadarrRequest: Failed to get server", "error", err)
		return errors.New("failed to get server")
	}
	arrReq, err := createArrRequest(db, userId, ur.ServerName, MOVIE, ur.TMDBID)
	if err != nil {
		slog.Error("createRadarrRequest: Failed when creating arr request", "error", err)
		return errors.New("failed when creating request")
	}
	if hasPermission(userPerms, PERM_REQUEST_CONTENT_AUTO_APPROVE) {
		slog.Debug("createRadarrRequest: User has auto approve permission.. sending request to Radarr.")
		ur.AutomaticSearch = server.AutomaticSearch
		radarr := arr.New(arr.RADARR, &server.Host, &server.Key)
		resp, err := radarr.AddContent(radarr.BuildAddMovieBody(ur))
		if err != nil {
			slog.Error("createRadarrRequest: Failed to add content", "error", err)
			return errors.New("failed to add content")
		}
		dbResp := db.Model(&ArrRequest{}).Where("id = ?", arrReq.ID).Update("arr_id", resp["id"])
		if dbResp.Error != nil {
			slog.Error("createSonarrRequest: Failed to update request in db", "error", err)
			return errors.New("content was requested, but we failed to update the db")
		}
	}
	return nil
}

func getArrRequest(db *gorm.DB, contentType ContentType, tmdbId int) (ArrRequest, error) {
	var req ArrRequest
	resp := db.Joins("JOIN contents ON contents.id = arr_requests.content_id AND contents.tmdb_id = ? AND contents.type = ?", tmdbId, contentType).Find(&req)
	if resp.Error != nil {
		slog.Error("getArrRequest: Failed to search for request in db", "error", resp.Error)
		return ArrRequest{}, errors.New("failed to find request")
	}
	return req, nil
}

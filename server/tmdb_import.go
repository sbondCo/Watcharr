package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TMDBImportRequest struct {
	TmdbID        uint        `json:"tmdbId"`
	ImdbID        string      `json:"imdbID"`
	Type          ContentType `json:"type"`
	Name          string      `json:"name"`
	ReleaseDate   time.Time   `json:"releaseDate"`
	SeasonNumber  uint        `json:"seasonNumber"`
	EpisodeNumber uint        `json:"episodeNumber"`
	Rating        float64     `json:"rating"`
	DateRated     time.Time   `json:"dateRated"`
	State         string      `json:"state"`
	Year          string      `json:"year"`
}

func importTmdbContent(db *gorm.DB, userId uint, importRequest TMDBImportRequest) (ImportResponse, error) {
	slog.Info("Importing item", "userId", userId, "contentType", importRequest.Type, "contentId", importRequest.TmdbID)

	err := fetchAndAddContentToDb(db, importRequest)
	if err != nil {
		return ImportResponse{}, err
	}

	watchedEntry, err := createWatchedEntry(db, userId, importRequest)
	if err != nil {
		return ImportResponse{}, err
	}
	// TmdbImportAddActivity(db, userId, content, ar, IMPORTED_WATCHED, watchedEntry)

	return ImportResponse{Type: IMPORT_SUCCESS, WatchedEntry: watchedEntry}, nil
}

func parseMovieContent(resp []byte) (ContentDetails, error) {
	content := new(TMDBMovieDetails)
	err := json.Unmarshal([]byte(resp), &content)
	if err != nil {
		slog.Error("Failed to unmarshal movie details", "error", err)
		return ContentDetails{}, errors.New("failed to process movie details response")
	}
	releaseDate, err := time.Parse("2006-01-02", content.ReleaseDate)
	if err != nil {
		slog.Error("Failed to parse movie release date", "error", err)
		return ContentDetails{}, errors.New("failed to parse movie release date")
	}
	return ContentDetails{
		TmdbID:           content.ID,
		Title:            content.Title,
		Overview:         content.Overview,
		PosterPath:       content.PosterPath,
		Type:             "movie",
		ReleaseDate:      &releaseDate,
		Popularity:       content.Popularity,
		VoteAverage:      content.VoteAverage,
		VoteCount:        content.VoteCount,
		ImdbID:           content.ImdbID,
		Status:           content.Status,
		Budget:           content.Budget,
		Revenue:          content.Revenue,
		Runtime:          content.Runtime,
		NumberOfEpisodes: 0,
		NumberOfSeasons:  0,
	}, nil
}

func parseShowContent(resp []byte) (ContentDetails, error) {
	content := new(TMDBShowDetails)
	err := json.Unmarshal([]byte(resp), &content)
	if err != nil {
		slog.Error("Failed to unmarshal show details", "error", err)
		return ContentDetails{}, errors.New("failed to process show details response")
	}
	releaseDate, err := time.Parse("2006-01-02", content.FirstAirDate)
	if err != nil {
		slog.Error("Failed to parse show release date", "error", err)
		return ContentDetails{}, errors.New("failed to parse show release date")
	}
	var runtime uint32
	if len(content.EpisodeRunTime) > 0 {
		runtime = uint32(content.EpisodeRunTime[0])
	}
	return ContentDetails{
		TmdbID:           content.ID,
		Title:            content.Name,
		Overview:         content.Overview,
		PosterPath:       content.PosterPath,
		Type:             "tv",
		ReleaseDate:      &releaseDate,
		Popularity:       content.Popularity,
		VoteAverage:      content.VoteAverage,
		VoteCount:        content.VoteCount,
		ImdbID:           "",
		Status:           content.Status,
		Budget:           0,
		Revenue:          0,
		Runtime:          runtime,
		NumberOfEpisodes: content.NumberOfEpisodes,
		NumberOfSeasons:  content.NumberOfSeasons,
	}, nil
}

func fetchContent(db *gorm.DB, contentType ContentType, tmdbID int) (ContentDetails, error) {
	slog.Debug("Content not in db, fetching...")

	resp, err := tmdbAPIRequest("/"+string(contentType)+"/"+strconv.Itoa(tmdbID), map[string]string{})
	if err != nil {
		slog.Error("addWatched content tmdb api request failed", "error", err)
		return ContentDetails{}, errors.New("failed to find requested media")
	}

	// Get details from movie/show response and fill out needed vars
	if contentType == "movie" {
		return parseMovieContent(resp)
	} else {
		return parseShowContent(resp)
	}
}

func fetchAndAddContentToDb(db *gorm.DB, importRequest TMDBImportRequest) error {

	// Exists if content found in db
	var content ContentDetails
	db.Where("tmdb_id = ?", importRequest.TmdbID).Find(&content)
	if content != (ContentDetails{}) {
		return nil
	}

	var err error
	content, err = fetchContent(db, importRequest.Type, int(importRequest.TmdbID))
	if err != nil {
		slog.Error("addWatched, failed to get content details", "error", err)
		return err
	}
	res := db.Create(&content)
	if res.Error != nil {
		// Error if anything but unique contraint error
		if !strings.Contains(res.Error.Error(), "UNIQUE") {
			slog.Error("Error creating content in database", "error", res.Error.Error())
			return errors.New("failed to cache content in database")
		}
	}
	// If row created, download the image
	if res.RowsAffected > 0 {
		err := download("https://image.tmdb.org/t/p/w500"+content.PosterPath, path.Join("./data/img", content.PosterPath))
		if err != nil {
			slog.Error("Failed to download content image!", "error", err.Error())
		}
	}
	return nil
}

func createWatchedEntry(db *gorm.DB, userId uint, importRequest TMDBImportRequest) (Watched, error) {
	watched := Watched{Status: FINISHED, Rating: int8(importRequest.Rating), UserID: userId, ContentID: int(importRequest.TmdbID)}
	res := db.Create(&watched)
	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "UNIQUE") {
			res = db.Model(&Watched{}).Unscoped().Preload("Activity").Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Take(&watched)
			if res.Error != nil {
				return Watched{}, errors.New("content already on watched list. errored checking for soft deleted record")
			}
			if watched.DeletedAt.Time.IsZero() {
				return Watched{}, errors.New("content already on watched list")
			} else {
				slog.Info("addWatched: Watched list item for this content exists as soft deleted record.. attempting to restore")
				res = db.Model(&Watched{}).Unscoped().Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Updates(map[string]interface{}{"status": FINISHED, "rating": int8(importRequest.Rating), "deleted_at": nil})
				watched.Status = FINISHED
				watched.Rating = int8(importRequest.Rating)
				if res.Error != nil {
					slog.Error("addWatched: Failed to restore soft deleted watch list item", "error", res.Error)
					return Watched{}, errors.New("content already on watched list. errored removing soft delete timestamp")
				}
			}
		} else {
			slog.Error("Error adding watched content to database", "error", res.Error.Error())
			return Watched{}, errors.New("failed adding content to database")
		}
	}
	return watched, nil
}

func TmdbImportAddActivity(db *gorm.DB, userId uint, addRequest WatchedAddRequest, content ContentDetails, at ActivityType, watched Watched) error {

	var activity Activity
	activityJson, err := json.Marshal(map[string]interface{}{"status": addRequest.Status, "rating": addRequest.Rating})
	if err != nil {
		slog.Error("Failed to marshal json for data in ADD_WATCHED activity request, adding without data", "error", err.Error())
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: uint(addRequest.ContentID), Type: at})
	} else {
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: uint(addRequest.ContentID), Type: at, Data: string(activityJson)})
	}
	watched.Activity = append(watched.Activity, activity)
	watched.Content = content
	return nil
}

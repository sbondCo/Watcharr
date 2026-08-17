package season

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

// Add/edit a watched season.
func (s *Service) AddWatchedSeason(
	userId uint,
	ar domain.WatchedSeasonAddRequest,
) (domain.WatchedSeasonAddResponse, error) {
	slog.Debug("Adding watched season item", "userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber)
	// 1. Make sure watched item exists and it is the correct type (TV)
	var w entity.Watched
	if resp := s.db.
		Where("id = ? AND user_id = ?", ar.WatchedID, userId).
		Preload("Content").
		Preload("WatchedSeasons").
		Find(&w); resp.Error != nil {
		slog.Error("Failed when adding a watched season", "error", "failed to get watched item from db")
		return domain.WatchedSeasonAddResponse{}, errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("Failed when adding a watched season", "error", "watched item does not exist in db")
		return domain.WatchedSeasonAddResponse{}, errors.New("can't add a watched season for a show that doesnt have a status itself")
	}
	if w.Content.Type != entity.SHOW {
		return domain.WatchedSeasonAddResponse{}, errors.New("can't add watched season for non show content")
	}
	found := false
	updated := false
	for i, ws := range w.WatchedSeasons {
		if ws.SeasonNumber == ar.SeasonNumber {
			slog.Debug("Existing watched season item found, updating existing")
			found = true
			if ar.Status != "" && ar.Status != w.WatchedSeasons[i].Status {
				w.WatchedSeasons[i].Status = ar.Status
				updated = true
			}
			if ar.Rating != 0 && ar.Rating != w.WatchedSeasons[i].Rating {
				w.WatchedSeasons[i].Rating = ar.Rating
				updated = true
			}
			break
		}
	}
	var addedActivity entity.Activity
	if !found {
		slog.Debug("Existing watched season not found, adding as new entry")
		w.WatchedSeasons = append(w.WatchedSeasons, entity.WatchedSeason{
			UserID:       userId,
			WatchedID:    ar.WatchedID,
			SeasonNumber: ar.SeasonNumber,
			Status:       ar.Status,
			Rating:       ar.Rating,
		})
	}
	if resp := s.db.Save(&w.WatchedSeasons); resp.Error != nil {
		slog.Debug("Failed to save watched season item in db", "error", resp.Error)
		return domain.WatchedSeasonAddResponse{}, errors.New("failed to save")
	}
	// Add activity
	if found {
		// Only add change activity if we actually updated a value
		// (changing value to same value doesn't count).
		if updated {
			if ar.Status != "" {
				json, _ := json.Marshal(map[string]interface{}{
					"season": ar.SeasonNumber,
					"status": ar.Status,
				})
				addedActivity, _ = activity.
					NewCreator(s.db, userId, w.ID, entity.SEASON_STATUS_CHANGED, false, 0).
					SetData(string(json)).
					Create()
			}
			if ar.Rating != 0 {
				json, _ := json.Marshal(map[string]interface{}{
					"season": ar.SeasonNumber,
					"rating": ar.Rating,
				})
				addedActivity, _ = activity.
					NewCreator(s.db, userId, w.ID, entity.SEASON_RATING_CHANGED, false, 0).
					SetData(string(json)).
					Create()
			}
		}
	} else {
		actData := map[string]interface{}{
			"season": ar.SeasonNumber,
			"status": ar.Status,
			"rating": ar.Rating,
		}
		if len(ar.AddActivityData) > 0 {
			for k, v := range ar.AddActivityData {
				if _, ok := ar.AddActivityData[k]; ok {
					actData[k] = v
				}
			}
		}
		json, _ := json.Marshal(actData)
		act := activity.
			NewCreator(s.db, userId, w.ID, entity.SEASON_ADDED, false, 0).
			SetData(string(json))
		if ar.AddActivity != "" {
			act.SetType(ar.AddActivity)
		}
		if !ar.AddActivityDate.IsZero() {
			act.SetCustomDate(&ar.AddActivityDate)
		}
		addedActivity, _ = act.Create()
	}
	return domain.WatchedSeasonAddResponse{
		WatchedSeasons: w.WatchedSeasons,
		AddedActivity:  addedActivity,
	}, nil
}

// Remove a watched season
func (s *Service) RmWatchedSeason(userId uint, seasonId uint) (entity.Activity, error) {
	slog.Debug("rmWatchedSeason called", "user_id", userId, "season_id", seasonId)
	var watchedSeason entity.WatchedSeason
	resp := s.db.
		Clauses(clause.Returning{}).
		Model(&entity.WatchedSeason{}).
		Unscoped().
		Where("id = ? AND user_id = ?", seasonId, userId).
		Delete(&watchedSeason)
	if resp.Error != nil {
		slog.Error("Failed when removing a watched season", "error", resp.Error)
		return entity.Activity{}, errors.New("failed when removing watched season")
	}
	if resp.RowsAffected == 0 {
		slog.Error("Failed when removing a watched season", "error", "zero rows affected")
		return entity.Activity{}, errors.New("wasn't removed from db.. may not exist")
	}
	slog.Debug("rmWatchedSeason, deleted row", "row", watchedSeason)
	if watchedSeason.ID != 0 {
		json, _ := json.Marshal(map[string]interface{}{
			"season": watchedSeason.SeasonNumber,
			"status": watchedSeason.Status,
			"rating": watchedSeason.Rating,
		})
		addedActivity, _ := activity.
			NewCreator(s.db, userId, watchedSeason.WatchedID, entity.SEASON_REMOVED, false, 0).
			SetData(string(json)).
			Create()
		return addedActivity, nil
	}
	return entity.Activity{}, errors.New("removed, but failed to add activity entry")
}

func (s *Service) GetWatchedSeason(userId uint, watchedId uint, seasonNumber int) (*entity.WatchedSeason, error) {
	var ws *entity.WatchedSeason
	if res := s.db.Model(&entity.WatchedSeason{}).Where("watched_id = ? AND season_number = ? AND user_id = ?", watchedId, seasonNumber, userId).Take(&ws); res.Error != nil {
		slog.Error("getWatchedSeason: Failed to get:", "error", res.Error.Error())
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return &entity.WatchedSeason{}, errors.New("failed to get watched season")
	}
	return ws, nil
}

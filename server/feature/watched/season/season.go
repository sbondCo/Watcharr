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

type WatchedProvider interface {
	IsWatchedItemContentType(userId uint, id uint, ct entity.ContentType) error
}

type Service struct {
	db *gorm.DB
	wp WatchedProvider
}

func NewService(db *gorm.DB, wp WatchedProvider) *Service {
	return &Service{
		db,
		wp,
	}
}

// Get WatchedSeason.
// Returns nil for entity.WatchedSeason if it doesn't exist.
func (s *Service) GetWatchedSeason(
	userId uint,
	watchedId uint,
	seasonNumber int,
) (*entity.WatchedSeason, error) {
	var ws *entity.WatchedSeason
	err := s.db.
		Model(&entity.WatchedSeason{}).
		Where("watched_id = ? AND user_id = ? AND season_number = ?",
			watchedId, userId, seasonNumber).
		Take(&ws).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Debug("GetWatchedSeason: Record not found.")
			return nil, nil
		}
		slog.Error("GetWatchedSeason: Query failed!", "error", err)
		return &entity.WatchedSeason{}, errors.New("failed to get watched season")
	}
	return ws, nil
}

// Add/edit a watched season.
func (s *Service) SetWatchedSeason(
	userId uint,
	ar domain.WatchedSeasonSetRequest,
) (domain.WatchedSeasonSetResponse, error) {
	slog.Debug("SetWatchedSeason: Setting.",
		"userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber)

	// Make sure watched item exists and is the correct type (TV)
	err := s.wp.IsWatchedItemContentType(userId, ar.WatchedID, entity.SHOW)
	if err != nil {
		slog.Error("SetWatchedSeason: Failed!", "error", err)
		return domain.WatchedSeasonSetResponse{},
			errors.New("failed or invalid watched entry targetted")
	}

	// Try to lookup existing Watched Season.
	wSeason, err := s.GetWatchedSeason(userId, ar.WatchedID, ar.SeasonNumber)
	if err != nil {
		slog.Error("SetWatchedSeason: WatchedSeason query failed!",
			"error", err)
		return domain.WatchedSeasonSetResponse{},
			errors.New("couldn't get watched season")
	}

	resp := domain.WatchedSeasonSetResponse{}
	activityMC := activity.NewMultiCreator()

	// Add or update Watched Season.
	if wSeason != nil {
		slog.Debug("SetWatchedSeason: Updating existing.")

		if ar.Status != "" && ar.Status != wSeason.Status {
			wSeason.Status = ar.Status
			resp.Update = true
			activity.
				NewCreator(
					s.db,
					userId,
					ar.WatchedID,
					entity.SEASON_STATUS_CHANGED,
					false,
					ar.ActivityCreatedBy,
				).
				SetDataJSON(map[string]interface{}{
					"season": ar.SeasonNumber,
					"status": ar.Status,
				}).
				AddToMultiCreator(activityMC)
		}

		if ar.Rating != 0 && ar.Rating != wSeason.Rating {
			wSeason.Rating = ar.Rating
			resp.Update = true
			activity.
				NewCreator(
					s.db,
					userId,
					ar.WatchedID,
					entity.SEASON_RATING_CHANGED,
					false,
					ar.ActivityCreatedBy,
				).
				SetDataJSON(map[string]interface{}{
					"season": ar.SeasonNumber,
					"rating": ar.Rating,
				}).
				AddToMultiCreator(activityMC)
		}

		// If Update still = False, then no changes were made! Exit early.
		if !resp.Update {
			slog.Warn("SetWatchedSeason: No changes were necessary.")
			// We don't return an error so we retain idempotency for this method.
			// Update must be set to True before returning the response, otherwise
			// we are indicating that the WatchedSeason was just Created, which
			// could cause bugs in the client.
			// Made a new Response object instead of returning `resp` to avoid
			// bugs where we alter something above in it and it gets returned
			// here mistakenly.
			return domain.WatchedSeasonSetResponse{
				Update:        true,
				WatchedSeason: *wSeason,
			}, nil
		}
	} else {
		slog.Debug("SetWatchedSeason: Creating new.")

		wSeason = &entity.WatchedSeason{
			UserID:       userId,
			WatchedID:    ar.WatchedID,
			SeasonNumber: ar.SeasonNumber,
			Status:       ar.Status,
			Rating:       ar.Rating,
		}

		activity.
			NewCreator(
				s.db,
				userId,
				ar.WatchedID,
				entity.SEASON_ADDED,
				false,
				ar.ActivityCreatedBy,
			).
			SetDataJSON(map[string]interface{}{
				"season": ar.SeasonNumber,
				"status": ar.Status,
				"rating": ar.Rating,
			}).
			SetCustomDate(&ar.AddActivityDate).
			SetReason(ar.AddActivityReason).
			AddToMultiCreator(activityMC)
	}

	if err := s.db.Save(wSeason).Error; err != nil {
		slog.Debug("SetWatchedSeason: Save query failed.", "error", err)
		return domain.WatchedSeasonSetResponse{}, errors.New("failed to save")
	}

	addedActivities := activityMC.CreateAll()

	resp.WatchedSeason = *wSeason
	resp.AddedActivities = addedActivities

	return resp, nil
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

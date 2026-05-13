package activity

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

func (s *Service) GetActivity(userId uint, watchedId uint) ([]entity.Activity, error) {
	activity := new([]entity.Activity)
	res := s.db.Model(&entity.Activity{}).Where("user_id = ? AND watched_id = ?", userId, watchedId).Find(&activity)
	if res.Error != nil {
		slog.Error("Failed getting activity from database", "error", res.Error.Error())
		return []entity.Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

func (s *Service) AddActivity(userId uint, ar domain.ActivityAddRequest) (entity.Activity, error) {
	if ar.WatchedID == 0 {
		return entity.Activity{}, errors.New("watchedId must be set to add an activity")
	}
	// Verify watched entry belongs to the calling user.
	var count int64
	if res := s.db.Model(&entity.Watched{}).Where("id = ? AND user_id = ?", ar.WatchedID, userId).Count(&count); res.Error != nil {
		slog.Error("AddActivity: failed to verify watched ownership", "error", res.Error)
		return entity.Activity{}, errors.New("failed to verify watched entry")
	}
	if count == 0 {
		slog.Warn("AddActivity: user attempted to add activity for watched entry they do not own", "user_id", userId, "watched_id", ar.WatchedID)
		return entity.Activity{}, errors.New("watched entry does not exist")
	}
	activity := entity.Activity{UserID: userId, WatchedID: ar.WatchedID, Type: ar.Type, Data: ar.Data, CustomDate: ar.CustomDate}
	res := s.db.Create(&activity)
	if res.Error != nil {
		slog.Error("Error adding activity to database", "error", res.Error.Error())
		return entity.Activity{}, errors.New("failed adding new activity to database")
	}
	slog.Debug("Adding activity", "added_activity", activity)
	return activity, nil
}

func (s *Service) UpdateActivity(userId uint, id uint, activityUpdateRequest domain.ActivityUpdateRequest) error {
	if id == 0 {
		return errors.New("id must be set to update an activity")
	}
	if activityUpdateRequest.CustomDate.IsZero() {
		return errors.New("customDate must be set to update an activity")
	}
	res := s.db.
		Model(&entity.Activity{}).
		Where("user_id = ? AND id = ?", userId, id).
		Update("custom_date", activityUpdateRequest.CustomDate)
	if res.Error != nil {
		slog.Error("Error updating activity in database", "error", res.Error.Error())
		return errors.New("failed updating activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were updated. This may be because the activity doesn't exist or is not owned by the calling user.")
		return errors.New("failed updating activity in database")
	}
	slog.Debug("Updating activity", "updated_activity", id)
	return nil
}

func (s *Service) DeleteActivity(userId uint, id uint) error {
	if id == 0 {
		return errors.New("an id must be provided to delete an activity")
	}
	res := s.db.Where("user_id = ?", userId).Delete(&entity.Activity{}, id)
	if res.Error != nil {
		slog.Error("Error deleting activity in database", "error", res.Error.Error())
		return errors.New("failed deleting activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were deleted. This may be because the activity doesn't exist or is not owned by the calling user.")
		return errors.New("failed deleting activity from database")
	}
	return nil
}

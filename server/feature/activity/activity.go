package activity

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/tri"
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

func (s *Service) GetActivity(
	userId uint,
	watchedId uint,
) ([]entity.Activity, error) {
	activity := new([]entity.Activity)
	res := s.db.Model(&entity.Activity{}).
		Where("user_id = ? AND watched_id = ?", userId, watchedId).
		Find(&activity)
	if res.Error != nil {
		slog.Error("Failed getting activity from database",
			"error", res.Error.Error())
		return []entity.Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

// NOTE: Currently this function doesn't verify if `userId` owns the referenced
// watched item at `ar.WatchedID`. If we ever need this function to work from an
// "AddActivity" endpoint on the API, we should create another func that has
// that validation, since this func is only for internal operations!
// AddActivity: Only for internal use.
func (s *Service) AddActivity(
	userId uint,
	ar domain.ActivityAddProps,
	extra domain.ActivityAddExtraProps,
) (entity.Activity, error) {
	if ar.WatchedID == 0 {
		return entity.Activity{},
			errors.New("watchedId must be set to add an activity")
	}
	if extra.CountAsPlay == tri.Unset {
		// All callers should specify if activity CountsAsPlay!
		return entity.Activity{}, errors.New("didn't specific CountAsPlay")
	}
	activity := entity.Activity{
		UserID:      userId,
		WatchedID:   ar.WatchedID,
		Type:        ar.Type,
		Data:        ar.Data,
		CustomDate:  ar.CustomDate,
		CountAsPlay: tri.ToBool(extra.CountAsPlay),
		SyncedBy:    extra.SyncedBy,
	}
	res := s.db.Create(&activity)
	if res.Error != nil {
		slog.Error("Error adding activity to database",
			"error", res.Error.Error())
		return entity.Activity{},
			errors.New("failed adding new activity to database")
	}
	slog.Debug("Adding activity", "added_activity", activity)
	return activity, nil
}

func (s *Service) UpdateActivity(
	userId uint,
	id uint,
	activityUpdateRequest domain.ActivityUpdateRequest,
) error {
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
		slog.Error("Error updating activity in database",
			"error", res.Error.Error())
		return errors.New("failed updating activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were updated.")
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
		slog.Error("Error deleting activity in database",
			"error", res.Error.Error())
		return errors.New("failed deleting activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were deleted.")
		return errors.New("failed deleting activity from database")
	}
	return nil
}

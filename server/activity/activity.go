package activity

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

func Get(
	db *gorm.DB,
	userId uint,
	watchedId uint,
) ([]entity.Activity, error) {
	activity := new([]entity.Activity)
	res := db.Model(&entity.Activity{}).
		Where("user_id = ? AND watched_id = ?", userId, watchedId).
		Find(&activity)
	if res.Error != nil {
		slog.Error("Get: Query failed!", "error", res.Error)
		return []entity.Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

// Update activity. Currently only allows updating custom_date (needs refactoring if other edits are allowed).
func Update(
	db *gorm.DB,
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
	res := db.
		Model(&entity.Activity{}).
		Where("user_id = ? AND id = ?", userId, id).
		Update("custom_date", activityUpdateRequest.CustomDate)
	if res.Error != nil {
		slog.Error("Update: Query failed!", "error", res.Error)
		return errors.New("query failed")
	}
	if res.RowsAffected < 1 {
		slog.Error("Update: No activities were updated.")
		return errors.New("no rows affected")
	}
	slog.Debug("Update: Updated activity", "updated_activity", id)
	return nil
}

func Delete(db *gorm.DB, userId uint, id uint) error {
	if id == 0 {
		return errors.New("an id wasn't provided")
	}
	res := db.
		Where("user_id = ?", userId).
		Delete(&entity.Activity{}, id)
	if res.Error != nil {
		slog.Error("Delete: Query failed!", "error", res.Error)
		return errors.New("query failed")
	}
	if res.RowsAffected < 1 {
		slog.Error("Delete: No activities were deleted.")
		return errors.New("no rows affected")
	}
	return nil
}

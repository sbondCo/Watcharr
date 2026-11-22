package services

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

// Get all user services for UserID.
func (s *Service) GetAll(userID uint) ([]entity.UserServices, error) {
	us := []entity.UserServices{}
	err := s.db.
		Model(&entity.UserServices{}).
		Where("user_id = ?", userID).
		Find(&us).
		Error
	if err != nil {
		slog.Error("GetAll: DB Query failed!", "error", err)
		return us, errors.New("failed to get user services")
	}
	return us, nil
}

// Get one service by ID.
func (s *Service) Get(id uint, userID uint) (entity.UserServices, error) {
	us := entity.UserServices{}
	err := s.db.
		Model(&entity.UserServices{}).
		Where("id = ? AND user_id = ?", id, userID).
		Take(&us).
		Error
	if err != nil {
		slog.Error("Get: Query failed!", "error", err)
		return us, errors.New("failed to get user service")
	}
	return us, nil
}

func (s *Service) Update(
	id uint,
	userID uint,
	req domain.UserServicePatchSettingsDto,
) (domain.UserServicePatchSettingsDto, error) {
	// Technically we could just return `req` back to the user on success,
	// but that could potentially break if we forget or don't want to allow
	// updating certain settings in the future.
	resp := domain.UserServicePatchSettingsDto{}

	q := s.db.
		Model(&entity.UserServices{}).
		Where("id = ? AND user_id = ?", id, userID)

	updm := map[string]any{}
	if req.AutoSync != nil {
		updm["auto_sync"] = req.AutoSync
		resp.AutoSync = req.AutoSync
	}

	if len(updm) == 0 {
		slog.Warn("Update: Nothing provided to update.")
		return resp, errors.New("provided nothing to update")
	}

	qr := q.Updates(updm)
	if qr.Error != nil {
		slog.Error("Update: Query failed!", "error", qr.Error)
		return resp, errors.New("update failed")
	}
	if qr.RowsAffected == 0 {
		slog.Error("Update: No rows updated! Service doesn't exist or belong to this user.",
			"user_id", userID, "service_id", id)
		return resp, errors.New("service doesn't exist")
	}

	return resp, nil
}

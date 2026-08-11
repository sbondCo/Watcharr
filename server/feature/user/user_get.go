package user

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

// Get user by id.
// Returns a builder for easily adding preloads.
func (s *Service) GetUserByID(id uint) domain.GetUserQueryBuilder {
	db := s.db.
		Model(&entity.User{}).
		Where("id = ?", id)
	return &getUserQueryBuilder{
		db: db,
	}
}

type getUserQueryBuilder struct {
	db *gorm.DB
}

// Preload UserServices.
func (q *getUserQueryBuilder) WithUserServices() domain.GetUserQueryBuilder {
	q.db.Preload("UserServices")
	return q
}

// Done building the query, now execute.
func (q *getUserQueryBuilder) Done() (entity.User, error) {
	slog.Debug("getUserQueryBuilder: Running query.")
	user := new(entity.User)
	err := q.db.Take(&user).Error
	if err != nil {
		slog.Error("getUserQueryBuilder: Query failed!", "error", err)
		return *user, errors.New("query failed")
	}
	return *user, nil
}

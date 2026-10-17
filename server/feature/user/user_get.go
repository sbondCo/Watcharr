package user

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

// NOTE: I don't think this builder is useful. I can see it ending up with just
// GORM's api being wrapped in my api, so this won't be used for future code,
// and may just be dropped completely for services to always make their own
// queries. Possibly just create a simple UserRepository, but only for simple
// methods like GetUserByID, but not sure if that's worth the hassle rn.

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

package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

// User
type (
	UserBioUpdateRequest struct {
		NewBio string `json:"newBio" binding:"max=128"`
	}
)

// User Manage
type (
	// User details wanted for management views.
	ManagedUser struct {
		ID          uint            `json:"id"`
		CreatedAt   time.Time       `json:"createdAt"`
		Username    string          `json:"username"`
		Type        entity.UserType `json:"type"`
		Permissions int             `json:"permissions"`
		Private     bool            `json:"private"`
	}

	UpdateUserRequest struct {
		Permissions *int             `json:"permissions"`
		Type        *entity.UserType `json:"type"`
	}

	UserManageProvider interface {
		GetAll() ([]ManagedUser, error)
		Manage(userId uint, ur UpdateUserRequest) error
	}
)

// Query builder for getting User from db.
type GetUserQueryBuilder interface {
	WithUserServices() GetUserQueryBuilder
	Done() (entity.User, error)
}

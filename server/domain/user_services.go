package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type UserService struct {
	ID        uint                   `json:"id"`
	CreatedAt time.Time              `json:"createdAt"`
	Name      entity.UserServiceName `json:"name"`
	AutoSync  bool                   `json:"autoSync"`
}

// Request AND Response struct for PatchSettings.
type UserServicePatchSettingsDto struct {
	AutoSync *bool `json:"autoSync,omitempty"`
}

type AllUserServicesResponse []UserService

func NewAllUserServicesResponse(us []entity.UserServices) AllUserServicesResponse {
	r := AllUserServicesResponse{}
	for _, v := range us {
		r = append(r, UserService{
			ID:        v.ID,
			Name:      v.Name,
			CreatedAt: v.CreatedAt,
			AutoSync:  v.AutoSync,
		})
	}
	return r
}

package tag

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

// I think tags will be private for the user.
// If the user wants to make a public list, they should make a custom view.
// Hi from 2026, the above sounds stupid, why dont you just combine custom view features into tags view..

type WatchedProvider interface {
	GetWatchedPage(userId uint, pp util.PaginationParams, wr domain.WatchedGetPageRequest, extraProps *domain.WatchedGetPageExtraProps) (util.PaginationResponse[entity.Watched, util.None], error)
}

type Service struct {
	db              *gorm.DB
	watchedProvider WatchedProvider
}

func NewService(db *gorm.DB, watchedProvider WatchedProvider) *Service {
	return &Service{
		db,
		watchedProvider,
	}
}

func (s *Service) GetTags(userId uint) ([]entity.Tag, error) {
	tags := new([]entity.Tag)
	res := s.db.Model(&entity.Tag{}).Where("user_id = ?", userId).Find(&tags)
	if res.Error != nil {
		slog.Error("getTags: Failed getting tags from database", "error", res.Error.Error())
		return []entity.Tag{}, errors.New("failed getting tags")
	}
	return *tags, nil
}

func (s *Service) GetTag(userId uint, tagId uint) (entity.Tag, error) {
	tag := new(entity.Tag)
	res := s.db.
		Model(&entity.Tag{}).
		Where("id = ? AND user_id = ?", tagId, userId).
		Preload("Watched").
		Preload("Watched.Content").
		Find(&tag)
	if res.Error != nil {
		slog.Error("getTag: Failed getting tag from database",
			"tag_id", tagId, "error", res.Error.Error())
		return entity.Tag{}, errors.New("failed getting tag")
	}
	if tag.ID == 0 {
		slog.Error("getTag: Tag does not exist for this user.",
			"user_id", userId, "tag_id", tagId)
		return entity.Tag{}, errors.New("tag does not exist")
	}
	return *tag, nil
}

func (s *Service) GetTagPage(
	userId uint,
	tagId uint,
	pp util.PaginationParams,
	wr domain.WatchedGetPageRequest,
) (util.PaginationResponse[entity.Watched, util.None], error) {
	slog.Debug("GetTagPage: A page was requested.",
		"user_id", userId,
		"tagId", tagId,
		"pagination_params", pp,
		"wr", wr)

	// Attempt to get the tag, verifying it exists and the user own it.
	_, err := s.GetTag(userId, tagId)
	if err != nil {
		return util.PaginationResponse[entity.Watched, util.None]{}, err
	}

	// Get all watched ids for this tag.
	// This isn't great, but it has to be done, since we can't eliminate
	// any for sort/filters here.
	wids := new([]int)
	res := s.db.
		Table("watched_tags").
		Select("watched_id").
		Where("tag_id = ?", tagId).
		Find(&wids)
	if res.Error != nil {
		slog.Error("GetTagPage: Getting watched ids failed!", "error", err)
		return util.PaginationResponse[entity.Watched, util.None]{}, err
	}
	if len(*wids) <= 0 {
		slog.Debug("GetTagPage: The requested tag has no watched items!")
		return util.PaginationResponse[entity.Watched, util.None]{}, nil
	}

	// Now get a watched page, passing in our fetched watched ids
	// so that only our watched items in this tag are retrieved.
	wp, err := s.watchedProvider.GetWatchedPage(
		userId,
		pp,
		wr,
		&domain.WatchedGetPageExtraProps{
			WatchedIds: *wids,
		},
	)
	if err != nil {
		slog.Error("GetTagPage: Getting watcheds failed!", "error", err)
		return util.PaginationResponse[entity.Watched, util.None]{}, err
	}

	return wp, nil
}

// This method should only be used when we don't have the tagId
// (eg: when we are importing data) because this is not technically
// reliable, since users can have multiple tags with the same name/colors
// (realistically they probably won't, but...).
func (s *Service) GetTagByNameAndColor(
	userId uint,
	tagName string,
	tagColor string,
	tagBgColor string,
) (entity.Tag, error) {
	tag := new(entity.Tag)
	res := s.db.
		Model(&entity.Tag{}).
		Where("name = ? AND user_id = ? AND color = ? AND bg_color = ?",
			tagName, userId, tagColor, tagBgColor).
		Preload("Watched").
		Find(&tag)
	if res.Error != nil {
		slog.Error("getTagByNameAndColor: Failed getting tag from database", "error", res.Error.Error())
		return entity.Tag{}, errors.New("failed getting tag")
	}
	if tag.ID == 0 {
		slog.Error("getTagByNameAndColor: Tag does not exist for this user.", "user_id", userId)
		return entity.Tag{}, errors.New("tag does not exist")
	}
	return *tag, nil
}

// Let user create a tag.
func (s *Service) AddTag(userId uint, tr domain.TagAddRequest) (entity.Tag, error) {
	if tr.Name == "" {
		return entity.Tag{}, errors.New("tag must have a name")
	}
	tag := entity.Tag{UserID: userId, Name: tr.Name, Color: tr.Color, BgColor: tr.BgColor}
	res := s.db.Create(&tag)
	if res.Error != nil {
		slog.Error("Error adding tag to database", "error", res.Error.Error())
		return entity.Tag{}, errors.New("failed adding new tag to database")
	}
	slog.Debug("Adding tag", "added_tag", tag)
	return tag, nil
}

// Let user update one of their tags (replaces).
func (s *Service) UpdateTag(userId uint, tagId uint, tr domain.TagAddRequest) error {
	if tr.Name == "" {
		return errors.New("tag must have a name")
	}
	tag := entity.Tag{Name: tr.Name, Color: tr.Color, BgColor: tr.BgColor}
	res := s.db.Where("id = ? AND user_id = ?", tagId, userId).Updates(&tag)
	if res.Error != nil {
		slog.Error("Error updating tag in database", "error", res.Error.Error())
		return errors.New("failed updating tag in database")
	}
	if res.RowsAffected == 0 {
		slog.Error("updateTag: Zero rows affected.. tag likely does not exist", "tag_id", tagId, "user_id", userId)
		return errors.New("tag does not exist")
	}
	slog.Debug("updateTag:", "updated_tag", tag)
	return nil
}

// Let user delete their own tag.
func (s *Service) DeleteTag(userId uint, tagId uint) error {
	if tagId == 0 {
		return errors.New("no tag id provided")
	}
	slog.Debug("deleteTag:", "tag_id", tagId, "user_id", userId)
	// Select("Watched") so relations in watched_tags table are removed too.
	// ID is passed in the .Delete param so the .Select call can do it's job (relies on the primary key).
	res := s.db.Unscoped().Where("id = ? AND user_id = ?", tagId, userId).Select("Watched").Delete(&entity.Tag{GormModel: dbmodel.GormModel{ID: tagId}})
	if res.Error != nil {
		slog.Error("deleteTag: Error deleting tag from database", "error", res.Error.Error(), "tag_id", tagId, "user_id", userId)
		return errors.New("failed deleting tag from database")
	}
	if res.RowsAffected == 0 {
		slog.Error("deleteTag: Zero rows affected.. tag must not exist for user", "tag_id", tagId, "user_id", userId)
		return errors.New("tag does not exist")
	}
	return nil
}

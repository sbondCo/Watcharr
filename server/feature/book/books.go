package book

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/image"
	"github.com/sbondCo/Watcharr/media/openlibrary"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db               *gorm.DB
	openLibrary      *openlibrary.OpenLibrary
	activityProvider domain.ActivityAddProvider
}

func NewService(db *gorm.DB, openLibrary *openlibrary.OpenLibrary, activityProvider domain.ActivityAddProvider) *Service {
	return &Service{
		db,
		openLibrary,
		activityProvider,
	}
}

// Cache(save) book to our table
func (s *Service) saveBook(c *entity.Book, onlyUpdate bool) error {
	slog.Info("Saving book to db", "olid", c.OLID, "title", c.Title)
	if c.OLID == "" || c.Title == "" {
		slog.Error("savebook: content missing id or title!", "olid", c.OLID, "title", c.Title)
		return errors.New("book missing id or title")
	}
	coverUrl := fmt.Sprintf("%s/w/olid/%s-M.jpg", openlibrary.CoverBaseUrl, c.OLID)
	p, err := image.NewSaver(s.db, "books",
		image.ValidateOptions{
			// To avoid losing quality, we want to keep png format
			// for book covers.
			ToFormat: image.ValidateAllowedFormatPNG,
		},
	).DownloadAndInsertFromUrl(coverUrl)
	if err != nil {
		slog.Error("savebook: Failed to cache book cover.", "error", err)
	} else {
		slog.Debug("savebook: Cached book cover", "p", p)
		c.CoverID = &p.ID
	}

	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = s.db.Model(&entity.Book{}).Where("ol_id = ?", c.OLID).Updates(c)
		if res.Error != nil {
			slog.Error("savebook: Error updating book in database", "error", res.Error.Error())
			return errors.New("failed to update cached book in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = s.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "ol_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"isbn",
				"title",
				"storyline",
				"cover_id",
				"release_date",
				"rating_average",
				"rating_count",
				"genres",
				"author_names",
				"author_ids",
			}),
		}).Create(&c)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if res.Error != gorm.ErrDuplicatedKey {
				slog.Error("saveBook: Error creating book in database", "error", res.Error.Error())
				return errors.New("failed to cache book in database")
			}
		}
	}
	return nil
}

func (s *Service) cacheBook(b entity.Book, onlyUpdate bool) (entity.Book, error) {
	slog.Debug("cacheBook", "book_details", s)
	err := s.saveBook(&b, onlyUpdate)
	if err != nil {
		slog.Error("cacheBook: Failed to save book!", "error", err)
		return entity.Book{}, errors.New("failed to save book")
	}
	return b, nil
}

func (s *Service) GetOrCache(olid string) (entity.Book, error) {
	var book entity.Book
	s.db.Where("ol_id = ?", olid).Find(&book)

	// Create book if not found from our db
	if book == (entity.Book{}) {
		slog.Debug("GetOrCache: book not in db, fetching...")

		resp, err := s.openLibrary.GetBookDetails(olid)
		if err != nil {
			slog.Error("GetOrCache: content api request failed", "error", err)
			return book, errors.New("failed to find requested books")
		}

		book, err = s.cacheBook(resp, false)
		if err != nil {
			slog.Error("GetOrCache: failed to cache book",
				"olid", olid,
				"err", err)
			return book, errors.New("failed to cache content")
		}
	}

	return book, nil
}

package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/game"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// For storing cached games, so we can serve the basic local data for watched list to work
type Game struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UpdatedAt time.Time `json:"updatedAt"`
	IgdbID    int       `json:"igdbId" gorm:"uniqueIndex;not null"`
	Name      string    `json:"name"`
	CoverID   string    `json:"coverId"`
	Summary   string    `json:"summary"`
	Storyline string    `json:"storyline"`
	// First release date
	ReleaseDate *time.Time `json:"releaseDate,omitempty"`
	Rating      float64    `json:"rating"`
	RatingCount int        `json:"ratingCount"`
	Status      int        `json:"status"`
	Category    int        `json:"category"`
	// Arrays turned to strings that may be useful
	GameModes string `json:"gameModes"`
	Genres    string `json:"genres"`
	Platforms string `json:"platforms"`
	// Id to poster image row (cached game cover)
	PosterID *uint  `json:"-"`
	Poster   *Image `json:"poster,omitempty"`
}

type PlayedAddRequest struct {
	Status WatchedStatus `json:"status"`
	Rating float64       `json:"rating" binding:"max=10"`
	IgdbID int           `json:"igdbId" binding:"required"`
}

// Cache(save) game to our table
func saveGame(db *gorm.DB, c *Game, onlyUpdate bool) error {
	slog.Info("Saving game to db", "id", c.IgdbID, "name", c.Name)
	if c.IgdbID == 0 || c.Name == "" {
		slog.Error("saveGame: content missing id or name!", "id", c.IgdbID, "name", c.Name)
		return errors.New("game missing id or title")
	}
	if c.CoverID != "" {
		p, err := downloadAndInsertImage(db, "https://images.igdb.com/igdb/image/upload/t_cover_big/"+c.CoverID+".png", "games")
		if err != nil {
			slog.Error("saveGame: Failed to cache game cover.", "error", err)
		} else {
			slog.Debug("saveGame: Cached game cover", "p", p)
			c.PosterID = &p.ID
		}
	}
	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = db.Model(&Game{}).Where("igdb_id = ?", c.IgdbID).Updates(c)
		if res.Error != nil {
			slog.Error("saveGame: Error updating game in database", "error", res.Error.Error())
			return errors.New("failed to update cached game in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "igdb_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"cover_id",
				"summary",
				"storyline",
				"release_date",
				"rating",
				"rating_count",
				"status",
				"game_modes",
				"genres",
			}),
		}).Create(&c)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if res.Error != gorm.ErrDuplicatedKey {
				slog.Error("saveGame: Error creating game in database", "error", res.Error.Error())
				return errors.New("failed to cache game in database")
			}
		}
	}
	return nil
}

func cacheGame(db *gorm.DB, g game.GameDetailsBasicResponse, onlyUpdate bool) (Game, error) {
	slog.Debug("cacheGame", "game_details", g)
	var (
		gameModes string
		genres    string
		platforms string
	)
	if len(g.GameModes) > 0 {
		for _, v := range g.GameModes {
			gameModes += v.Name + "|"
		}
	}
	if len(g.Genres) > 0 {
		for _, v := range g.Genres {
			genres += v.Name + "|"
		}
	}
	if len(g.Platforms) > 0 {
		for _, v := range g.Platforms {
			platforms += v.Name + "|"
		}
	}
	c := Game{
		IgdbID:      g.ID,
		Name:        g.Name,
		CoverID:     g.Cover.ImageID,
		Summary:     g.Summary,
		Storyline:   g.Storyline,
		ReleaseDate: &g.FirstReleaseDate.Time,
		Rating:      (g.Rating),
		RatingCount: g.RatingCount,
		Status:      g.Status,
		Category:    g.Category,
		GameModes:   gameModes,
		Genres:      genres,
		Platforms:   platforms,
	}
	err := saveGame(db, &c, onlyUpdate)
	if err != nil {
		slog.Error("cacheGame: Failed to save game!", "error", err)
		return Game{}, errors.New("failed to save game")
	}
	return c, nil
}

// For adding/updating played games, we will reuse methods defined in watched.go where easily possible.

func addPlayed(db *gorm.DB, igdb *game.IGDB, userId uint, ar PlayedAddRequest, at ActivityType) (Watched, error) {
	slog.Debug("Adding played item", "userId", userId, "igdbId", ar.IgdbID)

	var game Game
	db.Where("igdb_id = ?", ar.IgdbID).Find(&game)

	// Create game if not found from our db
	if game == (Game{}) {
		slog.Debug("Game not in db, fetching...")

		resp, err := igdb.GameDetailsBasic(strconv.Itoa(ar.IgdbID))
		if err != nil {
			slog.Error("addPlayed content tmdb api request failed", "error", err)
			return Watched{}, errors.New("failed to find requested games")
		}

		game, err = cacheGame(db, resp, false)
		if err != nil {
			slog.Error("addPlayed failed to cache game", "igdb_id", ar.IgdbID, "err", err)
			return Watched{}, errors.New("failed to cache content")
		}
	}
	// Error if content has no id
	if game.ID == 0 {
		return Watched{}, errors.New("failed to find game by id")
	}
	// Create watched entry in db
	if ar.Status == "" {
		ar.Status = FINISHED
	}
	watched := Watched{Status: ar.Status, Rating: ar.Rating, UserID: userId, GameID: &game.ID}
	res := db.Create(&watched)
	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			res = db.Model(&Watched{}).Unscoped().Preload("Activity").Where("user_id = ? AND game_id = ?", userId, watched.GameID).Take(&watched)
			if res.Error != nil {
				return Watched{}, errors.New("content already on watched list. errored checking for soft deleted record")
			}
			if watched.DeletedAt.Time.IsZero() {
				return Watched{}, errors.New("content already on watched list")
			} else {
				slog.Info("addPlayed: Watched list item for this content exists as soft deleted record.. attempting to restore")
				res = db.Model(&Watched{}).Unscoped().Where("user_id = ? AND game_id = ?", userId, watched.GameID).Updates(map[string]interface{}{"status": ar.Status, "rating": ar.Rating, "deleted_at": nil})
				watched.Status = ar.Status
				watched.Rating = ar.Rating
				if res.Error != nil {
					slog.Error("addPlayed: Failed to restore soft deleted watch list item", "error", res.Error)
					return Watched{}, errors.New("content already on watched list. errored removing soft delete timestamp")
				}
			}
		} else {
			slog.Error("Error adding watched content to database", "error", res.Error.Error())
			return Watched{}, errors.New("failed adding content to database")
		}
	}
	slog.Debug("Added watched list item", "item", watched)

	var activity Activity
	activityJson, err := json.Marshal(map[string]interface{}{"status": ar.Status, "rating": ar.Rating})
	if err != nil {
		slog.Error("Failed to marshal json for data in ADD_WATCHED activity request, adding without data", "error", err.Error())
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: watched.ID, Type: at})
	} else {
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: watched.ID, Type: at, Data: string(activityJson)})
	}
	watched.Activity = append(watched.Activity, activity)
	watched.Game = &game
	return watched, nil
}

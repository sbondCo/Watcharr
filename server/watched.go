package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"

	"gorm.io/gorm"
)

type WatchedStatus string

const (
	FINISHED WatchedStatus = "FINISHED"
	WATCHING WatchedStatus = "WATCHING"
	PLANNED  WatchedStatus = "PLANNED"
	HOLD     WatchedStatus = "ONHOLD"
	DROPPED  WatchedStatus = "DROPPED"
)

type Watched struct {
	GormModel
	Status         WatchedStatus   `json:"status"`
	Rating         int8            `json:"rating"`
	Thoughts       string          `json:"thoughts"`
	UserID         uint            `json:"-" gorm:"uniqueIndex:usernctnidx"`
	ContentID      int             `json:"-" gorm:"uniqueIndex:usernctnidx"`
	Content        Content         `json:"content"`
	Activity       []Activity      `json:"activity"`
	WatchedSeasons []WatchedSeason `json:"watchedSeasons,omitempty"` // For shows
}

type WatchedAddRequest struct {
	Status      WatchedStatus `json:"status"`
	Rating      int8          `json:"rating" binding:"max=10"`
	ContentID   int           `json:"contentId" binding:"required"`
	ContentType ContentType   `json:"contentType" binding:"required,oneof=movie tv"`
}

type WatchedUpdateRequest struct {
	Status         WatchedStatus `json:"status" binding:"required_without_all=Rating Thoughts RemoveThoughts"`
	Rating         int8          `json:"rating" binding:"max=10,required_without_all=Status Thoughts RemoveThoughts"`
	Thoughts       string        `json:"thoughts" binding:"required_without_all=Status Rating RemoveThoughts"`
	RemoveThoughts bool          `json:"removeThoughts"`
}

type WatchedUpdateResponse struct {
	NewActivity Activity `json:"newActivity"`
}

type WatchedRemoveResponse struct {
	NewActivity Activity `json:"newActivity"`
}

func getWatched(db *gorm.DB, userId uint) []Watched {
	watched := new([]Watched)
	res := db.Model(&Watched{}).Preload("Content").Preload("Activity").Preload("WatchedSeasons").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}

// Get another users **public** watchlist.
func getPublicWatched(db *gorm.DB, userId uint, username string) ([]Watched, error) {
	slog.Debug("getPublicWatched running", "user_id", userId, "username", username)
	// First we need to make sure the users list is public
	user := new(User)
	// Figure we require knowlege of the users id and name to make it
	// harder to just type in random ids and see someones list.. dunno
	// if this is a thing we need but its here.. for now at least.
	res := db.Where("id = ? AND username = ?", userId, username).Take(&user)
	if res.Error != nil {
		slog.Error("Failed to get user for getPublicWatched request")
		return []Watched{}, errors.New("failed to check privacy settings")
	}
	if user.Private != nil && *user.Private {
		slog.Error("getPublicWatched attempted to get a private list")
		return []Watched{}, errors.New("this watched list is private")
	}
	// Now we know the user is public, return their list
	watched := new([]Watched)
	res = db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched, nil
}

func addWatched(db *gorm.DB, userId uint, ar WatchedAddRequest, at ActivityType) (Watched, error) {
	slog.Debug("Adding watched item", "userId", userId, "contentType", ar.ContentType, "contentId", ar.ContentID)

	var content Content
	db.Where("type = ? AND tmdb_id = ?", ar.ContentType, ar.ContentID).Find(&content)

	// Create content if not found from our db
	if content == (Content{}) {
		slog.Debug("Content not in db, fetching...")

		resp, err := tmdbAPIRequest("/"+string(ar.ContentType)+"/"+strconv.Itoa(ar.ContentID), map[string]string{})
		if err != nil {
			slog.Error("addWatched content tmdb api request failed", "error", err)
			return Watched{}, errors.New("failed to find requested media")
		}

		if ar.ContentType == "movie" {
			c := new(TMDBMovieDetails)
			err := json.Unmarshal([]byte(resp), &c)
			if err != nil {
				slog.Error("Failed to unmarshal movie details", "error", err)
				return Watched{}, errors.New("failed to process movie details response")
			}
			content, err = cacheContentMovie(db, *c, false)
			if err != nil {
				slog.Error("addWatched failed to cache movie content", "content_id", ar.ContentID, "err", err)
				return Watched{}, errors.New("failed to cache content")
			}
		} else {
			c := new(TMDBShowDetails)
			err := json.Unmarshal(resp, &c)
			if err != nil {
				slog.Error("Failed to unmarshal tv details", "error", err)
				return Watched{}, errors.New("failed to process tv details response")
			}
			content, err = cacheContentTv(db, *c, false)
			if err != nil {
				slog.Error("addWatched failed to cache tv content", "content_id", ar.ContentID, "err", err)
				return Watched{}, errors.New("failed to cache content")
			}
		}
	}
	// Error if content has no id
	if content.ID == 0 {
		return Watched{}, errors.New("failed to find content id")
	}
	// Create watched entry in db
	if ar.Status == "" {
		// Set default status for when content is added by
		// rating it instead of giving status first.
		if ar.ContentType == "movie" {
			ar.Status = FINISHED
		} else {
			ar.Status = WATCHING
		}
	}
	watched := Watched{Status: ar.Status, Rating: ar.Rating, UserID: userId, ContentID: content.ID}
	res := db.Create(&watched)
	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			res = db.Model(&Watched{}).Unscoped().Preload("Activity").Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Take(&watched)
			if res.Error != nil {
				return Watched{}, errors.New("content already on watched list. errored checking for soft deleted record")
			}
			if watched.DeletedAt.Time.IsZero() {
				return Watched{}, errors.New("content already on watched list")
			} else {
				slog.Info("addWatched: Watched list item for this content exists as soft deleted record.. attempting to restore")
				res = db.Model(&Watched{}).Unscoped().Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Updates(map[string]interface{}{"status": ar.Status, "rating": ar.Rating, "deleted_at": nil})
				watched.Status = ar.Status
				watched.Rating = ar.Rating
				if res.Error != nil {
					slog.Error("addWatched: Failed to restore soft deleted watch list item", "error", res.Error)
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
	watched.Content = content
	return watched, nil
}

// this method is too ugly to look at please make him look better, future irhm
func updateWatched(db *gorm.DB, userId uint, id uint, ar WatchedUpdateRequest) (WatchedUpdateResponse, error) {
	slog.Debug("UpdateWatched", "request_data", ar)
	upwat := Watched{}
	res := db.Model(&Watched{}).Where("id = ? AND user_id = ?", id, userId).Take(&upwat)
	if res.Error != nil {
		slog.Error("Watched entry update failed:", "id", id, "error", res.Error.Error())
		return WatchedUpdateResponse{}, errors.New("failed to update watched entry")
	}
	originalThoughts := upwat.Thoughts
	if ar.Rating != 0 {
		upwat.Rating = ar.Rating
	}
	if ar.Status != "" {
		upwat.Status = ar.Status
	}
	if ar.Thoughts != "" {
		upwat.Thoughts = ar.Thoughts
	}
	if ar.RemoveThoughts {
		upwat.Thoughts = ""
	}
	res = db.Save(upwat)
	if res.RowsAffected <= 0 {
		return WatchedUpdateResponse{}, errors.New("no watched entry found")
	}
	addedActivity := Activity{}
	if ar.Rating != 0 {
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: RATING_CHANGED, Data: strconv.Itoa(int(ar.Rating))})
	}
	if ar.Status != "" {
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: STATUS_CHANGED, Data: string(ar.Status)})
	}
	if ar.Thoughts != "" {
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: THOUGHTS_CHANGED})
	}
	if ar.RemoveThoughts {
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: THOUGHTS_REMOVED, Data: originalThoughts})
	}
	return WatchedUpdateResponse{NewActivity: addedActivity}, nil
}

func removeWatched(db *gorm.DB, userId uint, id uint) (WatchedRemoveResponse, error) {
	slog.Debug("Removing watched item:", "id", id, "user_id", userId)
	// Our model has a deleted_at field, which will make gorm do a soft delete.
	// Since other tables (eg activities) will link their rows to a watched_id, it's best to soft
	// delete, so if user restores watched item they still have activity for example (also so
	// someone else wont get other users activity if auto increment gives them the same watched id).
	res := db.Model(&Watched{}).Where("id = ? AND user_id = ?", id, userId).Delete(&Watched{})
	if res.Error != nil {
		slog.Error("Removing watched entry failed", "id", id, "error", res.Error.Error())
		return WatchedRemoveResponse{}, errors.New("failed to remove watched entry")
	}
	if res.RowsAffected <= 0 {
		return WatchedRemoveResponse{}, errors.New("no watched entry found")
	}
	addedActivity, _ := addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: REMOVED_WATCHED})
	return WatchedRemoveResponse{NewActivity: addedActivity}, nil
}

func download(url string, outf string) (err error) {
	slog.Debug("Attempting to download file", "url", url, "outf", outf)

	// Create the file
	out, err := os.Create(outf)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path.Dir(outf), 0764)
			if err != nil {
				return err
			}
			// If dirs made, try making file again
			out, err = os.Create(outf)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

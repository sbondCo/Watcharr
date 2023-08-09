package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

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
	Status    WatchedStatus `json:"status"`
	Rating    int8          `json:"rating"`
	Thoughts  string        `json:"thoughts"`
	UserID    uint          `json:"-" gorm:"uniqueIndex:usernctnidx"`
	ContentID int           `json:"-" gorm:"uniqueIndex:usernctnidx"`
	Content   Content       `json:"content"`
	Activity  []Activity    `json:"activity"`
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
	res := db.Model(&Watched{}).Preload("Content").Preload("Activity").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}

func addWatched(db *gorm.DB, userId uint, ar WatchedAddRequest) (Watched, error) {
	println(ar.ContentType, ar.ContentID)

	var content Content
	db.Where("tmdb_id = ?", ar.ContentID).Find(&content)

	// Create content if not found from our db
	if content == (Content{}) {
		println("Content not in db, fetching...")

		resp, err := tmdbAPIRequest("/"+string(ar.ContentType)+"/"+strconv.Itoa(ar.ContentID), map[string]string{})
		if err != nil {
			fmt.Printf("addWatched tmdb api request failed: %+v", err)
			return Watched{}, errors.New("failed to find requested media")
		}

		var (
			id               int
			title            string
			overview         string
			posterPath       string
			releaseDate      time.Time
			popularity       float32
			voteAverage      float32
			voteCount        uint32
			imdbID           string
			status           string
			budget           uint32
			revenue          uint32
			runtime          uint32
			numberOfEpisodes uint32
			numberOfSeasons  uint32
		)
		var dateFormat = "2006-01-02"
		// Get details from movie/show response and fill out needed vars
		if ar.ContentType == "movie" {
			content := new(TMDBMovieDetails)
			err = json.Unmarshal([]byte(resp), &content)
			if err != nil {
				println("Failed to unmarshal movie details:", err)
				return Watched{}, errors.New("failed to process movie details response")
			}
			id = content.ID
			overview = content.Overview
			posterPath = content.PosterPath
			title = content.Title
			releaseDate, err = time.Parse(dateFormat, content.ReleaseDate)
			if err != nil {
				println("Failed to parse movie release date:", err)
			}
			popularity = content.Popularity
			voteAverage = content.VoteAverage
			voteCount = content.VoteCount
			imdbID = content.ImdbID
			status = content.Status
			budget = content.Budget
			revenue = content.Revenue
			runtime = content.Runtime
		} else {
			content := new(TMDBShowDetails)
			err = json.Unmarshal(resp, &content)
			if err != nil {
				println("Failed to unmarshal show details:", err)
				return Watched{}, errors.New("failed to process show details response")
			}
			id = content.ID
			overview = content.Overview
			posterPath = content.PosterPath
			title = content.Name
			releaseDate, err = time.Parse(dateFormat, content.FirstAirDate)
			if err != nil {
				println("Failed to parse tv release date:", err)
			}
			popularity = content.Popularity
			voteAverage = content.VoteAverage
			voteCount = content.VoteCount
			status = content.Status
			if len(content.EpisodeRunTime) > 0 {
				runtime = uint32(content.EpisodeRunTime[0])
			}
			numberOfEpisodes = content.NumberOfEpisodes
			numberOfSeasons = content.NumberOfSeasons
		}
		// Save the content in our db
		println("id, etc:", id, title, overview, posterPath, "<-- end")
		if id == 0 || title == "" {
			println("addWatched, returned content missing id or title!", id, title)
			return Watched{}, errors.New("content response missing id or title")
		}
		content = Content{
			TmdbID:           id,
			Title:            title,
			Overview:         overview,
			PosterPath:       posterPath,
			Type:             ar.ContentType,
			ReleaseDate:      releaseDate,
			Popularity:       popularity,
			VoteAverage:      voteAverage,
			VoteCount:        voteCount,
			ImdbID:           imdbID,
			Status:           status,
			Budget:           budget,
			Revenue:          revenue,
			Runtime:          runtime,
			NumberOfEpisodes: numberOfEpisodes,
			NumberOfSeasons:  numberOfSeasons,
		}
		res := db.Create(&content)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if !strings.Contains(res.Error.Error(), "UNIQUE") {
				println("Error creating content in database:", res.Error.Error())
				return Watched{}, errors.New("failed to cache content in database")
			}
		}
		// If row created, download the image
		if res.RowsAffected > 0 {
			err := download("https://image.tmdb.org/t/p/w500"+posterPath, path.Join("./data/img", posterPath))
			if err != nil {
				println("Failed to download content image!", err.Error())
			}
		}
	}
	// Error if content has no id
	if content.ID == 0 {
		return Watched{}, errors.New("failed to find content id")
	}
	// Create watched entry in db
	if ar.Status == "" {
		ar.Status = WATCHING
	}
	watched := Watched{Status: ar.Status, Rating: ar.Rating, UserID: userId, ContentID: content.ID}
	res := db.Create(&watched)
	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "UNIQUE") {
			res = db.Model(&Watched{}).Unscoped().Preload("Activity").Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Take(&watched)
			if res.Error != nil {
				return Watched{}, errors.New("content already on watched list. errored checking for soft deleted record")
			}
			if watched.DeletedAt.Time.IsZero() {
				return Watched{}, errors.New("content already on watched list")
			} else {
				println("Watched list item exists as soft deleted record.. attempting to restore")
				res = db.Model(&Watched{}).Unscoped().Where("user_id = ? AND content_id = ?", userId, watched.ContentID).Updates(map[string]interface{}{"status": ar.Status, "rating": ar.Rating, "deleted_at": nil})
				watched.Status = ar.Status
				watched.Rating = ar.Rating
				if res.Error != nil {
					return Watched{}, errors.New("content already on watched list. errored removing soft delete timestamp")
				}
			}
		} else {
			println("Error adding watched content to database:", res.Error.Error())
			return Watched{}, errors.New("failed adding content to database")
		}
	}
	fmt.Printf("%+v\n", watched)

	var activity Activity
	activityJson, err := json.Marshal(map[string]interface{}{"status": ar.Status, "rating": ar.Rating})
	if err != nil {
		println("Failed to marshal json for data in ADD_WATCHED activity request, adding without data", err.Error())
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: watched.ID, Type: ADDED_WATCHED})
	} else {
		activity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: watched.ID, Type: ADDED_WATCHED, Data: string(activityJson)})
	}
	watched.Activity = append(watched.Activity, activity)
	watched.Content = content
	return watched, nil
}

func updateWatched(db *gorm.DB, userId uint, id uint, ar WatchedUpdateRequest) (WatchedUpdateResponse, error) {
	println("UpdateWatched", ar.Rating, ar.Status)
	// ugly but it works so no complaining
	upwat := map[string]interface{}{}
	if ar.Rating != 0 {
		upwat["rating"] = ar.Rating
	}
	if ar.Status != "" {
		upwat["status"] = ar.Status
	}
	if ar.Thoughts != "" {
		upwat["thoughts"] = ar.Thoughts
	}
	if ar.RemoveThoughts {
		upwat["thoughts"] = ""
	}
	res := db.Model(&Watched{}).Where("id = ? AND user_id = ?", id, userId).Updates(upwat)
	if res.Error != nil {
		println("Watched entry update failed:", id, res.Error.Error())
		return WatchedUpdateResponse{}, errors.New("failed to update watched entry")
	}
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
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: THOUGHTS_REMOVED, Data: ar.Thoughts})
	}
	return WatchedUpdateResponse{NewActivity: addedActivity}, nil
}

func removeWatched(db *gorm.DB, userId uint, id uint) (WatchedRemoveResponse, error) {
	println("Removing watched item:", id, "for user", userId)
	// Our model has a deleted_at field, which will make gorm do a soft delete.
	// Since other tables (eg activities) will link their rows to a watched_id, it's best to soft
	// delete, so if user restores watched item they still have activity for example (also so
	// someone else wont get other users activity if auto increment gives them the same watched id).
	res := db.Model(&Watched{}).Where("id = ? AND user_id = ?", id, userId).Delete(&Watched{})
	if res.Error != nil {
		println("Removing watched entry failed:", id, res.Error.Error())
		return WatchedRemoveResponse{}, errors.New("failed to remove watched entry")
	}
	if res.RowsAffected <= 0 {
		return WatchedRemoveResponse{}, errors.New("no watched entry found")
	}
	addedActivity, _ := addActivity(db, userId, ActivityAddRequest{WatchedID: id, Type: REMOVED_WATCHED})
	return WatchedRemoveResponse{NewActivity: addedActivity}, nil
}

func download(url string, outf string) (err error) {
	println("Attempting to download file from", url, "to", outf)

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

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
	Status    *WatchedStatus `json:"status"`
	Rating    *int8          `json:"rating"`
	UserID    uint           `json:"-" gorm:"uniqueIndex:usernctnidx"`
	ContentID int            `json:"-" gorm:"uniqueIndex:usernctnidx"`
	Content   Content        `json:"content"`
}

type WatchedAddRequest struct {
	Status      WatchedStatus `json:"status"`
	Rating      int8          `json:"rating" binding:"max=5"`
	ContentID   int           `json:"contentId" binding:"required"`
	ContentType ContentType   `json:"contentType" binding:"required,oneof=movie tv"`
}

type WatchedUpdateRequest struct {
	Status WatchedStatus `json:"status" binding:"required_without=Rating"`
	Rating int8          `json:"rating" binding:"max=10,required_without=Status"`
}

func getWatched(db *gorm.DB, userId uint) []Watched {
	watched := new([]Watched)
	res := db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}

func addWatched(db *gorm.DB, userId uint, ar WatchedAddRequest) (Watched, error) {
	println(ar.ContentType, ar.ContentID)

	resp, err := tmdbAPIRequest("/"+string(ar.ContentType)+"/"+strconv.Itoa(ar.ContentID), map[string]string{})
	if err != nil {
		fmt.Printf("addWatched tmdb api request failed: %+v", err)
		return Watched{}, errors.New("failed to find requested media")
	}

	var (
		id         int
		title      string
		overview   string
		posterPath string
	)
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
	}
	// Save the content in our db
	println("id, etc:", id, title, overview, posterPath, "<-- end")
	if id == 0 || title == "" {
		println("addWatched, returned content missing id or title!", id, title)
		return Watched{}, errors.New("content response missing id or title")
	}
	content := Content{ID: id, Title: title, Overview: overview, PosterPath: posterPath, Type: ar.ContentType}
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

	// Create watched entry in db
	if ar.Status == "" {
		ar.Status = WATCHING
	}
	watched := Watched{Status: &ar.Status, Rating: &ar.Rating, UserID: userId, ContentID: id}
	res = db.Create(&watched)
	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "UNIQUE") {
			return Watched{}, errors.New("content already on watched list")
		}
		println("Error adding watched content to database:", res.Error.Error())
		return Watched{}, errors.New("failed adding content to database")
	}
	println(res.RowsAffected)
	fmt.Printf("%+v\n", watched)

	watched.Content = content
	return watched, nil
}

func updateWatched(db *gorm.DB, userId uint, id uint, ar WatchedUpdateRequest) (bool, error) {
	res := db.Model(&Watched{}).Where("id = ? AND user_id = ?", id, userId).Updates(Watched{Rating: &ar.Rating, Status: &ar.Status})
	if res.Error != nil {
		println("Watched entry update failed:", id, res.Error.Error())
		return false, errors.New("failed to update watched entry")
	}
	if res.RowsAffected <= 0 {
		return false, errors.New("no watched entry found")
	}
	return true, nil
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

package main

import (
	"errors"
	"fmt"
	"io"
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
)

type Watched struct {
	GormModel
	Status    WatchedStatus `json:"status"`
	Rating    int8          `json:"rating"`
	UserID    uint          `json:"-"`
	ContentID int           `json:"-"`
	Content   Content       `json:"content"`
}

type WatchedAddRequest struct {
	Status      WatchedStatus `json:"status"`
	Rating      int8          `json:"rating"`
	ContentID   int           `json:"contentId" binding:"required"`
	ContentType ContentType   `json:"contentType" binding:"required,oneof=movie tv"`
}

func getWatched(db *gorm.DB, userId uint) []Watched {
	watched := new([]Watched)
	res := db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}

func addWatched(db *gorm.DB, userId uint, ar WatchedAddRequest) (bool, error) {
	println(ar.ContentType, ar.ContentID)

	content := new(TMDBMovieDetails)
	err := tmdbRequest("/"+string(ar.ContentType)+"/"+strconv.Itoa(ar.ContentID), map[string]string{}, &content)
	if err != nil {
		return false, errors.New("failed to find requested media")
	}

	// // Save the content in our db
	// res := db.Create(&content)
	// if res.Error != nil {
	// 	// Error if anything but unique contraint error
	// 	if !strings.Contains(res.Error.Error(), "UNIQUE") {
	// 		println("Error creating content in database:", res.Error.Error())
	// 		return false, errors.New("failed to cache content in database")
	// 	}
	// }
	// println(res.RowsAffected)
	// // If row created, download the image
	// if res.RowsAffected > 0 {
	// 	err := download("https://image.tmdb.org/t/p/w500"+content.PosterPath, path.Join("./data/img", content.PosterPath))
	// 	if err != nil {
	// 		println("Failed to download content image!", err.Error())
	// 	}
	// }

	// // Create watched entry in db
	// watched := Watched{Status: "d", Rating: 5, UserID: userId, ContentID: content.ID}
	// res = db.Create(&watched)
	// if res.Error != nil {
	// 	println("Error adding watched content to database:", res.Error.Error())
	// 	return false, errors.New("failed adding content to database")
	// }
	// println(res.RowsAffected)
	// fmt.Printf("%+v\n", watched)

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

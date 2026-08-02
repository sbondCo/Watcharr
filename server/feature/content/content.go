package content

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
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// inmemory content cache
var ContentStore = gocache.New(time.Hour*24, time.Minute)

// Download file over http (used for downloading poster images)
// url - The remote file url.
// outf - Where should we store the downloaded file.
// force - Should we overwrite an existing file? If false, existing files will be skipped.
func download(url string, outf string, force bool) (err error) {
	slog.Debug("download: Attempting to download file", "url", url, "outf", outf, "force", force)
	// If not forced, skip call if file already exists to save unnecessary requests.
	if !force {
		if _, err := os.Stat(outf); !errors.Is(err, os.ErrNotExist) {
			slog.Debug("download: Skipping file, it already exists locally.", "outf", outf, "error", err)
			return nil
		} else {
			slog.Debug("download: Continuing to download file, it does not already exist.", "outf", outf, "error", err)
		}
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("download: Failed to make request.", "outf", outf, "error", err)
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		slog.Error("download: Request failed. Non OK response.", "outf", outf, "status", resp.Status, "error", err)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(outf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("download: Failed to create out file, trying to recover by ensuring directories exist.", "outf", outf)
			err = os.MkdirAll(path.Dir(outf), 0764)
			if err != nil {
				slog.Error("download: Failed to create dir(s) in recovery attempt.", "outf", outf, "error", err)
				return err
			}
			// If dirs made, try making file again
			out, err = os.Create(outf)
			if err != nil {
				slog.Error("download: Failed to create out file again in recovery attempt.", "outf", outf, "error", err)
				return err
			}
			slog.Info("download: recovered by creating dir(s).", "outf", outf)
		} else {
			slog.Error("download: Failed to create out file. No known recovery path possible.", "outf", outf, "error", err)
			return err
		}
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		slog.Error("download: Failed to write file to our file.", "outf", outf, "error", err)
		return err
	}

	slog.Debug("download: Successfully downloaded file", "outf", outf)
	return nil
}

type Service struct {
	db   *gorm.DB
	tmdb *tmdb.TMDB
}

func NewService(db *gorm.DB, tmdb *tmdb.TMDB) *Service {
	return &Service{
		db,
		tmdb,
	}
}

// onlyUpdate - If we should only update existing row if exists, or false to create/update if not exist.
func (s *Service) saveContent(c *entity.Content, onlyUpdate bool) error {
	slog.Info("Saving content to db", "id", c.TmdbID, "title", c.Title)
	if c.TmdbID == 0 || c.Title == "" || c.Type == "" {
		slog.Error("saveContent: content missing id, title or type!", "id", c.TmdbID, "title", c.Title, "type", c.Type)
		return errors.New("content missing id or title")
	}
	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = s.db.Model(&entity.Content{}).Where("type = ? AND tmdb_id = ?", c.Type, c.TmdbID).Updates(c)
		if res.Error != nil {
			slog.Error("saveContent: Error updating content in database", "error", res.Error.Error())
			return errors.New("failed to update cached content in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = s.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tmdb_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title",
				"poster_path",
				"overview",
				"release_date",
				"popularity",
				"vote_average",
				"vote_count",
				"imdb_id",
				"status",
				"budget",
				"revenue",
				"runtime",
				"number_of_episodes",
				"number_of_seasons",
			}),
		}).Create(&c)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if res.Error != gorm.ErrDuplicatedKey {
				slog.Error("saveContent: Error creating content in database", "error", res.Error.Error())
				return errors.New("failed to cache content in database")
			}
		}
	}
	// If row created, download the image
	if res.RowsAffected > 0 {
		slog.Debug("saveContent: Downloading poster.")
		err := download(
			"https://image.tmdb.org/t/p/w500"+c.PosterPath,
			path.Join(config.DataPath, "img", c.PosterPath),
			false,
		)
		if err != nil {
			slog.Error("saveContent: Failed to download content image!", "error", err.Error())
		}
	}
	return nil
}

func (s *Service) CacheContentTv(content tmdb.TMDBShowDetails, onlyUpdate bool) (entity.Content, error) {
	slog.Debug("cacheContentTv", "content", content)
	var (
		releaseDate time.Time
		runtime     uint32
	)
	var dateFormat = "2006-01-02"
	releaseDate, err := time.Parse(dateFormat, content.FirstAirDate)
	if err != nil {
		slog.Error("Failed to parse tv release date", "error", err)
	}
	if len(content.EpisodeRunTime) > 0 {
		runtime = uint32(content.EpisodeRunTime[0])
	}

	c := entity.Content{
		TmdbID:           content.ID,
		Title:            content.Name,
		Overview:         content.Overview,
		PosterPath:       content.PosterPath,
		Type:             entity.SHOW,
		ReleaseDate:      &releaseDate,
		Popularity:       content.Popularity,
		VoteAverage:      content.VoteAverage,
		VoteCount:        content.VoteCount,
		Status:           content.Status,
		Runtime:          runtime,
		NumberOfEpisodes: content.NumberOfEpisodes,
		NumberOfSeasons:  content.NumberOfSeasons,
	}

	err = s.saveContent(&c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentTv: Failed to save content!", "error", err)
		return entity.Content{}, errors.New("failed to save content")
	}

	return c, nil
}

func (s *Service) CacheContentMovie(content tmdb.TMDBMovieDetails, onlyUpdate bool) (entity.Content, error) {
	var (
		releaseDate time.Time
	)
	var dateFormat = "2006-01-02"
	// Get details from movie/show response and fill out needed vars
	releaseDate, err := time.Parse(dateFormat, content.ReleaseDate)
	if err != nil {
		slog.Error("Failed to parse movie release date", "error", err)
	}

	c := entity.Content{
		TmdbID:      content.ID,
		Title:       content.Title,
		Overview:    content.Overview,
		PosterPath:  content.PosterPath,
		Type:        entity.MOVIE,
		ReleaseDate: &releaseDate,
		Popularity:  content.Popularity,
		VoteAverage: content.VoteAverage,
		VoteCount:   content.VoteCount,
		ImdbID:      content.ImdbID,
		Status:      content.Status,
		Budget:      content.Budget,
		Revenue:     content.Revenue,
		Runtime:     content.Runtime,
	}

	err = s.saveContent(&c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentMovie: Failed to save content!", "error", err)
		return entity.Content{}, errors.New("failed to save content")
	}

	return c, nil
}

// Get content from our db cache, or cache it if it doesn't exist.
func (s *Service) GetOrCacheContent(contentType entity.ContentType, tmdbId int) (entity.Content, error) {
	var content entity.Content
	// Look in db for content.
	s.db.Where("type = ? AND tmdb_id = ?", contentType, tmdbId).Find(&content)
	// Create content if not found from our db.
	if content == (entity.Content{}) {
		slog.Debug("Content not in db, fetching...", "type", contentType, "tmdbId", tmdbId)

		resp, err := s.tmdb.APIRequest("/"+string(contentType)+"/"+strconv.Itoa(tmdbId), map[string]string{})
		if err != nil {
			slog.Error("GetOrCacheContent: content tmdb api request failed", "error", err)
			return entity.Content{}, errors.New("failed to find requested media")
		}

		if contentType == "movie" {
			c := new(tmdb.TMDBMovieDetails)
			err := json.Unmarshal([]byte(resp), &c)
			if err != nil {
				slog.Error("Failed to unmarshal movie details", "error", err)
				return entity.Content{}, errors.New("failed to process movie details response")
			}
			content, err = s.CacheContentMovie(*c, false)
			if err != nil {
				slog.Error("GetOrCacheContent: failed to cache movie content", "type", contentType, "content_id", tmdbId, "err", err)
				return entity.Content{}, errors.New("failed to cache content")
			}
		} else {
			c := new(tmdb.TMDBShowDetails)
			err := json.Unmarshal(resp, &c)
			if err != nil {
				slog.Error("Failed to unmarshal tv details", "error", err)
				return entity.Content{}, errors.New("failed to process tv details response")
			}
			content, err = s.CacheContentTv(*c, false)
			if err != nil {
				slog.Error("GetOrCacheContent: failed to cache tv content", "type", contentType, "content_id", tmdbId, "err", err)
				return entity.Content{}, errors.New("failed to cache content")
			}
		}
	}
	return content, nil
}

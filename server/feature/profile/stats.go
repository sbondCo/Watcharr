package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type StatsResponse struct {
	Summary                  StatsSummary        `json:"summary"`
	RatingDistribution       []DistributionBucket `json:"ratingDistribution"`
	StatusDistribution       []StatusBucket       `json:"statusDistribution"`
	ReleaseYear              []YearBucket         `json:"releaseYear"`
	WatchYear                []YearBucket         `json:"watchYear"`
	EpisodeCountDistribution []DistributionBucket `json:"episodeCountDistribution,omitempty"`
}

type StatsSummary struct {
	TotalCount      int32   `json:"totalCount"`
	TotalRuntime    uint32  `json:"totalRuntime"`
	DaysWatched     float64 `json:"daysWatched"`
	MeanScore       float64 `json:"meanScore"`
	EpisodesWatched uint32  `json:"episodesWatched"`
}

type DistributionBucket struct {
	Label string `json:"label"`
	Count int32  `json:"count"`
}

type StatusBucket struct {
	Status string `json:"status"`
	Count  int32  `json:"count"`
}

type YearBucket struct {
	Year  int   `json:"year"`
	Count int32 `json:"count"`
}

func (s *Service) getStats(userId uint, contentType entity.ContentType) (StatsResponse, error) {
	user := new(entity.User)
	res := s.db.Model(&entity.User{}).Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("Failed to get user for stats:", "error", res.Error.Error())
		return StatsResponse{}, errors.New("failed to get user")
	}

	watched := new([]entity.Watched)
	res = s.db.Model(&entity.Watched{}).Preload("Content").Preload("Activity").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		slog.Error("Stats: Failed to get watched for processing:", "error", res.Error.Error())
		return StatsResponse{}, errors.New("failed to get watched for processing")
	}

	var (
		totalCount      int32
		totalRuntime    uint32
		episodesWatched uint32
		ratingSum       float64
		ratingCount     int32
	)

	ratingDist := make(map[int]int32)
	statusDist := make(map[entity.WatchedStatus]int32)
	releaseYearDist := make(map[int]int32)
	watchYearDist := make(map[int]int32)
	episodeCountDist := make(map[string]int32)

	includePrev := user.IncludePreviouslyWatched != nil && *user.IncludePreviouslyWatched

	for _, w := range *watched {
		if w.Content == nil {
			continue
		}
		c := *w.Content
		if c.Type != contentType {
			continue
		}

		// Count by status
		statusDist[w.Status]++
		totalCount++

		// Rating distribution (non-zero ratings only)
		if w.Rating > 0 {
			bucket := int(math.Round(w.Rating))
			if bucket < 1 {
				bucket = 1
			}
			if bucket > 10 {
				bucket = 10
			}
			ratingDist[bucket]++
			ratingSum += w.Rating
			ratingCount++
		}

		// Release year
		if c.ReleaseDate != nil {
			year := c.ReleaseDate.Year()
			if year > 0 {
				releaseYearDist[year]++
			}
		}

		// Runtime calculation (same logic as getProfile)
		isFinished := false
		if w.Status == entity.FINISHED {
			isFinished = true
		} else if includePrev && s.hasBeenPreviouslyWatched(&w.Activity) {
			isFinished = true
		}

		if isFinished {
			if c.Type == entity.SHOW {
				if c.NumberOfEpisodes != 0 {
					var showRuntime uint32 = 30
					if c.Runtime != 0 {
						showRuntime = c.Runtime
					}
					totalRuntime += showRuntime * c.NumberOfEpisodes
					episodesWatched += c.NumberOfEpisodes
				}
			} else if c.Type == entity.MOVIE {
				totalRuntime += c.Runtime
			}
		}

		// Watch year: find the latest FINISHED activity date
		watchYear := findWatchYear(&w, includePrev, s)
		if watchYear > 0 {
			watchYearDist[watchYear]++
		}

		// Episode count distribution (TV only)
		if contentType == entity.SHOW {
			label := classifyEpisodeCount(c.NumberOfEpisodes)
			episodeCountDist[label]++
		}
	}

	// Build response
	resp := StatsResponse{
		Summary: StatsSummary{
			TotalCount:      totalCount,
			TotalRuntime:    totalRuntime,
			DaysWatched:     math.Round(float64(totalRuntime)/60/24*10) / 10,
			EpisodesWatched: episodesWatched,
		},
		RatingDistribution: buildRatingDistribution(ratingDist),
		StatusDistribution: buildStatusDistribution(statusDist),
		ReleaseYear:        buildYearBuckets(releaseYearDist),
		WatchYear:          buildYearBuckets(watchYearDist),
	}

	if ratingCount > 0 {
		resp.Summary.MeanScore = math.Round(ratingSum/float64(ratingCount)*10) / 10
	}

	if contentType == entity.SHOW {
		resp.EpisodeCountDistribution = buildEpisodeCountDistribution(episodeCountDist)
	}

	return resp, nil
}

// findWatchYear determines the year an item was watched/finished.
func findWatchYear(w *entity.Watched, includePrev bool, s *Service) int {
	isFinished := w.Status == entity.FINISHED
	if !isFinished && includePrev {
		isFinished = s.hasBeenPreviouslyWatched(&w.Activity)
	}
	if !isFinished {
		return 0
	}

	// Look for the latest FINISHED activity
	var finishTime *time.Time
	for i := len(w.Activity) - 1; i >= 0; i-- {
		a := w.Activity[i]
		isFinishActivity := false

		if a.Type == entity.STATUS_CHANGED || a.Type == entity.STATUS_CHANGED_AUTO {
			if a.Data == "FINISHED" {
				isFinishActivity = true
			}
		} else if a.Type == entity.ADDED_WATCHED || a.Type == entity.IMPORTED_ADDED_WATCHED || a.Type == entity.IMPORTED_WATCHED {
			if a.Data != "" {
				var v map[string]any
				err := json.Unmarshal([]byte(a.Data), &v)
				if err == nil {
					if status, ok := v["status"]; ok && status == "FINISHED" {
						isFinishActivity = true
					}
				}
			}
			if a.Type == entity.IMPORTED_ADDED_WATCHED {
				isFinishActivity = true
			}
		}

		if isFinishActivity {
			if a.CustomDate != nil {
				finishTime = a.CustomDate
			} else {
				t := a.CreatedAt
				finishTime = &t
			}
			break
		}
	}

	if finishTime != nil {
		return finishTime.Year()
	}

	// Fall back to watched CreatedAt
	return w.CreatedAt.Year()
}

func classifyEpisodeCount(eps uint32) string {
	switch {
	case eps <= 1:
		return "1"
	case eps <= 12:
		return "2-12"
	case eps <= 24:
		return "13-24"
	case eps <= 50:
		return "25-50"
	case eps <= 100:
		return "51-100"
	default:
		return "101+"
	}
}

func buildRatingDistribution(m map[int]int32) []DistributionBucket {
	buckets := make([]DistributionBucket, 10)
	for i := 1; i <= 10; i++ {
		buckets[i-1] = DistributionBucket{
			Label: fmt.Sprintf("%d", i),
			Count: m[i],
		}
	}
	return buckets
}

func buildStatusDistribution(m map[entity.WatchedStatus]int32) []StatusBucket {
	statuses := []entity.WatchedStatus{entity.FINISHED, entity.WATCHING, entity.PLANNED, entity.HOLD, entity.DROPPED}
	var buckets []StatusBucket
	for _, s := range statuses {
		if count, ok := m[s]; ok && count > 0 {
			buckets = append(buckets, StatusBucket{
				Status: string(s),
				Count:  count,
			})
		}
	}
	return buckets
}

func buildYearBuckets(m map[int]int32) []YearBucket {
	var buckets []YearBucket
	for year, count := range m {
		buckets = append(buckets, YearBucket{Year: year, Count: count})
	}
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Year < buckets[j].Year
	})
	return buckets
}

func buildEpisodeCountDistribution(m map[string]int32) []DistributionBucket {
	order := []string{"1", "2-12", "13-24", "25-50", "51-100", "101+"}
	var buckets []DistributionBucket
	for _, label := range order {
		if count, ok := m[label]; ok && count > 0 {
			buckets = append(buckets, DistributionBucket{
				Label: label,
				Count: count,
			})
		}
	}
	return buckets
}

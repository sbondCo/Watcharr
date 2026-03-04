package util

import (
	"log/slog"
	"math"

	"gorm.io/gorm"
)

// Parameters that the paginator uses to
// know what to return.
type PaginationParams struct {
	// Page limit (max amount of items to get for each page).
	Limit int `json:"limit"`
	// Page number.
	Page int `json:"page"`
}

// Pagination response struct.
type PaginationResponse[T any, U any] struct {
	PaginationParams
	// Max amount of pages we can produce from total_results
	TotalPages int `json:"totalPages"`
	// Specifically an `int64` so we can use it
	// directly in gorms Count function.
	TotalResults int64 `json:"totalResults"`
	Results      []T   `json:"results"`
	// Metadata, any service can use this to add extra properties that don't
	// fit in the normal PaginationResponse struct (or in the Results slice).
	Meta U `json:"meta,omitzero"`
}

// Call when finished with PaginationResponse, before returning to user.
// Performs final calculations.
func (r *PaginationResponse[T, U]) Finished(p PaginationParams) {
	r.PaginationParams = p
	if r.TotalResults != 0 && r.Limit != 0 {
		r.TotalPages = int(math.Ceil(float64(r.TotalResults) / float64(r.Limit)))
	} else {
		slog.Warn(
			"PaginationResponse->Finished: TotalPages not calculated.",
			"total_results", r.TotalResults,
			"limit", r.Limit,
		)
	}
}

// Pagination gorm scope.
// Pass in `PaginationParams` and the `PaginationResponse` will be filled out,
// just fill out the `Results` manually.
func Paginate[T any, U any](
	p PaginationParams,
	r *PaginationResponse[T, U],
) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.Page - 1) * p.Limit
		return db.Offset(offset).Limit(p.Limit)
	}
}

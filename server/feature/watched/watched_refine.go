// Watched sorting & filtering.

package watched

import (
	"log/slog"
	"strings"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Applies 'Type' filter.
func refineFilterType(db *gorm.DB, ft []util.SupportedMedia) {
	if len(ft) <= 0 {
		return
	}
	qstr := []string{}
	astr := []any{}
	for _, v := range ft {
		switch v {
		case util.SupportedMediaMovie:
			qstr = append(qstr, "Content.type = ?")
			astr = append(astr, "movie")
		case util.SupportedMediaShow:
			qstr = append(qstr, "Content.type = ?")
			astr = append(astr, "tv")
		case util.SupportedMediaGame:
			qstr = append(qstr, "watcheds.game_id IS NOT NULL")
		}
	}
	q := strings.Join(qstr, " OR ")
	slog.Debug("watchedRefine: Filter type.", "q", q, "astr", astr)
	db.Where(q, astr...)
}

// Applies 'Status' filter.
func refineFilterStatus(db *gorm.DB, f []entity.WatchedStatus) {
	if len(f) <= 0 {
		return
	}
	for i := range f {
		// Ensure string **case** is valid WatchedStatus by converting to uppercase.
		f[i] = entity.WatchedStatus(strings.ToUpper(string(f[i])))
	}
	db.Where("watcheds.status IN ?", f)
}

// Applies sorts to list.
// Takes in userId of user who owns the list we are sorting, since some sorts
// may require it for subqueries (eg LastFinished).
func refineSort(
	db *gorm.DB,
	userId uint,
	sort domain.WatchedSort,
	dir domain.SortDirection,
) {
	if sort == "" {
		return
	}
	obc := func(c clause.Column) clause.OrderByColumn {
		o := clause.OrderByColumn{}
		if dir == domain.WatchedSortDirAsc {
			o.Desc = false
		} else {
			o.Desc = true
		}
		o.Column = c
		return o
	}
	switch sort {
	case domain.WatchedSortDateAdded:
		db.Order(obc(clause.Column{Name: "watcheds.created_at"}))
	case domain.WatchedSortLastChanged:
		db.Order(obc(clause.Column{Name: "watcheds.updated_at"}))
	case domain.WatchedSortLastFinished:
		db.
			// This join looks for the latest activity that counts as a play
			// for each watched entry. The date of these is used in the sort
			// below.
			// Note: Technically the join subquery will process ALL activities
			// that the user has by their user_id, BUT this is okay since we
			// use this sorting over the users entire watched list, so we want
			// to process every activity to join to main list for the sort
			// anyways. I'm making this note because previously I had left out
			// the user_id WHERE, which results in all activities in the table
			// being processed, which is obviously NOT wanted (cuz its slower).
			Joins(`LEFT JOIN (
					SELECT
						watched_id AS a_watched_id,
						MAX(COALESCE(custom_date, created_at)) AS a_sort_by_date
					FROM activities
					WHERE
						count_as_play = 1
						AND deleted_at IS NULL
						AND user_id = ?
					GROUP BY watched_id
				) q ON q.a_watched_id = watcheds.id`, userId).
			Order(obc(clause.Column{Name: "q.a_sort_by_date"}))
	case domain.WatchedSortRating:
		db.Order(obc(clause.Column{Name: "watcheds.rating"}))
	case domain.WatchedSortAlphabetical:
		db.Order(obc(clause.Column{
			Name: "COALESCE(`Content`.`title`, `Game`.`name`)",
			Raw:  true,
		}))
	case domain.WatchedSortDateReleased:
		db.Order(obc(clause.Column{
			Name: "COALESCE(`Content`.`release_date`, `Game`.`release_date`)",
			Raw:  true,
		}))
	}
}

// Keep pinned items to top.
func refineSortPinned(db *gorm.DB) {
	db.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "watcheds.pinned"},
		Desc:   true,
	})
}

// list data.
// gorm scope for applying sort and filters to watched
func watchedRefine(wr domain.WatchedGetPageRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Apply filters
		refineFilterType(db, wr.FilterType)
		refineFilterStatus(db, wr.FilterStatus)
		// Apply sort
		refineSortPinned(db)
		refineSort(db, wr.Sort, wr.SortDir)
		return db
	}
}

// Watched sorting & filtering.

package main

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gorm scope for applying sort and filters to watched
// list data.
func watchedRefine(wr WatchedGetPageRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Apply filters

		// Apply sort
		if wr.Sort != "" {
			obc := func(cn string) clause.OrderByColumn {
				o := clause.OrderByColumn{}
				if wr.SortDir == sortAscending {
					o.Desc = false
				} else {
					o.Desc = true
				}
				o.Column = clause.Column{Name: cn}
				return o
			}
			switch wr.Sort {
			case watchedSortDateAdded:
				db.Order(obc("watcheds.created_at"))
			case watchedSortLastChanged:
				db.Order(obc("watcheds.updated_at"))
			case watchedSortLastFinished:
				// TODO This can make the query quite slow, look at improving performance.
				db.
					Joins("LEFT JOIN activities ON activities.watched_id = watcheds.id").
					// TODO this whole query looks to work, but we have to add `watcheds.*` to this SELECT,
					// otherwise it doesn't select them (like it does when we use Model, in the original query build),
					// is there a better way? Cuz if we modify the Model later in main query, this would break, which isn't ideal.
					Select("watcheds.*, MAX(MAX(activities.created_at), MAX(activities.custom_date)) as latest_watched_activity").
					Group("watcheds.id").
					Order(obc("latest_watched_activity"))
			case watchedSortRating:
				db.Order(obc("watcheds.rating"))
			case watchedSortAlphabetical:
				db.
					Order(obc("Content__title")).
					Order(obc("Game__name"))
			}
		}
		return db
	}
}

package search

import (
	"log/slog"
	"strconv"
	"strings"
)

// All accepted query filters for any media search.
type AllParsableQueryFilters struct {
	Year      int
	FirstYear int
	Adult     bool
}

// Takes query in and parses out any inline filters into a struct.
// Takes out anything in supported format: `a:b`.
// Returns:
//   - query string with any parsed filters removed.
//   - the parsed filters;
func parseQueryFilters(query string) (string, AllParsableQueryFilters) {
	const segSplitStr = " "

	m := AllParsableQueryFilters{}
	segs /* hehe */ := strings.SplitSeq(query, segSplitStr)
	// Segments of the query that aren't filters are added back to this slice
	// and filters are not, so when we join this slice back into a string at the
	// end, we are left with only the query and not any filters.
	finalQuery := []string{}
	notParsing := func(s string) {
		finalQuery = append(finalQuery, s)
	}
	for seg := range segs {
		split := strings.Split(seg, ":")
		if len(split) != 2 || (split[0] == "" || split[1] == "") {
			// We only support filtername:value, so:
			// 	- more or less than len of 2 = wrong; and
			//	- k or v being empty = wrong.
			notParsing(seg)
			continue
		}
		// Add filter to struct.
		fkey := strings.ToLower(split[0])
		switch fkey {
		case "year", "y":
			i, _ := strconv.Atoi(split[1])
			m.Year = i
		case "fyear", "fy":
			i, _ := strconv.Atoi(split[1])
			m.FirstYear = i
		case "adult":
			if split[1] != "" {
				m.Adult = true
			}
		default:
			// If no key matches a supported one, then don't parse this either.
			notParsing(seg)
		}
	}
	slog.Debug("ParseQueryFilters: Done job.",
		"query", query,
		"finalQuery", finalQuery,
		"parsed_filters", m)
	return strings.Join(finalQuery, segSplitStr), m
}

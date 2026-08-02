package search

import (
	"testing"
)

type Expectation struct {
	Query  string
	Struct AllParsableQueryFilters
}

func TestParseQueryFilters(t *testing.T) {
	testSet := map[string]Expectation{
		// No effect on query without any filters.
		"Joker": {
			Query:  "Joker",
			Struct: AllParsableQueryFilters{},
		},
		// Parses a lone filter correctly.
		"Joker year:2024": {
			Query: "Joker",
			Struct: AllParsableQueryFilters{
				Year: 2024,
			},
		},
		// Multiple filters should parse successfully.
		"Harvest MOON year:2024 adult:true": {
			Query: "Harvest MOON",
			Struct: AllParsableQueryFilters{
				Year:  2024,
				Adult: true,
			},
		},
		// Filter before or after the title should still be parsed.
		"y:1999 Joker 2 adult:1": {
			Query: "Joker 2",
			Struct: AllParsableQueryFilters{
				Year:  1999,
				Adult: true,
			},
		},
		// Test that the `2:` doesnt get parsed, since it's common for media to
		// have names that use colons like that.
		"y:t Joker 2: The sun rises! adult:1": {
			Query: "Joker 2: The sun rises!",
			Struct: AllParsableQueryFilters{
				Year:  0,
				Adult: true,
			},
		},
		// Test that a non whitelisted key "man", doesn't get parsed and removed
		// from the query.
		"spider man:new": {
			Query:  "spider man:new",
			Struct: AllParsableQueryFilters{},
		},
	}

	for query, exp := range testSet {
		q, m := parseQueryFilters(query)
		if q != exp.Query {
			t.Errorf("query '%v' doesn't match expected query '%v'", q, exp.Query)
		}
		if m != exp.Struct {
			t.Errorf("%v doesn't match expected %v", m, exp.Struct)
		}
	}
}

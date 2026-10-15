package content

import "testing"

func TestActingJob(t *testing.T) {
	for character, want := range map[string]string{
		"Himself":                   "Self",
		"Herself (archive footage)": "Self",
		"Self - Host":               "Self",
		" self ":                    "Self",
		"Bruce Wayne / Batman":      "Actor",
		"Selfridge":                 "Actor",
		"":                          "Actor",
	} {
		if got := actingJob(character); got != want {
			t.Errorf("actingJob(%q) = %q, want %q", character, got, want)
		}
	}
}

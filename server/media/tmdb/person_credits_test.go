package tmdb

import (
	"encoding/json"
	"testing"
)

func TestCrewByDepartment(t *testing.T) {
	// Shape of TMDB's /person/{id}/combined_credits: one crew entry per job.
	raw := `{"id":525,"cast":[{"id":1,"media_type":"movie","title":"Cameo"}],"crew":[
		{"id":27205,"media_type":"movie","title":"Inception","department":"Directing","job":"Director"},
		{"id":27205,"media_type":"movie","title":"Inception","department":"Writing","job":"Screenplay"},
		{"id":27205,"media_type":"movie","title":"Inception","department":"Production","job":"Producer"},
		{"id":27205,"media_type":"movie","title":"Inception","department":"Production","job":"Executive Producer"},
		{"id":27205,"media_type":"tv","name":"Same id, a series","department":"Directing","job":"Director"},
		{"id":9,"media_type":"person","department":"Directing","job":"Director"},
		{"id":10,"media_type":"movie","title":"No department","department":"","job":"?"}
	]}`
	var c PersonCombinedCredits
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Cast) != 1 || len(c.Crew) != 7 {
		t.Fatalf("decoded %d cast, %d crew", len(c.Cast), len(c.Crew))
	}
	// Promoted from the embedded cast fields.
	if c.Crew[0].Title != "Inception" || c.Crew[0].Job != "Director" {
		t.Errorf("crew[0] = %+v", c.Crew[0])
	}

	got := c.CrewByDepartment()
	count := map[string]int{}
	for d, l := range got {
		count[d] = len(l)
	}
	// Directing: the film and the series sharing its id are two titles.
	// Production: two jobs on one film are one entry. Non-media and
	// department-less entries are dropped.
	want := map[string]int{"Directing": 2, "Writing": 1, "Production": 1}
	if len(count) != len(want) {
		t.Fatalf("departments = %v, want %v", count, want)
	}
	for d, n := range want {
		if count[d] != n {
			t.Errorf("%s: %d titles, want %d", d, count[d], n)
		}
	}
	if got["Writing"][0].Name != "Inception" {
		t.Errorf("writing: %+v", got["Writing"][0])
	}
	// One entry per title, but every job kept, in TMDB's order.
	if j := got["Production"][0].Jobs; len(j) != 2 || j[0] != "Producer" || j[1] != "Executive Producer" {
		t.Errorf("production jobs = %v", j)
	}
	if j := got["Directing"][0].Jobs; len(j) != 1 || j[0] != "Director" {
		t.Errorf("directing jobs = %v", j)
	}
}

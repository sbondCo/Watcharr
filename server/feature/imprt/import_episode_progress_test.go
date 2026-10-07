package imprt

import (
	"errors"
	"testing"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/util"
)

// TestEpisodesForCount covers turning a watched count into the episodes it
// covers, which is all a MyAnimeList export gives us to go on. A MyAnimeList
// entry is a single season, so the count never leaves the first season.
func TestEpisodesForCount(t *testing.T) {
	tests := []struct {
		name    string
		seasons []seasonEpisodeCount
		count   int
		want    []episodeRef
	}{
		{
			name:    "part of a season",
			seasons: []seasonEpisodeCount{{Number: 1, EpisodeCount: 12}},
			count:   3,
			want:    []episodeRef{{1, 1}, {1, 2}, {1, 3}},
		},
		{
			name:    "a whole season",
			seasons: []seasonEpisodeCount{{Number: 1, EpisodeCount: 3}},
			count:   3,
			want:    []episodeRef{{1, 1}, {1, 2}, {1, 3}},
		},
		{
			name: "never carries over into the next season",
			seasons: []seasonEpisodeCount{
				{Number: 1, EpisodeCount: 2},
				{Number: 2, EpisodeCount: 4},
			},
			count: 4,
			want:  []episodeRef{{1, 1}, {1, 2}},
		},
		{
			name: "a count mal counts higher than tmdb is clamped",
			seasons: []seasonEpisodeCount{
				{Number: 1, EpisodeCount: 12},
				{Number: 2, EpisodeCount: 12},
			},
			// Bakemonogatari: mal says 15, tmdb season one holds 12.
			count: 15,
			want: []episodeRef{
				{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6},
				{1, 7}, {1, 8}, {1, 9}, {1, 10}, {1, 11}, {1, 12},
			},
		},
		{
			name: "specials are skipped so they dont shift episodes along",
			seasons: []seasonEpisodeCount{
				{Number: 0, EpisodeCount: 5},
				{Number: 1, EpisodeCount: 3},
			},
			count: 2,
			want:  []episodeRef{{1, 1}, {1, 2}},
		},
		{
			name: "seasons out of order still fill the first one",
			seasons: []seasonEpisodeCount{
				{Number: 2, EpisodeCount: 2},
				{Number: 1, EpisodeCount: 2},
			},
			count: 2,
			want:  []episodeRef{{1, 1}, {1, 2}},
		},
		{
			name: "a season with no episodes is stepped over",
			seasons: []seasonEpisodeCount{
				{Number: 1, EpisodeCount: 0},
				{Number: 2, EpisodeCount: 2},
			},
			count: 2,
			want:  []episodeRef{{2, 1}, {2, 2}},
		},
		{
			name:    "a count of zero gives nothing",
			seasons: []seasonEpisodeCount{{Number: 1, EpisodeCount: 12}},
			count:   0,
			want:    nil,
		},
		{
			name:    "a negative count gives nothing",
			seasons: []seasonEpisodeCount{{Number: 1, EpisodeCount: 12}},
			count:   -3,
			want:    nil,
		},
		{
			name:    "a show we know no seasons for gives nothing",
			seasons: nil,
			count:   5,
			want:    nil,
		},
		{
			name:    "a show with only specials gives nothing",
			seasons: []seasonEpisodeCount{{Number: 0, EpisodeCount: 5}},
			count:   5,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := episodesForCount(tt.seasons, tt.count)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d episodes %v, want %d %v",
					len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("episode %d = s%de%d, want s%de%d", i,
						got[i].SeasonNumber, got[i].EpisodeNumber,
						tt.want[i].SeasonNumber, tt.want[i].EpisodeNumber)
				}
			}
		})
	}
}

// TestEpisodesForCountSequelEntry checks the case that made us stop spreading
// over seasons. MyAnimeList gives a sequel its own entry, and that entry
// matches the same tmdb show as the original, so its count must not be used
// to mark episodes of a season it knows nothing about. Clannad is the case we
// hit: the entry says 23 and "Clannad: After Story" says 24, and the second
// entry used to leave season two with two episodes marked out of 24.
func TestEpisodesForCountSequelEntry(t *testing.T) {
	seasons := []seasonEpisodeCount{
		{Number: 1, EpisodeCount: 22},
		{Number: 2, EpisodeCount: 24},
	}

	for _, count := range []int{23, 24} {
		got := episodesForCount(seasons, count)
		for _, e := range got {
			if e.SeasonNumber != 1 {
				t.Fatalf("count %d marked s%de%d, want season 1 only",
					count, e.SeasonNumber, e.EpisodeNumber)
			}
		}
		if len(got) != 22 {
			t.Errorf("count %d marked %d episodes, want 22 (season one clamped)",
				count, len(got))
		}
	}
}

// TestHasWatchedEpisode checks we only count an episode as already watched
// when both its season and episode number match, so that the same episode
// number in a different season is not mistaken for it.
func TestHasWatchedEpisode(t *testing.T) {
	eps := []entity.WatchedEpisode{
		{SeasonNumber: 1, EpisodeNumber: 2},
		{SeasonNumber: 2, EpisodeNumber: 5},
	}

	tests := []struct {
		name string
		ep   episodeRef
		want bool
	}{
		{"an episode we have", episodeRef{1, 2}, true},
		{"another episode we have", episodeRef{2, 5}, true},
		{"same episode number, different season", episodeRef{2, 2}, false},
		{"same season, different episode number", episodeRef{1, 5}, false},
		{"an episode we have nothing for", episodeRef{3, 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasWatchedEpisode(eps, tt.ep); got != tt.want {
				t.Errorf("hasWatchedEpisode(s%de%d) = %v, want %v",
					tt.ep.SeasonNumber, tt.ep.EpisodeNumber, got, tt.want)
			}
		})
	}
}

// Stands in for the watched episode service, recording what it was asked to
// add and handing back the episode list it would have saved.
type fakeWatchedEpisodeProvider struct {
	calls []episode.WatchedEpisodeAddRequest
	saved []entity.WatchedEpisode
	err   error
}

func (f *fakeWatchedEpisodeProvider) AddWatchedEpisodes(
	userId uint,
	ar episode.WatchedEpisodeAddRequest,
) (episode.WatchedEpisodeAddResponse, error) {
	f.calls = append(f.calls, ar)
	if f.err != nil {
		return episode.WatchedEpisodeAddResponse{}, f.err
	}
	f.saved = append(f.saved, entity.WatchedEpisode{
		WatchedID:     ar.WatchedID,
		SeasonNumber:  ar.SeasonNumber,
		EpisodeNumber: ar.EpisodeNumber,
		Status:        ar.Status,
	})
	return episode.WatchedEpisodeAddResponse{WatchedEpisodes: f.saved}, nil
}

// TestAddEpisodes covers marking episodes as watched on an entry: episodes
// the user already has are left alone, and the rest are added as FINISHED.
func TestAddEpisodes(t *testing.T) {
	tests := []struct {
		name string
		// Episodes already on the entry.
		existing []entity.WatchedEpisode
		eps      []episodeRef
		addErr   error
		// Episodes we expect to be asked to add, in order.
		wantCalls []episodeRef
		wantAdded int
	}{
		{
			name:      "adds every episode when the entry has none",
			eps:       []episodeRef{{1, 1}, {1, 2}},
			wantCalls: []episodeRef{{1, 1}, {1, 2}},
			wantAdded: 2,
		},
		{
			name: "skips episodes the user already has",
			existing: []entity.WatchedEpisode{
				{SeasonNumber: 1, EpisodeNumber: 1},
			},
			eps:       []episodeRef{{1, 1}, {1, 2}},
			wantCalls: []episodeRef{{1, 2}},
			wantAdded: 1,
		},
		{
			name: "adds nothing when the user already has it all",
			existing: []entity.WatchedEpisode{
				{SeasonNumber: 1, EpisodeNumber: 1},
				{SeasonNumber: 1, EpisodeNumber: 2},
			},
			eps:       []episodeRef{{1, 1}, {1, 2}},
			wantCalls: nil,
			wantAdded: 0,
		},
		{
			name:      "counts nothing as added when saving fails",
			eps:       []episodeRef{{1, 1}, {1, 2}},
			addErr:    errors.New("failed to save"),
			wantCalls: []episodeRef{{1, 1}, {1, 2}},
			wantAdded: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeWatchedEpisodeProvider{err: tt.addErr}
			s := &Service{wep: fake}
			w := entity.Watched{
				GormModel:       dbmodel.GormModel{ID: 7},
				WatchedEpisodes: tt.existing,
			}

			added := s.addEpisodes(1, &w, tt.eps)

			if added != tt.wantAdded {
				t.Errorf("added = %d, want %d", added, tt.wantAdded)
			}
			if len(fake.calls) != len(tt.wantCalls) {
				t.Fatalf("AddWatchedEpisodes calls = %d, want %d",
					len(fake.calls), len(tt.wantCalls))
			}
			for i, c := range fake.calls {
				if c.SeasonNumber != tt.wantCalls[i].SeasonNumber ||
					c.EpisodeNumber != tt.wantCalls[i].EpisodeNumber {
					t.Errorf("call %d = s%de%d, want s%de%d", i,
						c.SeasonNumber, c.EpisodeNumber,
						tt.wantCalls[i].SeasonNumber, tt.wantCalls[i].EpisodeNumber)
				}
				if c.WatchedID != w.ID {
					t.Errorf("call %d watched id = %d, want %d", i, c.WatchedID, w.ID)
				}
				if c.Status != entity.FINISHED {
					t.Errorf("call %d status = %q, want %q", i, c.Status, entity.FINISHED)
				}
			}
		})
	}
}

// TestAddEpisodesTwiceAddsNothingTheSecondTime checks re-running an import
// over an entry we already filled in leaves it exactly as it was.
func TestAddEpisodesTwiceAddsNothingTheSecondTime(t *testing.T) {
	fake := &fakeWatchedEpisodeProvider{}
	s := &Service{wep: fake}
	w := entity.Watched{GormModel: dbmodel.GormModel{ID: 7}}
	eps := []episodeRef{{1, 1}, {1, 2}, {2, 1}}

	if added := s.addEpisodes(1, &w, eps); added != 3 {
		t.Fatalf("first run added = %d, want 3", added)
	}
	if added := s.addEpisodes(1, &w, eps); added != 0 {
		t.Errorf("second run added = %d, want 0", added)
	}
	if len(fake.calls) != 3 {
		t.Errorf("AddWatchedEpisodes calls = %d, want 3 (the second run must"+
			" not ask for anything)", len(fake.calls))
	}
}

// TestFillInMissingEpisodes_NothingToDo covers the cases we bail out of
// before ever looking a show up, so an import that carries no progress (or
// content that has no episodes at all) is reported as existing, untouched.
func TestFillInMissingEpisodes_NothingToDo(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		contentType util.SupportedMedia
		existing    entity.Watched
		getErr      error
	}{
		{
			name:        "import carries no episode count",
			count:       0,
			contentType: util.SupportedMediaShow,
			existing:    entity.Watched{GormModel: dbmodel.GormModel{ID: 1}},
		},
		{
			name:        "content is a movie, which has no episodes",
			count:       12,
			contentType: util.SupportedMediaMovie,
			existing:    entity.Watched{GormModel: dbmodel.GormModel{ID: 1}},
		},
		{
			name:        "content is a game, which is not looked up by tmdb id",
			count:       12,
			contentType: util.SupportedMediaGame,
			existing:    entity.Watched{GormModel: dbmodel.GormModel{ID: 1}},
		},
		{
			name:        "no existing entry was found to fill in",
			count:       12,
			contentType: util.SupportedMediaShow,
			existing:    entity.Watched{},
		},
		{
			name:        "looking the existing entry up failed",
			count:       12,
			contentType: util.SupportedMediaShow,
			getErr:      errors.New("db is down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeEps := &fakeWatchedEpisodeProvider{}
			s := &Service{
				wp:  &fakeWatchedProvider{existing: tt.existing, getErr: tt.getErr},
				wep: fakeEps,
			}

			resp := s.fillInMissingEpisodes(
				1,
				&domain.ImportRequest{WatchedEpisodesCount: tt.count},
				domain.SuccessfulImportProps{
					TmdbID:      43167,
					ContentType: tt.contentType,
				},
			)

			if resp.Type != domain.IMPORT_EXISTS {
				t.Errorf("response type = %q, want %q", resp.Type, domain.IMPORT_EXISTS)
			}
			if len(fakeEps.calls) != 0 {
				t.Errorf("AddWatchedEpisodes called %d times, want 0",
					len(fakeEps.calls))
			}
		})
	}
}

package imprt

import (
	"errors"
	"testing"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
)

// Stands in for the watched service. Only the methods fillInMissingRating
// uses are given real behaviour, the rest are here to satisfy the interface.
type fakeWatchedProvider struct {
	existing    entity.Watched
	getErr      error
	updateErr   error
	updateCalls []ratingUpdate
}

type ratingUpdate struct {
	watchedId uint
	rating    float64
}

func (f *fakeWatchedProvider) AddWatched(
	userId uint,
	ar domain.WatchedAddRequest,
	extraProps domain.WatchedAddExtraProps,
) (entity.Watched, error) {
	return entity.Watched{}, domain.ErrWatchedExists
}

func (f *fakeWatchedProvider) GetWatchedItemByTmdbId(
	userId uint,
	tmdbId uint,
	contentType entity.ContentType,
) (entity.Watched, error) {
	if f.getErr != nil {
		return entity.Watched{}, f.getErr
	}
	return f.existing, nil
}

func (f *fakeWatchedProvider) UpdateWatchedRating(
	userId uint,
	watchedId uint,
	rating float64,
) error {
	f.updateCalls = append(f.updateCalls, ratingUpdate{watchedId, rating})
	return f.updateErr
}

// TestFillInMissingRating covers the rules for filling in a rating on content
// that is already on the users list: a rating is only ever added when the
// import has one and the existing entry does not, and an existing rating is
// never overwritten.
func TestFillInMissingRating(t *testing.T) {
	const tmdbId = 43167

	tests := []struct {
		name string
		// Rating carried by the import.
		importRating float64
		// The entry already on the users list.
		existing   entity.Watched
		getErr     error
		updateErr  error
		wantType   domain.ImportResponseType
		wantUpdate bool
		wantRating float64
	}{
		{
			name:         "fills in a missing rating",
			importRating: 9,
			existing:     entity.Watched{GormModel: dbmodel.GormModel{ID: 1}, Rating: 0},
			wantType:     domain.IMPORT_RATING_UPDATED,
			wantUpdate:   true,
			wantRating:   9,
		},
		{
			name:         "does not overwrite an existing rating",
			importRating: 5,
			existing:     entity.Watched{GormModel: dbmodel.GormModel{ID: 1}, Rating: 9},
			wantType:     domain.IMPORT_EXISTS,
			wantUpdate:   false,
		},
		{
			name:         "import without a rating changes nothing",
			importRating: 0,
			existing:     entity.Watched{GormModel: dbmodel.GormModel{ID: 1}, Rating: 0},
			wantType:     domain.IMPORT_EXISTS,
			wantUpdate:   false,
		},
		{
			name:         "no existing entry found",
			importRating: 9,
			existing:     entity.Watched{},
			wantType:     domain.IMPORT_EXISTS,
			wantUpdate:   false,
		},
		{
			name:         "lookup failure is not fatal",
			importRating: 9,
			getErr:       errors.New("db exploded"),
			wantType:     domain.IMPORT_EXISTS,
			wantUpdate:   false,
		},
		{
			name:         "update failure falls back to exists",
			importRating: 9,
			existing:     entity.Watched{GormModel: dbmodel.GormModel{ID: 1}, Rating: 0},
			updateErr:    errors.New("db exploded"),
			wantType:     domain.IMPORT_EXISTS,
			wantUpdate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeWatchedProvider{
				existing:  tt.existing,
				getErr:    tt.getErr,
				updateErr: tt.updateErr,
			}
			s := &Service{wp: fake}

			resp := s.fillInMissingRating(
				1,
				&domain.ImportRequest{Rating: tt.importRating},
				domain.SuccessfulImportProps{
					TmdbID:      tmdbId,
					ContentType: util.SupportedMediaShow,
				},
			)

			if resp.Type != tt.wantType {
				t.Fatalf("response type = %q, want %q", resp.Type, tt.wantType)
			}
			if tt.wantUpdate {
				if len(fake.updateCalls) != 1 {
					t.Fatalf("UpdateWatchedRating calls = %d, want 1", len(fake.updateCalls))
				}
				if fake.updateCalls[0].rating != tt.importRating {
					t.Errorf("updated with rating %v, want %v",
						fake.updateCalls[0].rating, tt.importRating)
				}
				if fake.updateCalls[0].watchedId != tt.existing.ID {
					t.Errorf("updated watched id %d, want %d",
						fake.updateCalls[0].watchedId, tt.existing.ID)
				}
			} else if len(fake.updateCalls) != 0 {
				t.Fatalf("UpdateWatchedRating called %d times, want 0 (a rating"+
					" must not be touched here)", len(fake.updateCalls))
			}
			if tt.wantRating != 0 && resp.WatchedEntry.Rating != tt.wantRating {
				t.Errorf("returned entry rating = %v, want %v",
					resp.WatchedEntry.Rating, tt.wantRating)
			}
		})
	}
}

// TestFillInMissingRating_NonTmdbContent checks content that is not looked up
// by tmdb id (games) is left alone rather than being matched incorrectly.
func TestFillInMissingRating_NonTmdbContent(t *testing.T) {
	fake := &fakeWatchedProvider{
		existing: entity.Watched{GormModel: dbmodel.GormModel{ID: 1}, Rating: 0},
	}
	s := &Service{wp: fake}

	resp := s.fillInMissingRating(
		1,
		&domain.ImportRequest{Rating: 9},
		domain.SuccessfulImportProps{
			IgdbID:      123,
			ContentType: util.SupportedMediaGame,
		},
	)

	if resp.Type != domain.IMPORT_EXISTS {
		t.Fatalf("response type = %q, want %q", resp.Type, domain.IMPORT_EXISTS)
	}
	if len(fake.updateCalls) != 0 {
		t.Fatalf("UpdateWatchedRating called for a game, want no calls")
	}
}

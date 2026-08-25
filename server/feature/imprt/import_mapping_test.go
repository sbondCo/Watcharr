package imprt

import (
	"testing"

	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

// TestImportMappingRoundTrip checks a choice saved for one name is found
// again for that same name, and that near misses (different user, different
// content type, unknown name) do not match it.
func TestImportMappingRoundTrip(t *testing.T) {
	testutil.SetupLogging()
	s := &Service{db: testutil.SetupDB(t)}

	const userId = 1
	const otherUserId = 2

	s.saveImportMapping(userId, &domain.ImportRequest{
		Name:   "Oshi no Ko",
		Type:   domain.ImportContentTypeShow,
		TmdbID: 203737,
	})

	t.Run("same name is found", func(t *testing.T) {
		m := s.findImportMapping(userId, "Oshi no Ko", domain.ImportContentTypeShow)
		if m == nil {
			t.Fatal("saved mapping was not found")
		}
		if m.TmdbID != 203737 {
			t.Errorf("tmdb id = %d, want 203737", m.TmdbID)
		}
	})

	t.Run("matching ignores case and surrounding whitespace", func(t *testing.T) {
		m := s.findImportMapping(userId, "  oShI No KO  ", domain.ImportContentTypeShow)
		if m == nil {
			t.Fatal("mapping was not found for a differently cased name")
		}
		if m.TmdbID != 203737 {
			t.Errorf("tmdb id = %d, want 203737", m.TmdbID)
		}
	})

	t.Run("another users mapping is not used", func(t *testing.T) {
		if m := s.findImportMapping(otherUserId, "Oshi no Ko", domain.ImportContentTypeShow); m != nil {
			t.Fatalf("found another users mapping (tmdb %d), mappings must be per user", m.TmdbID)
		}
	})

	t.Run("same name as a different content type is not used", func(t *testing.T) {
		if m := s.findImportMapping(userId, "Oshi no Ko", domain.ImportContentTypeMovie); m != nil {
			t.Fatalf("show mapping was returned for a movie import (tmdb %d)", m.TmdbID)
		}
	})

	t.Run("unknown name returns nothing", func(t *testing.T) {
		if m := s.findImportMapping(userId, "Some Other Show", domain.ImportContentTypeShow); m != nil {
			t.Fatalf("found a mapping for a name that was never saved (tmdb %d)", m.TmdbID)
		}
	})
}

// TestImportMappingIsUpdated checks that changing your mind about what a name
// refers to replaces the old choice rather than leaving it in place or
// creating a second row.
func TestImportMappingIsUpdated(t *testing.T) {
	testutil.SetupLogging()
	s := &Service{db: testutil.SetupDB(t)}

	const userId = 1
	req := domain.ImportRequest{
		Name:   "Oshi no Ko",
		Type:   domain.ImportContentTypeShow,
		TmdbID: 203737,
	}
	s.saveImportMapping(userId, &req)

	req.TmdbID = 244377
	s.saveImportMapping(userId, &req)

	m := s.findImportMapping(userId, "Oshi no Ko", domain.ImportContentTypeShow)
	if m == nil {
		t.Fatal("mapping was not found after being updated")
	}
	if m.TmdbID != 244377 {
		t.Errorf("tmdb id = %d, want 244377 (the newer choice)", m.TmdbID)
	}

	var count int64
	s.db.Table("import_mappings").
		Where("user_id = ? AND name = ?", userId, "oshi no ko").
		Count(&count)
	if count != 1 {
		t.Errorf("stored rows = %d, want 1 (updating must not add a row)", count)
	}
}

// TestSaveImportMappingIgnoresUnusableRequests checks we do not store rows we
// could never match on later.
func TestSaveImportMappingIgnoresUnusableRequests(t *testing.T) {
	testutil.SetupLogging()
	s := &Service{db: testutil.SetupDB(t)}

	const userId = 1

	tests := []struct {
		name string
		req  domain.ImportRequest
	}{
		{
			name: "no name to key on",
			req:  domain.ImportRequest{Type: domain.ImportContentTypeShow, TmdbID: 203737},
		},
		{
			name: "no content type to key on",
			req:  domain.ImportRequest{Name: "Oshi no Ko", TmdbID: 203737},
		},
		{
			name: "no id worth remembering",
			req:  domain.ImportRequest{Name: "Oshi no Ko", Type: domain.ImportContentTypeShow},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.saveImportMapping(userId, &tt.req)
			var count int64
			s.db.Table("import_mappings").Count(&count)
			if count != 0 {
				t.Fatalf("stored %d mappings, want 0", count)
			}
		})
	}
}

// TestImportMappingCRUD covers the endpoints behind the mappings page:
// listing, repointing at different content, and forgetting. Each is scoped
// by user id, so one user cannot read or change another users mappings.
func TestImportMappingCRUD(t *testing.T) {
	testutil.SetupLogging()
	s := &Service{db: testutil.SetupDB(t)}

	const userId = 1
	const otherUserId = 2

	s.saveImportMapping(userId, &domain.ImportRequest{
		Name: "Oshi no Ko", Type: domain.ImportContentTypeShow, TmdbID: 203737,
	})
	s.saveImportMapping(userId, &domain.ImportRequest{
		Name: "Bleach", Type: domain.ImportContentTypeShow, TmdbID: 30984,
	})
	s.saveImportMapping(otherUserId, &domain.ImportRequest{
		Name: "Someone Elses Show", Type: domain.ImportContentTypeShow, TmdbID: 111,
	})

	t.Run("list only returns our own mappings", func(t *testing.T) {
		got, err := s.GetImportMappings(userId)
		if err != nil {
			t.Fatalf("GetImportMappings failed: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("returned %d mappings, want 2", len(got))
		}
		for _, m := range got {
			if m.UserID != userId {
				t.Errorf("returned a mapping owned by user %d", m.UserID)
			}
		}
	})

	t.Run("update repoints the mapping", func(t *testing.T) {
		list, _ := s.GetImportMappings(userId)
		target := list[0]
		updated, err := s.UpdateImportMapping(userId, target.ID,
			domain.ImportMappingUpdateRequest{TmdbID: 999})
		if err != nil {
			t.Fatalf("UpdateImportMapping failed: %v", err)
		}
		if updated.TmdbID != 999 {
			t.Errorf("tmdb id = %d, want 999", updated.TmdbID)
		}
		found := s.findImportMapping(userId, target.Name, domain.ImportContentType(target.Type))
		if found == nil || found.TmdbID != 999 {
			t.Errorf("lookup after update did not return the new id")
		}
	})

	t.Run("update requires an id to point at", func(t *testing.T) {
		list, _ := s.GetImportMappings(userId)
		if _, err := s.UpdateImportMapping(userId, list[0].ID,
			domain.ImportMappingUpdateRequest{}); err == nil {
			t.Error("expected an error when no tmdbId or igdbId is given")
		}
	})

	t.Run("cannot update another users mapping", func(t *testing.T) {
		otherList, _ := s.GetImportMappings(otherUserId)
		if _, err := s.UpdateImportMapping(userId, otherList[0].ID,
			domain.ImportMappingUpdateRequest{TmdbID: 5}); err == nil {
			t.Fatal("was allowed to update another users mapping")
		}
		still, _ := s.GetImportMappings(otherUserId)
		if still[0].TmdbID != 111 {
			t.Errorf("another users mapping was changed to %d", still[0].TmdbID)
		}
	})

	t.Run("cannot delete another users mapping", func(t *testing.T) {
		otherList, _ := s.GetImportMappings(otherUserId)
		if err := s.DeleteImportMapping(userId, otherList[0].ID); err == nil {
			t.Fatal("was allowed to delete another users mapping")
		}
		still, _ := s.GetImportMappings(otherUserId)
		if len(still) != 1 {
			t.Errorf("another users mappings now number %d, want 1", len(still))
		}
	})

	t.Run("delete forgets the mapping", func(t *testing.T) {
		list, _ := s.GetImportMappings(userId)
		target := list[0]
		if err := s.DeleteImportMapping(userId, target.ID); err != nil {
			t.Fatalf("DeleteImportMapping failed: %v", err)
		}
		if m := s.findImportMapping(userId, target.Name, domain.ImportContentType(target.Type)); m != nil {
			t.Error("mapping was still found after being deleted")
		}
	})

	t.Run("deleting a mapping that does not exist errors", func(t *testing.T) {
		if err := s.DeleteImportMapping(userId, 99999); err == nil {
			t.Error("expected an error deleting a non existent mapping")
		}
	})
}

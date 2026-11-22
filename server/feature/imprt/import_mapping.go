package imprt

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

// Names are matched case insensitively, and with surrounding whitespace
// ignored, so that trivial differences between two exports of the same list
// do not cause a second prompt.
func normalizeMappingName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Look for a previously saved choice for this name and content type. The
// choice can be a match to some content, or a decision to ignore the name.
//
// The content type is only a preference, not a requirement. Import files
// don't always say what type an entry is (a MyAnimeList export gives OVA,
// ONA and Special entries a type we have no equivalent for, so they arrive
// with no type at all), and the type the user picked can differ from the one
// the file declared. In both of those cases the name on its own is enough,
// as long as only one saved choice uses it.
//
// Returns a nil mapping when there is nothing saved, which is the normal case
// and not an error.
func (s *Service) findImportMapping(
	userId uint,
	name string,
	contentType domain.ImportContentType,
) *entity.ImportMapping {
	normalized := normalizeMappingName(name)
	if normalized == "" {
		return nil
	}
	var mappings []entity.ImportMapping
	res := s.db.
		Where("user_id = ? AND name = ?", userId, normalized).
		Find(&mappings)
	if res.Error != nil {
		slog.Error("findImportMapping: Lookup failed",
			"user_id", userId, "name", normalized, "error", res.Error)
		return nil
	}
	// Mappings with nothing usable saved are treated as if they aren't there,
	// including when deciding whether the name is ambiguous. An ignored
	// mapping is usable despite having no ids, its answer is just that there
	// is nothing to import.
	usable := []entity.ImportMapping{}
	for _, m := range mappings {
		if m.Ignored || m.TmdbID != 0 || m.IgdbID != 0 {
			usable = append(usable, m)
		}
	}
	// A mapping saved for this exact type is always the best answer.
	if contentType != "" {
		for i := range usable {
			if usable[i].Type == string(contentType) {
				return &usable[i]
			}
		}
	}
	// Otherwise the name alone decides, but only while it can't mean more
	// than one thing.
	if len(usable) == 1 {
		return &usable[0]
	}
	return nil
}

// Remember which content a name was imported as, so a later import of the
// same name does not have to ask again.
//
// A failure to save is logged but never fails the import itself, since the
// content has already been added by this point and losing a convenience is
// not worth failing on.
func (s *Service) saveImportMapping(
	userId uint,
	ar *domain.ImportRequest,
) {
	normalized := normalizeMappingName(ar.Name)
	if normalized == "" || ar.Type == "" {
		return
	}
	if ar.TmdbID == 0 && ar.IgdbID == 0 {
		return
	}
	s.upsertImportMapping(userId, entity.ImportMapping{
		UserID: userId,
		Name:   normalized,
		Type:   string(ar.Type),
		TmdbID: ar.TmdbID,
		IgdbID: ar.IgdbID,
	})
}

// Remember that a name should be skipped, so later imports of the same name
// do not ask about it again.
//
// Unlike a match, this is saved even when the import file gave no content
// type, since the entries a user gives up on matching are often the ones we
// have no type for.
func (s *Service) saveIgnoredImportMapping(
	userId uint,
	ar *domain.ImportRequest,
) {
	normalized := normalizeMappingName(ar.Name)
	if normalized == "" {
		return
	}
	s.upsertImportMapping(userId, entity.ImportMapping{
		UserID:  userId,
		Name:    normalized,
		Type:    string(ar.Type),
		Ignored: true,
	})
}

// Write a mapping, replacing whatever was saved for the same name and type.
//
// The user can change their mind about what a name refers to, so an existing
// mapping is updated rather than left alone.
func (s *Service) upsertImportMapping(userId uint, mapping entity.ImportMapping) {
	res := s.db.
		Where("user_id = ? AND name = ? AND type = ?", userId, mapping.Name, mapping.Type).
		// Assigned as a map rather than a struct so that clearing a field
		// back to its zero value sticks, which gorm skips for struct
		// updates. Un-ignoring a name by matching it depends on it.
		Assign(map[string]any{
			"tmdb_id": mapping.TmdbID,
			"igdb_id": mapping.IgdbID,
			"ignored": mapping.Ignored,
		}).
		FirstOrCreate(&mapping)
	if res.Error != nil {
		slog.Error("upsertImportMapping: Failed to save mapping",
			"user_id", userId, "name", mapping.Name, "error", res.Error)
		return
	}
	slog.Debug("upsertImportMapping: Saved mapping",
		"user_id", userId, "name", mapping.Name, "type", mapping.Type,
		"tmdb_id", mapping.TmdbID, "igdb_id", mapping.IgdbID,
		"ignored", mapping.Ignored)
}

// All of a users saved mappings, newest first.
func (s *Service) GetImportMappings(userId uint) ([]entity.ImportMapping, error) {
	mappings := []entity.ImportMapping{}
	res := s.db.
		Where("user_id = ?", userId).
		Order("name asc").
		Find(&mappings)
	if res.Error != nil {
		slog.Error("GetImportMappings: Failed", "user_id", userId, "error", res.Error)
		return nil, errors.New("failed to get import mappings")
	}
	return mappings, nil
}

// Point an existing mapping at different content.
//
// Scoped by user id as well as mapping id so one user can't edit another
// users mapping by guessing at ids.
func (s *Service) UpdateImportMapping(
	userId uint,
	mappingId uint,
	ur domain.ImportMappingUpdateRequest,
) (entity.ImportMapping, error) {
	if ur.TmdbID == 0 && ur.IgdbID == 0 {
		return entity.ImportMapping{}, errors.New("a tmdbId or igdbId must be provided")
	}
	var mapping entity.ImportMapping
	res := s.db.
		Where("id = ? AND user_id = ?", mappingId, userId).
		Take(&mapping)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return entity.ImportMapping{}, errors.New("mapping does not exist")
		}
		slog.Error("UpdateImportMapping: Failed to find mapping",
			"user_id", userId, "mapping_id", mappingId, "error", res.Error)
		return entity.ImportMapping{}, errors.New("failed to find mapping")
	}
	mapping.TmdbID = ur.TmdbID
	mapping.IgdbID = ur.IgdbID
	// Pointing a mapping at real content is also how an ignored name is
	// un-ignored.
	mapping.Ignored = false
	// An ignored name can have been saved with no type, which would leave the
	// new id with nothing to say what it is an id of, so the picked type is
	// taken when the mapping has none of its own.
	if mapping.Type == "" && ur.Type != "" {
		mapping.Type = string(ur.Type)
	}
	if res := s.db.Save(&mapping); res.Error != nil {
		slog.Error("UpdateImportMapping: Failed to save mapping",
			"user_id", userId, "mapping_id", mappingId, "error", res.Error)
		return entity.ImportMapping{}, errors.New("failed to save mapping")
	}
	slog.Info("UpdateImportMapping: Mapping updated",
		"user_id", userId, "mapping_id", mappingId,
		"tmdb_id", mapping.TmdbID, "igdb_id", mapping.IgdbID)
	return mapping, nil
}

// Forget every saved mapping, so all names are searched for again on the next
// import. Only ever touches the given users mappings.
func (s *Service) DeleteAllImportMappings(userId uint) (int64, error) {
	// Deleted outright rather than soft deleted. A soft deleted row still
	// occupies its spot in the unique index, so saving the same name again
	// afterwards would fail, and a forgotten choice is not worth keeping.
	res := s.db.
		Unscoped().
		Where("user_id = ?", userId).
		Delete(&entity.ImportMapping{})
	if res.Error != nil {
		slog.Error("DeleteAllImportMappings: Failed",
			"user_id", userId, "error", res.Error)
		return 0, errors.New("failed to delete mappings")
	}
	slog.Info("DeleteAllImportMappings: Mappings deleted",
		"user_id", userId, "num_deleted", res.RowsAffected)
	return res.RowsAffected, nil
}

// Forget a mapping, so the name is searched for again on the next import.
func (s *Service) DeleteImportMapping(userId uint, mappingId uint) error {
	// Deleted outright, for the reason given in DeleteAllImportMappings.
	res := s.db.
		Unscoped().
		Where("id = ? AND user_id = ?", mappingId, userId).
		Delete(&entity.ImportMapping{})
	if res.Error != nil {
		slog.Error("DeleteImportMapping: Failed",
			"user_id", userId, "mapping_id", mappingId, "error", res.Error)
		return errors.New("failed to delete mapping")
	}
	if res.RowsAffected == 0 {
		return errors.New("mapping does not exist")
	}
	slog.Info("DeleteImportMapping: Mapping deleted",
		"user_id", userId, "mapping_id", mappingId)
	return nil
}

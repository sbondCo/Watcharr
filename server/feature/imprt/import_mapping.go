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

// Look for a previously saved choice for this name and content type.
//
// Returns a nil mapping when there is nothing saved, which is the normal case
// and not an error.
func (s *Service) findImportMapping(
	userId uint,
	name string,
	contentType domain.ImportContentType,
) *entity.ImportMapping {
	normalized := normalizeMappingName(name)
	if normalized == "" || contentType == "" {
		return nil
	}
	var mapping entity.ImportMapping
	res := s.db.
		Where("user_id = ? AND name = ? AND type = ?", userId, normalized, string(contentType)).
		Take(&mapping)
	if res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			slog.Error("findImportMapping: Lookup failed",
				"user_id", userId, "name", normalized, "error", res.Error)
		}
		return nil
	}
	if mapping.TmdbID == 0 && mapping.IgdbID == 0 {
		// Nothing usable saved, treat it as if we found nothing.
		return nil
	}
	return &mapping
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
	mapping := entity.ImportMapping{
		UserID: userId,
		Name:   normalized,
		Type:   string(ar.Type),
		TmdbID: ar.TmdbID,
		IgdbID: ar.IgdbID,
	}
	// The user can change their mind about what a name refers to, so an
	// existing mapping is updated rather than left alone.
	res := s.db.
		Where("user_id = ? AND name = ? AND type = ?", userId, normalized, string(ar.Type)).
		Assign(entity.ImportMapping{TmdbID: ar.TmdbID, IgdbID: ar.IgdbID}).
		FirstOrCreate(&mapping)
	if res.Error != nil {
		slog.Error("saveImportMapping: Failed to save mapping",
			"user_id", userId, "name", normalized, "error", res.Error)
		return
	}
	slog.Debug("saveImportMapping: Saved mapping",
		"user_id", userId, "name", normalized, "type", ar.Type,
		"tmdb_id", ar.TmdbID, "igdb_id", ar.IgdbID)
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

// Forget a mapping, so the name is searched for again on the next import.
func (s *Service) DeleteImportMapping(userId uint, mappingId uint) error {
	res := s.db.
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

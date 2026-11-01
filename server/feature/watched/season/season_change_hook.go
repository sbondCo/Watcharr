package season

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
)

// Called after season status is changed.
// Automates setting other statuses.
func (s *Service) hookStatusChanged(
	userID uint,
	watchedID uint,
	seasonNum int,
	newSeasonStatus entity.WatchedStatus,
) domain.SeasonStatusChangedHookResponse {
	userSettings, err := s.userProvider.UserGetSettings(userID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get user settings! Continuing.",
			"error", err)
	} else {
		if !*userSettings.AutomateShowStatuses {
			slog.Debug("hookStatusChanged: User has AutomateShowStatuses disabled. Skipping hook.",
				"user_id", userID)
			return domain.SeasonStatusChangedHookResponse{}
		}
	}

	hookResponse := domain.SeasonStatusChangedHookResponse{}

	// Get watched entry.
	// NOTE: Content type is validated from caller (SetWatchedSeason), so its
	// safe to not validate that here.
	watchedEntry, err := s.wp.GetWatchedItemById(userID, watchedID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get watched entry!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get watched entry")
		return hookResponse
	}

	newShowStatus, addedActivity, err := s.hookStatusChangedSetShowStatus(
		userID,
		watchedID,
		watchedEntry.Content.TmdbID,
		watchedEntry.Status,
		seasonNum,
		newSeasonStatus,
	)
	if err != nil {
		slog.Error("hookStatusChanged: Setting show status Failed!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to update show")
	} else {
		hookResponse.NewShowStatus = newShowStatus
		if addedActivity.ID != 0 {
			hookResponse.AddedActivities = append(
				hookResponse.AddedActivities, addedActivity)
		}
	}

	return hookResponse
}

// Set top level watched entry status.
// Returns newShowStatus, addedActivity.
func (s *Service) hookStatusChangedSetShowStatus(
	userID uint,
	watchedID uint,
	tmdbID int,
	showWatchedStatus entity.WatchedStatus,
	seasonNum int,
	newSeasonStatus entity.WatchedStatus,
) (entity.WatchedStatus, entity.Activity, error) {
	var newShowStatus entity.WatchedStatus
	var reason string

	// If newSeasonStatus is WATCHING or FINISHED & the show is PLANNED, then
	// show status should be set to WATCHING (unless checks below change that).
	if (newSeasonStatus == entity.WATCHING || newSeasonStatus == entity.FINISHED) &&
		(showWatchedStatus == "" || showWatchedStatus == entity.PLANNED) {
		newShowStatus = entity.WATCHING
		reason = fmt.Sprintf(
			"Season %d was set to %s.",
			seasonNum,
			newSeasonStatus,
		)
	}

	// If all seasons are finished, the show status needs to be set to:
	// - PLANNED (if show is continuing)
	// - FINISHED (if show is ended/cancelled)
	if newSeasonStatus == entity.FINISHED {
		// NOTE: We only need to do this check if newSeasonStatus is FINISHED, since
		// if it was anything else, not all seasons can be FINISHED.
		allSeasonsFinished, err := s.allSeasonsCompletedForShow(
			userID, watchedID, tmdbID)
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Seasons completion check failed.",
				"error", err)
			return "", entity.Activity{}, err
		}
		if allSeasonsFinished {
			content, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
				ID: strconv.Itoa(tmdbID),
			})
			if err != nil {
				slog.Error("hookStatusChangedSetShowStatus: Couldn't get ShowDetails",
					"error", err)
				return "", entity.Activity{}, err
			}

			if content.Status == tmdb.ShowStatusCanceled ||
				content.Status == tmdb.ShowStatusEnded {

				newShowStatus = entity.FINISHED
				reason = "All seasons were completed and the show is not continuing."
			} else {
				newShowStatus = entity.PLANNED
				reason = "All seasons were completed and the show is continuing."
			}
		}
	}

	// Finally save the newShowStatus

	if newShowStatus == "" || reason == "" {
		slog.Info("hookStatusChangedSetShowStatus: No new status or reason set. Not saving.")
		return "", entity.Activity{}, nil
	} else if newShowStatus == showWatchedStatus {
		slog.Info("hookStatusChangedSetShowStatus: New status is same as current. Not saving.")
		return "", entity.Activity{}, nil
	}

	res := s.db.
		Model(&entity.Watched{}).
		Where("user_id = ? AND id = ?", userID, watchedID).
		Update("status", newShowStatus)
	if res.Error != nil {
		slog.Error("hookStatusChangedSetShowStatus: Failed to update show status!",
			"error", res.Error)
		return "", entity.Activity{}, errors.New("failed to update db")
	}
	if res.RowsAffected == 0 {
		slog.Error("hookStatusChangedSetShowStatus: Nothing matched to update.")
		return "", entity.Activity{}, errors.New("watched entry does not exist")
	}

	countAsPlay := false
	if newShowStatus == entity.FINISHED {
		countAsPlay = true

		nextWSeason, err := s.GetWatchedSeason(userID, watchedID, seasonNum+1)
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Get next watched season failed, assuming doesn't exist.",
				"error", err)
		} else if nextWSeason != nil {
			// If the show has a next season and we a watched entry for it, then
			// countAsPlay must be false, because it should only be true for
			// when the last season is set to finished.
			// This is a protection for second complete watch of a series
			// where the user could start and finish S1, meanwhile all other
			// seasons are already/still FINISHED from a previous complete watch.
			// Without this check, we'd count another play, but in reality, we
			// should only count the play when all the rest of the seasons are
			// re-watched too.
			countAsPlay = false
			slog.Debug("hookStatusChangedSetShowStatus: Overrid count_as_play to false. We found a next watched season entry.")
		}
	}
	addedActivity, _ := activity.
		NewCreator(
			s.db,
			userID,
			watchedID,
			entity.STATUS_CHANGED,
			countAsPlay,
			entity.ActivityCreatedByWatcharr,
		).
		SetDataJSON(map[string]any{
			"status": newShowStatus,
		}).
		SetReason(reason).
		Create()

	return newShowStatus, addedActivity, nil
}

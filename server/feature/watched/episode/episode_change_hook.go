package episode

import (
	"fmt"
	"log/slog"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

// Called after episode status is changed.
// Automates setting other statuses.
func (s *Service) hookStatusChanged(
	userID uint,
	watchedID uint,
	seasonNum int,
	episodeNum int,
	newEpStatus entity.WatchedStatus,
) domain.EpisodeStatusChangedHookResponse {
	userSettings, err := s.userProvider.UserGetSettings(userID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get user settings! Continuing.",
			"error", err)
	} else {
		if !*userSettings.AutomateShowStatuses {
			slog.Debug("hookStatusChanged: User has AutomateShowStatuses disabled. Skipping hook.",
				"user_id", userID)
			return domain.EpisodeStatusChangedHookResponse{}
		}
	}

	hookResponse := domain.EpisodeStatusChangedHookResponse{}

	// Small local add activity helper.
	addHookActivity := func(aType entity.ActivityType, data map[string]any, reason string) {
		addedActivity, _ := activity.
			NewCreator(
				s.db,
				userID,
				watchedID,
				aType,
				false,
				entity.ActivityCreatedByWatcharr,
			).
			SetDataJSON(data).
			SetReason(reason).
			Create()
		hookResponse.AddedActivities = append(hookResponse.AddedActivities, addedActivity)
	}

	// Get watched entry.
	// NOTE: Content type is validated from caller (SetWatchedEpisode), so its
	// safe to not validate that here.
	watchedEntry, err := s.wp.GetWatchedItemById(userID, watchedID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get watched entry!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get watched entry")
		return hookResponse
	}

	// Set the episodes *season* status.
	// SetWatchedSeason is idempotent so we don't need to worry if nothing will
	// change.
	seasonSetResp, err := s.hookStatusChangedSetSeasonStatus(
		userID,
		watchedID,
		watchedEntry.Content.TmdbID,
		seasonNum,
		episodeNum,
		newEpStatus,
	)
	if err != nil {
		slog.Error("hookStatusChanged: Setting season status Failed!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to update season")
	} else {
		hookResponse.WatchedSeason = &seasonSetResp.WatchedSeason
		hookResponse.AddedActivities = append(
			hookResponse.AddedActivities, seasonSetResp.AddedActivities...)
	}

	// Update main show watched entry status.
	// Show status shouldn't be empty, but watevs, handle it just incase
	if watchedEntry.Status == "" || watchedEntry.Status == entity.PLANNED {
		watchedEntry.Status = entity.WATCHING
		if res := s.db.Save(watchedEntry); res.Error != nil {
			slog.Error("hookStatusChanged: Failed to update show status!",
				"error", res.Error)
		} else {
			hookResponse.NewShowStatus = watchedEntry.Status
			addHookActivity(
				entity.STATUS_CHANGED,
				map[string]any{
					"status": watchedEntry.Status,
				},
				fmt.Sprintf(
					"S%dE%d was set to %s.",
					seasonNum,
					episodeNum,
					newEpStatus,
				),
			)
		}
	}

	return hookResponse
}

// Sets season status depending on newEpStatus and if all episodes in the season
// have been completed.
func (s *Service) hookStatusChangedSetSeasonStatus(
	userID uint,
	watchedID uint,
	tmdbID int,
	seasonNum int,
	episodeNum int,
	newEpStatus entity.WatchedStatus,
) (domain.WatchedSeasonSetResponse, error) {
	seasonNewStatus := newEpStatus

	if newEpStatus == entity.FINISHED || newEpStatus == entity.DROPPED {
		seasonNewStatus = entity.WATCHING
	}

	if newEpStatus == entity.FINISHED {
		// If all episodes are FINISHED, override seasonNewStatus to FINISHED.
		// NOTE: We only need to do this check if newEpStatus is FINISHED, since
		// if it was anything else, not all episodes can be FINISHED. This is a
		// "big" call, so its worth only running it if necessary.
		allEpsFinished, err := s.allEpisodesCompletedForSeason(
			userID, watchedID, tmdbID, seasonNum)
		if err != nil {
			slog.Error("hookStatusChangedSetSeasonStatus: Episode completion check failed.",
				"error", err)
			return domain.WatchedSeasonSetResponse{}, err
		}
		if allEpsFinished {
			seasonNewStatus = entity.FINISHED
		}
	}

	return s.wsp.SetWatchedSeason(userID, domain.WatchedSeasonSetRequest{
		WatchedID:    watchedID,
		SeasonNumber: seasonNum,
		Status:       seasonNewStatus,

		ActivityReason: fmt.Sprintf(
			"Episode %d was set to %s.",
			episodeNum,
			newEpStatus,
		),
		ActivityCreatedBy: entity.ActivityCreatedByWatcharr,
	})
}

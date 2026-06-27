package watchedutil_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/watched/watchedutil"
)

func TestGetLatestWatchedInTv(t *testing.T) {
	watchedSeasons := []entity.WatchedSeason{
		{
			SeasonNumber: 1,
			Status:       entity.FINISHED,
		},
		{
			SeasonNumber: 2,
			Status:       entity.WATCHING,
		},
		{
			SeasonNumber: 3,
			Status:       entity.FINISHED,
		},
	}
	watchedEps := []entity.WatchedEpisode{
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 60},
			EpisodeNumber:  1,
			SeasonNumber:   1,
			Status:         entity.FINISHED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 70},
			EpisodeNumber:  5,
			SeasonNumber:   2,
			Status:         entity.FINISHED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 72},
			EpisodeNumber:  6,
			SeasonNumber:   3,
			Status:         entity.DROPPED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 90},
			EpisodeNumber:  2,
			SeasonNumber:   3,
			Status:         entity.FINISHED,
		},
	}
	resp := watchedutil.GetLatestWatchedInTv(watchedSeasons, watchedEps)
	want := "S2E5"
	if resp != want {
		t.Errorf("%s != %s", resp, want)
	}
}

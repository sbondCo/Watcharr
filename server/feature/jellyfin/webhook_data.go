package jellyfin

// The webhook data format expected from Jellyfin.
type WebhookData struct {
	//
	// Server
	//

	NotificationType WebhookType
	ServerID         string `json:"ServerId"`

	// UserDataSaved events include a save reason.
	SaveReason WebhookUserDataSaveReason

	//
	// User
	//

	NotificationUsername string
	UserID               string `json:"UserId"`

	//
	// BaseItem
	//

	Name           string
	ItemType       WebhookItemType
	ItemID         string `json:"ItemId"`
	ProviderIMDBID string `json:"Provider_imdb"`
	ProviderTVDBID string `json:"Provider_tvdb"`
	ProviderTMDBID string `json:"Provider_tmdb"`

	//
	// Season
	//

	SeriesName string
	SeriesID   string `json:"SeriesId"`
	// Is a ptr so we can tell when its `unset` vs `season 0`
	SeasonNumber *int

	//
	// Episode
	//

	EpisodeNumber int

	// Only for `NotificationType == PlaybackStop`.
	// Tells us if this Playback resulted in a completed play.
	PlayedToCompletion bool
	// Other NotificationTypes.
	// I believe this comes from the existing value of if the item is played in
	// jellyfins database, so isn't specific to the session (so always prefer
	// PlayedToCompletion when we have it in PlaybackStop event).
	Played bool
}

// Webhook notification types that we use/support.
// https://github.com/jellyfin/jellyfin-plugin-webhook/blob/master/Jellyfin.Plugin.Webhook/Destinations/NotificationType.cs
type WebhookType string

const (
	// When user starts playback on jellyfin.
	WebhookTypePlaybackStart WebhookType = "PlaybackStart"
	// User stops watching (finished media / quit half way through).
	WebhookTypePlaybackStop WebhookType = "PlaybackStop"
	// User data is updated (eg: Marks movie finished in webui)
	WebhookTypeUserDataSaved WebhookType = "UserDataSaved"
)

// Webhook ItemTypes that we support.
type WebhookItemType string

const (
	WebhookItemTypeMovie WebhookItemType = "Movie"
	// WebhookItemTypeSeries  WebhookItemType = "Series"
	// WebhookItemTypeSeason  WebhookItemType = "Season"
	WebhookItemTypeEpisode WebhookItemType = "Episode"
)

type WebhookUserDataSaveReason string

const (
	// User toggled 'Played' on the item in jellyfin web ui.
	WebhookUserDataSaveReasonTogglePlayed WebhookUserDataSaveReason = "TogglePlayed"
)

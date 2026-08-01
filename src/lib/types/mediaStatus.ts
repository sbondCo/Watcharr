/**
 * Media statuses.
 */

export enum MediaStatusShow {
	ReturningSeries = "Returning Series",
	Planned = "Planned",
	InProduction = "In Production",
	Ended = "Ended",
	Canceled = "Canceled",
	Pilot = "Pilot",
}

export enum MediaStatusMovie {
	Rumored = "Rumored",
	Planned = "Planned",
	InProduction = "In Production",
	PostProduction = "Post Production",
	Released = "Released",
	Canceled = "Canceled",
}

/**
 * Taken from what `https://api.igdb.com/v4/game_statuses` returned at the time.
 */
export enum MediaStatusGame {
	Released = "Released",
	Alpha = "Alpha",
	Beta = "Beta",
	EarlyAccess = "Early Access",
	Offline = "Offline",
	Cancelled = "Cancelled",
	Rumored = "Rumored",
	Delisted = "Delisted",
}

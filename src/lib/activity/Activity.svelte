<script lang="ts">
	import {
		type Activity,
		ActivityCreatedBy,
		type WatchedStatus,
	} from "@/types";
	import { seasonAndEpToReadable, statusLabel } from "../util/helpers";
	import { formatDate, formatTime, ordinalDay } from "../util/date";
	import { getLanguage, t } from "../i18n/index.svelte";
	import ActivityEditor from "./ActivityEditor.svelte";
	import ActivityIcon from "./ActivityIcon.svelte";
	import CountBadge from "../generic/CountBadge.svelte";

	interface Props {
		activity: Activity[] | undefined;
		onRemoved: (activity: Activity) => void;
	}

	let { activity = undefined, onRemoved }: Props = $props();

	let clickedActivity: Activity | undefined = $state();
	let groupedActivities: { [index: string]: Activity[] } = $derived(
		getGroupedActivity(activity),
	);
	let count = $derived(activity ? activity.length : 0);

	function getMsg(a: Activity) {
		switch (a?.type) {
			case "ADDED_WATCHED":
				if (a.data) {
					const data = JSON.parse(a.data);
					const status = data?.status ? statusLabel(data.status) : undefined;
					if (status && data?.rating) {
						return t("activity.msg.addedToWatchedListAsWithRating", {
							status,
							rating: data.rating,
						});
					} else if (status) {
						return t("activity.msg.addedToWatchedListAs", { status });
					} else if (data?.rating) {
						return t("activity.msg.addedToWatchedListWithRating", {
							rating: data.rating,
						});
					}
				}
				return t("activity.msg.addedToWatchedList");
			case "REMOVED_WATCHED":
				return t("activity.msg.removedFromWatchedList");
			case "RATING_CHANGED":
				if (a.data) {
					return t("activity.msg.ratingChangedTo", { rating: a.data });
				}
				return t("activity.msg.ratingChanged");
			case "STATUS_CHANGED":
				if (a.data) {
					let status = a.data;
					if (a.data.startsWith("{")) {
						const data = JSON.parse(a.data);
						status = data.status;
					}
					return t("activity.msg.statusChangedTo", {
						status: statusLabel(status as WatchedStatus),
					});
				}
				return t("activity.msg.statusChanged");
			case "THOUGHTS_CHANGED":
				return t("activity.msg.thoughtsChanged");
			case "THOUGHTS_REMOVED":
				return t("activity.msg.thoughtsRemoved");
			case "IMPORTED_RATING":
				if (a.data) {
					const data = JSON.parse(a.data);
					if (data.rating) {
						return t("activity.msg.ratingChangedTo", { rating: data.rating });
					} else {
						return t("activity.msg.addedToWatchlistNoRating");
					}
				}
				return t("activity.msg.importedRating");
			case "IMPORTED_ADDED_WATCHED":
				return t("activity.msg.importedWatchDate");
			case "SEASON_ADDED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.seasonAddedAs", {
						season: data.season,
						status: statusLabel(data.status),
					});
				}
				return t("activity.msg.seasonAdded");
			case "SEASON_RATING_CHANGED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.seasonRatingChangedTo", {
						season: data.season,
						rating: data.rating,
					});
				}
				return t("activity.msg.seasonRatingChanged");
			case "SEASON_STATUS_CHANGED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.seasonStatusChangedTo", {
						season: data.season,
						status: statusLabel(data.status),
					});
				}
				return t("activity.msg.seasonStatusChanged");
			case "SEASON_REMOVED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.seasonStatusRemoved", {
						season: data.season,
					});
				}
				return t("activity.msg.seasonRemoved");
			case "EPISODE_ADDED":
				if (a.data) {
					const data = JSON.parse(a.data);
					const episode = seasonAndEpToReadable(data.season, data.episode);
					if (data.status) {
						return t("activity.msg.episodeAddedAs", {
							episode,
							status: statusLabel(data.status),
						});
					} else if (data.rating) {
						return t("activity.msg.episodeAddedWithRating", {
							episode,
							rating: data.rating,
						});
					}
					return t("activity.msg.episodeAddedNamed", { episode });
				}
				return t("activity.msg.episodeAdded");
			case "EPISODE_RATING_CHANGED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.episodeRatingChangedTo", {
						episode: seasonAndEpToReadable(data.season, data.episode),
						rating: data.rating,
					});
				}
				return t("activity.msg.episodeRatingChanged");
			case "EPISODE_STATUS_CHANGED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.episodeStatusChangedTo", {
						episode: seasonAndEpToReadable(data.season, data.episode),
						status: statusLabel(data.status),
					});
				}
				return t("activity.msg.episodeStatusChanged");
			case "EPISODE_REMOVED":
				if (a.data) {
					const data = JSON.parse(a.data);
					return t("activity.msg.episodeRemovedNamed", {
						episode: seasonAndEpToReadable(data.season, data.episode),
					});
				}
				return t("activity.msg.episodeRemoved");
			default:
				return a.type;
		}
	}

	function toDayTime(d: Date) {
		return t("activity.dayAtTime", {
			day: ordinalDay(d),
			time: formatTime(d),
		});
	}

	/**
	 * Get what will be the visible date the activity was 'created'.
	 * @returns customDate if defined or createdAt if not.
	 */
	function getCreatedAtVis(a: Activity) {
		return Date.parse(a.customDate ?? a.createdAt);
	}

	function getGroupedActivity(activities?: Activity[]) {
		activities = activities?.filter((a) => a.type);
		const a = activities?.sort(
			(a, b) => getCreatedAtVis(b) - getCreatedAtVis(a),
		);
		let grouped: { [index: string]: Activity[] } = {};
		if (a) {
			for (let i = 0; i < a.length; i++) {
				const activity = a[i];
				const date = new Date(getCreatedAtVis(activity));
				const month = formatDate(date, { month: "long", year: "numeric" });
				const key =
					month.charAt(0).toLocaleUpperCase(getLanguage()) + month.slice(1);
				if (grouped[key]) {
					grouped[key].push(activity);
				} else {
					grouped[key] = [activity];
				}
			}
		}
		return grouped;
	}

	function openEditor(a: Activity) {
		clickedActivity = a;
		return;
	}
</script>

{#if clickedActivity}
	<ActivityEditor
		activity={clickedActivity}
		activityMessage={getMsg(clickedActivity)}
		onClose={() => (clickedActivity = undefined)}
		onUpdated={(activityId, updatedActivity) => {
			if (!activity) {
				console.error(
					"ActivityEditor->onUpdated: 'activity' doesn't exist somehow..",
				);
				return;
			}
			const ai = activity.findIndex((a) => a.id === activityId);
			activity[ai] = updatedActivity;
		}}
		onRemoved={(a) => {
			if (!activity) {
				console.error(
					"ActivityEditor->onRemoved: 'activity' doesn't exist somehow..",
				);
				return;
			}
			onRemoved(a);
		}}
	/>
{/if}

<div class="activity">
	<div class="title">
		<h2>{t("activity.title")}</h2>
		<CountBadge {count} />
	</div>
	{#if groupedActivities && Object.keys(groupedActivities).length > 0}
		<ul>
			{#each Object.keys(groupedActivities) as k (k)}
				<h3>{k}</h3>

				{#each groupedActivities[k] as a (a.id)}
					{@const d = new Date(getCreatedAtVis(a))}
					<li>
						<button class="plain" onclick={() => openEditor(a)}>
							<span title={formatDate(d, { dateStyle: "full" })}
								>{toDayTime(d)}</span
							>
							<span>{getMsg(a)}</span>
						</button>
						{#if a.createdBy == ActivityCreatedBy.Watcharr}
							<ActivityIcon
								icon="sparkles"
								tooltipText={a.reason
									? t("activity.automatedReason", { reason: a.reason })
									: t("activity.completedByAutomation")}
							/>
						{:else if a.createdBy == ActivityCreatedBy.GenericImport}
							<ActivityIcon
								icon="calendar"
								tooltipText={t("activity.imported")}
							/>
						{:else if a.createdBy == ActivityCreatedBy.JellyfinImport}
							<ActivityIcon
								icon="jellyfin"
								tooltipText={t("activity.importFrom", { service: "Jellyfin" })}
							/>
						{:else if a.createdBy == ActivityCreatedBy.JellyfinWebhook}
							<ActivityIcon
								icon="jellyfin"
								tooltipText={t("activity.jellyfinWebhook")}
							/>
						{:else if a.createdBy == ActivityCreatedBy.PlexImport}
							<ActivityIcon
								icon="plex"
								tooltipText={t("activity.importFrom", { service: "Plex" })}
							/>
						{/if}
						{#if a.countAsPlay}
							<ActivityIcon
								icon="play"
								tooltipText={t("activity.countsAsPlay")}
							/>
						{/if}
					</li>
				{/each}
			{/each}
		</ul>
	{:else}
		<span>{t("activity.noActivity")}</span>
	{/if}
</div>

<style lang="scss">
	.activity {
		width: 100%;

		div.title {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 10px;
		}

		ul {
			display: flex;
			flex-flow: column;
			gap: 8px;
			margin-top: 8px;
			margin-left: calc(30px + 8px);
			list-style: none;
			max-height: 250px;
			overflow-y: auto;
			overflow-x: hidden;

			h3 {
				position: sticky;
				top: 0;
				font-size: 16px;
				font-family:
					sans-serif,
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				padding-bottom: 1px;
				background-color: $bg-color;
			}

			li {
				display: flex;
				flex-flow: row;
				gap: 8px;
				align-items: center;

				button {
					all: unset;
					display: flex;
					flex-flow: row;
					align-items: center;
					flex-shrink: 1;
					gap: 8px;
					width: max-content;
					min-width: 0;
					max-width: 100%;
					cursor: pointer;
				}

				span {
					width: max-content;
					margin-left: 0px;

					&:first-child {
						min-width: max-content;
					}

					&:last-of-type {
						background-color: $accent-color;
						color: $text-color;
						border-radius: 8px;
						padding: 10px 12px;
					}
				}
			}
		}

		span {
			margin-left: calc(30px);
			width: 100%;
			display: flex;
			justify-content: center;
		}
	}

	h2 {
		font-size: 30px;
		font-weight: bold;
		margin-left: 30px;
	}
</style>

<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import Stats from "@/lib/stats/Stats.svelte";
	import { req } from "@/lib/util/api";
	import { formatShortDate } from "@/lib/util/date";
	import { store } from "@/store.svelte";
	import { type Image, type Profile } from "@/types";
	import { notify } from "@/lib/util/notify";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import { ReqerError } from "@/lib/util/fetch";
	import { resolve } from "$app/paths";

	let user = $derived(store.userInfo);
	let getProfilePromise = $state(getProfile());

	async function getProfile() {
		return await req.get<Profile>(`/profile`);
	}

	function updateBio(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLTextAreaElement },
	) {
		const newBio = ev?.currentTarget?.value;
		if (typeof newBio !== "string") {
			console.warn("updateBio called without any value", newBio);
			return;
		}
		const nid = notify({ text: t("profile.updatingBio"), type: "loading" });
		req
			.post("/user/bio", { newBio: newBio })
			.then(() => {
				if (user) {
					user.bio = newBio;
					notify({ id: nid, text: t("profile.updatedBio"), type: "success" });
				}
			})
			.catch((err) => {
				notify({
					id: nid,
					text: ReqerError.getMsg(err, t("profile.failedToUpdateBio")),
					type: "error",
				});
			});
	}

	function avatarDropped(ev: Event) {
		const files = (ev.currentTarget as HTMLInputElement)?.files;
		if (!files || files?.length <= 0) {
			console.error("avatarDropped: no file found");
			return;
		}
		const nid = notify({ text: t("profile.uploadingAvatar"), type: "loading" });
		const fd = new FormData();
		fd.append("avatar", files[0]);
		req
			.post<Image>("/user/avatar", fd)
			.then((r) => {
				if (user) {
					user.avatar = r;
					notify({
						id: nid,
						text: t("profile.avatarUploaded"),
						type: "success",
					});
				}
			})
			.catch((err) => {
				console.error("uploading avatar failed", err);
				notify({
					id: nid,
					text: ReqerError.getMsg(err, t("profile.failedToUploadAvatar")),
					type: "error",
				});
			});
	}

	/**
	 * Takes in number of minutes and converts to readable.
	 * eg into months, weeks, days, hours and minutes.
	 */
	function toFormattedTimeLong(m: number) {
		// Considers a 30 days long month
		const countInMinutes: [string, number][] = [
			["time.months", 43200],
			["time.weeks", 10080],
			["time.days", 1440],
			["time.hours", 60],
		];
		const parts: string[] = [];
		let tmp;
		for (const [key, minutes] of countInMinutes) {
			tmp = Math.floor(m / minutes);

			// Ignore fields with fewer than 1 unit
			if (tmp) parts.push(t(key, { count: tmp }));
			m -= tmp * minutes;
		}
		if (parts.length <= 0) {
			return t("time.hours", { count: 0 });
		}
		return parts.join(", ");
	}
</script>

<svelte:head>
	<title>{t("profile.pageTitle")}</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="user-basic-info">
			<UserAvatar img={user?.avatar} {avatarDropped} />
			<div>
				<h2 title={user?.username}>
					<span style="font-weight: normal; font-variant: all-small-caps;"
						>{t("profile.hey")}</span
					>
					{user?.username}
				</h2>
				<textarea
					rows="1"
					placeholder={t("profile.myBio")}
					onblur={updateBio}
					value={user?.bio}></textarea>
			</div>
		</div>

		<Stats>
			{#await getProfilePromise}
				<Spinner />
			{:then profile}
				<Stat
					name={t("profile.joined")}
					value={formatShortDate(new Date(profile.joined))}
				/>
				<Stat
					name={t("profile.moviesWatched")}
					value={profile.moviesWatched}
					large
				/>
				<Stat
					name={t("profile.showsWatched")}
					value={profile.showsWatched}
					large
				/>
				<Stat
					name={t("profile.watchingMovies")}
					value={toFormattedTimeLong(profile.moviesWatchedRuntime)}
				/>
				<Stat
					name={t("profile.watchingShows")}
					value={toFormattedTimeLong(profile.showsWatchedRuntime)}
					disc={t("profile.veryInaccurate")}
				/>
			{:catch err}
				<Error error={err} pretty={t("profile.failedToGetStats")} />
			{/await}
		</Stats>

		<div style="margin-top: 10px;">
			<!-- Temporary link to settings, since we just moved them to
			 another page, can remove this later. -->
			<a href={resolve("/settings")}>{t("profile.lookingForSettings")}</a>
		</div>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			min-width: 400px;
			max-width: 400px;
			overflow: hidden;

			h2 {
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
			}

			& > div:not(:first-of-type) {
				margin-top: 30px;
			}

			@media screen and (max-width: 440px) {
				width: 100%;
				min-width: unset;
			}
		}
	}

	.user-basic-info {
		display: flex;
		gap: 20px;

		& > div {
			display: flex;
			flex-flow: column;
			gap: 5px;
			width: 100%;
			overflow: hidden;

			textarea {
				resize: none;

				&:not(:focus) {
					border: 0;
					padding: 0;
					height: 32px;
				}
			}
		}
	}
</style>

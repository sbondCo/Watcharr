<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import Stats from "@/lib/stats/Stats.svelte";
	import { req } from "@/lib/util/api";
	import { getOrdinalSuffix, monthsShort } from "@/lib/util/helpers";
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

	function formatDate(d: Date) {
		return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${
			monthsShort[d.getMonth()]
		} ${d.getFullYear()}`;
	}

	function updateBio(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLTextAreaElement },
	) {
		const newBio = ev?.currentTarget?.value;
		if (typeof newBio !== "string") {
			console.warn("updateBio called without any value", newBio);
			return;
		}
		const nid = notify({ text: "Updating Bio", type: "loading" });
		req
			.post("/user/bio", { newBio: newBio })
			.then(() => {
				if (user) {
					user.bio = newBio;
					notify({ id: nid, text: "Updated Bio", type: "success" });
				}
			})
			.catch((err) => {
				notify({
					id: nid,
					text: ReqerError.getMsg(err, "Failed to update bio"),
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
		const nid = notify({ text: "Uploading avatar", type: "loading" });
		const fd = new FormData();
		fd.append("avatar", files[0]);
		req
			.post<Image>("/user/avatar", fd)
			.then((r) => {
				if (user) {
					user.avatar = r;
					notify({ id: nid, text: "Avatar Uploaded", type: "success" });
				}
			})
			.catch((err) => {
				console.error("uploading avatar failed", err);
				notify({
					id: nid,
					text: ReqerError.getMsg(err, "Failed to upload avatar"),
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
		const countInMinutes = [
			["month", 43200],
			["week", 10080],
			["day", 1440],
			["hour", 60],
		];
		let ansString = "";
		let tmp;
		for (const c of countInMinutes) {
			tmp = Math.floor(m / (c[1] as number));

			// Ignore fields with fewer than 1 unit
			if (tmp) ansString += `${tmp} ${c[0]}${tmp >= 2 ? "s, " : ", "}`;
			m -= tmp * (c[1] as number);
		}
		if (!ansString) {
			return "0 hours";
		}
		return ansString.slice(0, -2);
	}
</script>

<svelte:head>
	<title>My Profile</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="user-basic-info">
			<UserAvatar img={user?.avatar} {avatarDropped} />
			<div>
				<h2 title={user?.username}>
					<span style="font-weight: normal; font-variant: all-small-caps;"
						>Hey</span
					>
					{user?.username}
				</h2>
				<textarea
					rows="1"
					placeholder="my bio"
					onblur={updateBio}
					value={user?.bio}></textarea>
			</div>
		</div>

		<Stats>
			{#await getProfilePromise}
				<Spinner />
			{:then profile}
				<Stat name="Joined" value={formatDate(new Date(profile.joined))} />
				<Stat name="Movies Watched" value={profile.moviesWatched} large />
				<Stat name="Shows Watched" value={profile.showsWatched} large />
				<Stat
					name="Watching Movies"
					value={toFormattedTimeLong(profile.moviesWatchedRuntime)}
				/>
				<Stat
					name="Watching Shows"
					value={toFormattedTimeLong(profile.showsWatchedRuntime)}
					disc="This is very inaccurate 🚀"
				/>
			{:catch err}
				<Error error={err} pretty="Failed to get stats!" />
			{/await}
		</Stats>

		<div style="margin-top: 10px;">
			<!-- Temporary link to settings, since we just moved them to
			 another page, can remove this later. -->
			<a href={resolve("/settings")}>Looking for settings?</a>
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

<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Trans from "@/lib/i18n/Trans.svelte";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import type {
		Content,
		RadarrSettings,
		ServerConfig,
		SonarrSettings,
	} from "@/types";
	import SonarrModal from "./modals/SonarrModal.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingButton from "@/lib/settings/SettingButton.svelte";
	import RadarrModal from "./modals/RadarrModal.svelte";
	import { getServerFeatures, req } from "@/lib/util/api";
	import Stats from "@/lib/stats/Stats.svelte";
	import Error from "@/lib/Error.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import TwitchModal from "./modals/TwitchModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import TaskScheduleModal from "./modals/TaskScheduleModal.svelte";
	import TrustedHeaderAuthModal from "./modals/TrustedHeaderAuthModal.svelte";
	import { resolve } from "$app/paths";

	let serverConfig: ServerConfig | undefined = $state();
	let jellyfinOrEmby = $derived(serverConfig?.USE_EMBY ? "Emby" : "Jellyfin");
	let sonarrModalOpen = $state(false);
	let sonarrServerEditing: SonarrSettings | undefined = $state();
	let sonarrModalEditing = $state(false);
	let radarrModalOpen = $state(false);
	let radarrServerEditing: RadarrSettings | undefined = $state();
	let radarrModalEditing = $state(false);
	let twitchModalOpen = $state(false);
	let taskScheduleModalOpen = $state(false);
	let headerSSOModalOpen = $state(false);
	// Disabled vars for disabling inputs until api request completes
	let signupDisabled = $state(false);
	let debugDisabled = $state(false);
	let jfDisabled = $state(false);
	let jfWebhookEnabledDisabled = $state(false);
	let tmdbkDisabled = $state(false);
	let plexHostDisabled = $state(false);
	let countryDisabled = $state(false);
	let useEmbyDisabled = $state(false);

	let jellyfinWebhookEndpoint: string | undefined = $derived.by(() => {
		if (
			!serverConfig ||
			!serverConfig.JELLYFIN_WEBHOOK_ENABLED ||
			!serverConfig.JELLYFIN_WEBHOOK_KEY
		) {
			return "Error: Please try reloading the page OR re-toggling the webhook setting to get the address!";
		}
		return `${location.protocol}//${location.host}/api/jellyfin/webhook/${serverConfig.JELLYFIN_WEBHOOK_KEY}`;
	});

	async function getServerConfig() {
		serverConfig = await req.get<ServerConfig>(`/server/config`);
	}

	export function updateServerConfig<K extends keyof ServerConfig>(
		name: K,
		value: ServerConfig[K],
		done?: (respData?: object) => void,
	) {
		if (!serverConfig) {
			console.error("updateServerConfig: No server config to update!");
			notify({ type: "error", text: t("server.noConfig") });
			return;
		}
		console.log("Updating server setting", name, "to", value);
		const originalValue = serverConfig[name];
		const nid = notify({ type: "loading", text: t("common.updating") });
		let ep = "/server/config";
		if (name === "PLEX_HOST") {
			ep = "/server/config/plex_host";
		} else if (name === "JELLYFIN_WEBHOOK_ENABLED") {
			ep = "/server/config/JELLYFIN_WEBHOOK_ENABLED";
		}
		req
			.postWhole<object>(ep, { key: name, value: value })
			.then((r) => {
				if (r.status === 200) {
					serverConfig![name] = value;
					notify({ id: nid, type: "success", text: t("common.updated") });
					if (typeof done !== "undefined") done(r?.body);
				}
			})
			.catch((err) => {
				console.error("Failed to update user setting", err);
				notify({ id: nid, type: "error", text: t("common.failedToUpdate") });
				serverConfig![name] = originalValue;
				if (typeof done !== "undefined") done();
			});
	}

	interface ServerStats {
		users: number;
		privateUsers: number;
		watchedMovies: number;
		watchedShows: number;
		watchedSeasons: number;
		mostWatchedMovie: Content;
		mostWatchedShow: Content;
		activities: number;
	}

	async function getServerStats() {
		return await req.get<ServerStats>("/server/stats");
	}
</script>

<svelte:head>
	<title>{t("server.pageTitle")}</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<SettingsList>
			<h2>{t("server.title")}</h2>

			<Stats>
				{#await getServerStats()}
					<Spinner />
				{:then stats}
					<Stat
						name={t("server.stats.users")}
						value={stats.users}
						href={resolve("/manage_users")}
						large
					/>
					<Stat
						name={t("server.stats.privateUsers")}
						value={stats.privateUsers}
						large
					/>
					<Stat
						name={t("server.stats.watchedMovies")}
						value={stats.watchedMovies}
						large
					/>
					<Stat
						name={t("server.stats.watchedShows")}
						value={stats.watchedShows}
						large
					/>
					<Stat
						name={t("server.stats.watchedSeasons")}
						value={stats.watchedSeasons}
						large
					/>
					<Stat
						name={t("server.stats.activities")}
						value={stats.activities}
						large
					/>
					{#if stats.mostWatchedMovie?.title}
						<Stat
							name={t("server.stats.mostWatchedMovie")}
							value={stats.mostWatchedMovie.title}
							href={resolve(`/movie/${stats.mostWatchedMovie.tmdbId}`)}
						/>
					{/if}
					{#if stats.mostWatchedShow?.title}
						<Stat
							name={t("server.stats.mostWatchedShow")}
							value={stats.mostWatchedShow.title}
							href={resolve(`/tv/${stats.mostWatchedShow.tmdbId}`)}
						/>
					{/if}
				{:catch err}
					<Error error={err} pretty={t("server.stats.failedToGet")} />
				{/await}
			</Stats>

			{#await getServerConfig()}
				<Spinner />
			{:then}
				<h3>{t("settings.general")}</h3>
				{#if serverConfig}
					<Setting
						title={t("server.defaultCountry")}
						desc={t("server.defaultCountryDesc")}
					>
						<RegionDropDown
							selectedCountry={serverConfig.DEFAULT_COUNTRY}
							disabled={countryDisabled}
							onChange={(c) => {
								countryDisabled = true;
								updateServerConfig("DEFAULT_COUNTRY", c, () => {
									countryDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title={t("server.host", { service: jellyfinOrEmby })}
						desc={t("server.hostDesc", { service: jellyfinOrEmby })}
					>
						<input
							type="text"
							placeholder="https://{jellyfinOrEmby.toLowerCase()}.example.com"
							bind:value={serverConfig.JELLYFIN_HOST}
							onblur={() => {
								jfDisabled = true;
								updateServerConfig(
									"JELLYFIN_HOST",
									serverConfig!.JELLYFIN_HOST,
									() => {
										jfDisabled = false;
									},
								);
							}}
							disabled={jfDisabled}
						/>
					</Setting>
					<Setting
						title={t("server.jellyfinWebhook")}
						desc={t("server.jellyfinWebhookDesc")}
						row
						handbook="jellyfin-webhook-setup"
					>
						<Checkbox
							name="JELLYFIN_WEBHOOK_ENABLED"
							disabled={jfWebhookEnabledDisabled}
							value={serverConfig.JELLYFIN_WEBHOOK_ENABLED}
							toggled={(on) => {
								jfWebhookEnabledDisabled = true;
								updateServerConfig("JELLYFIN_WEBHOOK_ENABLED", on, (rData) => {
									jfWebhookEnabledDisabled = false;
									if (
										serverConfig &&
										rData &&
										"JELLYFIN_WEBHOOK_KEY" in rData &&
										typeof rData.JELLYFIN_WEBHOOK_KEY === "string"
									) {
										serverConfig.JELLYFIN_WEBHOOK_KEY =
											rData.JELLYFIN_WEBHOOK_KEY;
									}
								});
							}}
						/>
					</Setting>
					{#if serverConfig.JELLYFIN_WEBHOOK_ENABLED}
						<Setting
							title={t("server.jellyfinWebhookEndpoint")}
							desc={t("server.jellyfinWebhookEndpointDesc")}
						>
							<code>
								{jellyfinWebhookEndpoint}
							</code>
						</Setting>
					{/if}
					<Setting
						title={t("server.useEmby")}
						desc={t("server.useEmbyDesc")}
						row
					>
						<Checkbox
							name="USE_EMBY"
							disabled={useEmbyDisabled}
							value={serverConfig.USE_EMBY}
							toggled={(on) => {
								useEmbyDisabled = true;
								updateServerConfig("USE_EMBY", on, () => {
									useEmbyDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title={t("server.host", { service: "Plex" })}
						desc={t("server.hostDesc", { service: "Plex" })}
					>
						<input
							type="text"
							placeholder="https://plex.example.com"
							bind:value={serverConfig.PLEX_HOST}
							onblur={() => {
								plexHostDisabled = true;
								updateServerConfig(
									"PLEX_HOST",
									serverConfig!.PLEX_HOST,
									(rData) => {
										plexHostDisabled = false;
										serverConfig!.PLEX_MACHINE_ID =
											rData &&
											"PLEX_MACHINE_ID" in rData &&
											rData?.PLEX_MACHINE_ID
												? String(rData?.PLEX_MACHINE_ID)
												: undefined;
									},
								);
							}}
							disabled={plexHostDisabled}
						/>
						{#if serverConfig.PLEX_MACHINE_ID}
							<span style="font-size: 10px"
								>{t("server.machineId", {
									id: serverConfig.PLEX_MACHINE_ID,
								})}</span
							>
						{/if}
					</Setting>
					<Setting
						title={t("server.tmdbKey")}
						desc={t("server.tmdbKeyDesc")}
						handbook="your-own-tmdb-api-key"
					>
						<input
							type="password"
							placeholder={t("server.tmdbKey")}
							bind:value={serverConfig.TMDB_KEY}
							onblur={() => {
								tmdbkDisabled = true;
								updateServerConfig("TMDB_KEY", serverConfig!.TMDB_KEY, () => {
									tmdbkDisabled = false;
								});
							}}
							disabled={tmdbkDisabled}
						/>
					</Setting>
					<Setting title={t("server.signup")} desc={t("server.signupDesc")} row>
						<Checkbox
							name="SIGNUP_ENABLED"
							disabled={signupDisabled}
							value={serverConfig.SIGNUP_ENABLED}
							toggled={(on) => {
								signupDisabled = true;
								updateServerConfig("SIGNUP_ENABLED", on, () => {
									signupDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting title={t("server.debug")} desc={t("server.debugDesc")} row>
						<Checkbox
							name="DEBUG"
							disabled={debugDisabled}
							value={serverConfig.DEBUG}
							toggled={(on) => {
								debugDisabled = true;
								updateServerConfig("DEBUG", on, () => {
									debugDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting>
						<SettingButton
							title={t("server.taskSchedule")}
							desc={t("server.taskScheduleDesc")}
							icon="arrow"
							onClick={() => {
								taskScheduleModalOpen = true;
							}}
						/>
					</Setting>
					{#if taskScheduleModalOpen}
						<TaskScheduleModal onClose={() => (taskScheduleModalOpen = false)}
						></TaskScheduleModal>
					{/if}
					<Setting>
						<SettingButton
							title={t("server.headerAuth")}
							desc={t("server.headerAuthDesc")}
							icon="arrow"
							onClick={() => {
								headerSSOModalOpen = true;
							}}
						/>
					</Setting>
					{#if headerSSOModalOpen}
						<TrustedHeaderAuthModal onClose={() => (headerSSOModalOpen = false)}
						></TrustedHeaderAuthModal>
					{/if}
					<div>
						<h3>{t("settings.services")}</h3>
						<h5 class="norm">
							<Trans key="server.servicesDesc">
								{#snippet link(text: string)}
									<a
										style="text-decoration: underline;"
										href="https://github.com/sbondCo/Watcharr/issues/new/choose"
										target="_blank"
									>
										{text}
									</a>
								{/snippet}
							</Trans>
						</h5>
					</div>

					<Setting title="Twitch">
						<SettingButton
							title="Twitch"
							desc={t("server.twitchDesc")}
							icon={serverConfig.TWITCH &&
							Object.keys(serverConfig.TWITCH).length > 0
								? "arrow"
								: "add"}
							onClick={() => {
								twitchModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Sonarr">
						{#if serverConfig.SONARR && serverConfig.SONARR?.length > 0}
							{#each serverConfig.SONARR as server (server.name)}
								<SettingButton
									title={server.name}
									desc={t("server.configureServerAt", { host: server.host })}
									onClick={() => {
										sonarrServerEditing = server;
										sonarrModalEditing = true;
										sonarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Sonarr"
							desc={t("server.addServer", { service: "Sonarr" })}
							icon="add"
							onClick={() => {
								let name = "Sonarr";
								if (serverConfig?.SONARR && serverConfig.SONARR.length > 0) {
									// if this still exists ya on yur own
									name = `Sonarr${serverConfig!.SONARR.length + 1}`;
								}
								sonarrServerEditing = { name };
								sonarrModalEditing = false;
								sonarrModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Radarr">
						{#if serverConfig.RADARR && serverConfig.RADARR?.length > 0}
							{#each serverConfig.RADARR as server (server.name)}
								<SettingButton
									title={server.name}
									desc={t("server.configureServerAt", { host: server.host })}
									onClick={() => {
										radarrServerEditing = server;
										radarrModalEditing = true;
										radarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Radarr"
							desc={t("server.addServer", { service: "Radarr" })}
							icon="add"
							onClick={() => {
								let name = "Radarr";
								if (serverConfig?.RADARR && serverConfig.RADARR.length > 0) {
									// if this still exists ya on yur own
									name = `Radarr${serverConfig!.RADARR.length + 1}`;
								}
								radarrServerEditing = { name };
								radarrModalEditing = false;
								radarrModalOpen = true;
							}}
						/>
					</Setting>

					{#if twitchModalOpen}
						<TwitchModal
							cfg={serverConfig.TWITCH}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								twitchModalOpen = false;
							}}
						/>
					{/if}

					{#if sonarrModalOpen && sonarrServerEditing}
						<SonarrModal
							servarr={sonarrServerEditing}
							isEditing={sonarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								sonarrModalOpen = false;
								sonarrServerEditing = undefined;
							}}
						/>
					{/if}

					{#if radarrModalOpen && radarrServerEditing}
						<RadarrModal
							servarr={radarrServerEditing}
							isEditing={radarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								radarrModalOpen = false;
								radarrServerEditing = undefined;
							}}
						/>
					{/if}
				{/if}
			{:catch err}
				<Error error={err} pretty={t("server.failedToLoadConfig")} />
			{/await}
		</SettingsList>
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
</style>

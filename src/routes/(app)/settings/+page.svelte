<script lang="ts">
	import { goto } from "$app/navigation";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import { updateUserSetting } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import { UserType } from "@/types";
	import PwChangeModal from "./PwChangeModal.svelte";
	import SyncModal from "./SyncModal.svelte";
	import ExportListModal from "./ExportListModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import RatingSetting from "@/lib/rating/RatingSetting.svelte";
	import { toggleTheme } from "@/lib/util/theme";
	import { resolve } from "$app/paths";
	import UserServices from "./UserServices.svelte";
	import LanguageDropDown from "@/lib/LanguageDropDown.svelte";
	import { setLanguage, t } from "@/lib/i18n/index.svelte";

	let user = $derived(store.userInfo);
	let settings = $derived(store.userSettings);
	let selectedTheme = $derived(store.appTheme);

	let privateDisabled = $state(false);
	let privateThoughtsDisabled = $state(false);
	let exportModalOpen = $state(false);
	let hideSpoilersDisabled = $state(false);
	let countryDisabled = $state(false);
	let languageDisabled = $state(false);
	let includePreviouslyWatchedDisabled = $state(false);
	let automateShowStatusesDisabled = $state(false);
	let pwChangeModalOpen = $state(false);
	let jellyfinSyncModalOpen = $state(false);
	let plexSyncModalOpen = $state(false);
</script>

<svelte:head>
	<title>{t("settings.pageTitle")}</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<h2 title={user?.username}>
			<span style="font-weight: normal; font-variant: all-small-caps;">
				{t("settings.settingsFor")}
			</span>
			{user?.username}
		</h2>

		<div class="settings">
			<h3>{t("settings.general")}</h3>

			<div class="theme">
				<h4 class="norm">{t("settings.theme.title")}</h4>
				<div class="row">
					<button
						class={`plain${selectedTheme === "system" ? " selected" : ""}`}
						id="system"
						onclick={() => toggleTheme("system")}
					>
						<span>{t("settings.theme.system")}</span>
					</button>
					<button
						class={`plain${selectedTheme === "light" ? " selected" : ""}`}
						id="light"
						onclick={() => toggleTheme("light")}
					>
						{t("settings.theme.light")}
					</button>
					<button
						class={`plain${selectedTheme === "dark" ? " selected" : ""}`}
						id="dark"
						onclick={() => toggleTheme("dark")}
					>
						{t("settings.theme.dark")}
					</button>
				</div>
			</div>

			<Setting
				title={t("settings.language.title")}
				desc={t("settings.language.desc")}
			>
				<LanguageDropDown
					selectedLanguage={settings?.language ?? ""}
					disabled={languageDisabled}
					onChange={(l) => {
						languageDisabled = true;
						updateUserSetting("language", l, async () => {
							// Apply whatever is now in the store (the setting
							// is reverted there if the update failed).
							await setLanguage(store.userSettings?.language);
							languageDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title={t("settings.country.title")}
				desc={t("settings.country.desc")}
			>
				<RegionDropDown
					selectedCountry={settings?.country}
					disabled={countryDisabled}
					onChange={(c) => {
						countryDisabled = true;
						updateUserSetting("country", c, () => {
							countryDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title={t("settings.private.title")}
				desc={t("settings.private.desc")}
				row
			>
				<Checkbox
					name="private"
					disabled={privateDisabled}
					value={settings?.private}
					toggled={(on) => {
						privateDisabled = true;
						updateUserSetting("private", on, () => {
							privateDisabled = false;
						});
					}}
				/>
			</Setting>

			{#if !settings?.private}
				<Setting
					title={t("settings.privateThoughts.title")}
					desc={t("settings.privateThoughts.desc")}
					row
				>
					<Checkbox
						name="privateThoughts"
						disabled={privateThoughtsDisabled}
						value={settings?.privateThoughts}
						toggled={(on) => {
							privateThoughtsDisabled = true;
							updateUserSetting("privateThoughts", on, () => {
								privateThoughtsDisabled = false;
							});
						}}
					/>
				</Setting>
			{/if}

			<Setting
				title={t("settings.hideSpoilers.title")}
				desc={t("settings.hideSpoilers.desc")}
				row
			>
				<Checkbox
					name="hideSpoilers"
					disabled={hideSpoilersDisabled}
					value={settings?.hideSpoilers}
					toggled={(on) => {
						hideSpoilersDisabled = true;
						updateUserSetting("hideSpoilers", on, () => {
							hideSpoilersDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title={t("settings.automateShowStatuses.title")}
				desc={t("settings.automateShowStatuses.desc")}
				tag={t("common.experimental")}
				row
			>
				<Checkbox
					name="automateShowStatusesDisabled"
					disabled={automateShowStatusesDisabled}
					value={settings?.automateShowStatuses}
					toggled={(on) => {
						automateShowStatusesDisabled = true;
						updateUserSetting("automateShowStatuses", on, () => {
							automateShowStatusesDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title={t("settings.includePreviouslyWatched.title")}
				desc={t("settings.includePreviouslyWatched.desc")}
				row
			>
				<Checkbox
					name="includePreviouslyWatched"
					disabled={includePreviouslyWatchedDisabled}
					value={settings?.includePreviouslyWatched}
					toggled={(on) => {
						includePreviouslyWatchedDisabled = true;
						updateUserSetting("includePreviouslyWatched", on, () => {
							includePreviouslyWatchedDisabled = false;
						});
					}}
				/>
			</Setting>

			<RatingSetting />

			<h3>{t("settings.services")}</h3>

			<UserServices />

			<div class="row btns">
				<button onclick={() => goto(resolve("/import"))}
					>{t("settings.import")}</button
				>
				<button onclick={() => (exportModalOpen = true)}
					>{t("settings.export")}</button
				>
				{#if user?.type == UserType.Watcharr}
					<button
						onclick={() => {
							pwChangeModalOpen = true;
						}}>{t("settings.changePassword")}</button
					>
				{/if}
				{#if user?.type === UserType?.Jellyfin}
					<button onclick={() => (jellyfinSyncModalOpen = true)}>
						{t("settings.syncWith", {
							service: localStorage.getItem("useEmby") ? "Emby" : "Jellyfin",
						})}
					</button>
				{/if}
				{#if user?.type === UserType?.Plex}
					<button onclick={() => (plexSyncModalOpen = true)}>
						{t("settings.syncWith", { service: "Plex" })}
					</button>
				{/if}
			</div>
			{#if exportModalOpen}
				<ExportListModal
					onClose={() => {
						exportModalOpen = false;
					}}
				/>
			{/if}
			{#if pwChangeModalOpen}
				<PwChangeModal
					userName={user?.username}
					onClose={() => {
						pwChangeModalOpen = false;
					}}
				></PwChangeModal>
			{/if}
			{#if jellyfinSyncModalOpen}
				<SyncModal onClose={() => (jellyfinSyncModalOpen = false)} />
			{/if}
			{#if plexSyncModalOpen}
				<SyncModal type="plex" onClose={() => (plexSyncModalOpen = false)} />
			{/if}
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

			h3 {
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

	.settings {
		display: flex;
		flex-flow: column;
		gap: 20px;
		width: 100%;
		margin-top: 10px;

		div {
			&.row {
				display: flex;
				flex-flow: row;
				gap: 10px;
				align-items: center;

				&.btns button {
					width: max-content;
				}
			}
		}

		.theme {
			display: flex;
			flex-flow: column;
			gap: 10px;

			& .row {
				margin: 0 5px;
			}

			& button {
				width: 50%;
				height: 80px;
				border-radius: 10px;
				outline: 3px solid;
				font-size: 20px;
				text-transform: uppercase;
				font-family: "Rampart One";
				color: transparent;
				transition: all 200ms ease-in;

				&#light {
					background-color: white;
					outline-color: $accent-color;
					&:hover {
						color: black;
						-webkit-text-stroke: 0.5px black;
					}
				}

				&#dark {
					background-color: black;
					outline-color: white;
					&:hover {
						color: white;
						-webkit-text-stroke: 0.5px white;
					}
				}

				&#system {
					background: linear-gradient(to right bottom, white 50%, black 50.3%);
					outline-color: black;

					span {
						mix-blend-mode: difference;
					}

					&:hover {
						color: white;
						-webkit-text-stroke: 0.5px white;
					}
				}

				&.selected {
					outline-color: gold !important;
				}
			}
		}
	}
</style>

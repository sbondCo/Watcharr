<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { req } from "../util/api";
	import Modal from "../Modal.svelte";
	import type {
		ArrRequestResponse,
		ListBoxItem,
		Media,
		SonarrSettingsPublicResponseResult,
		SonarrTestResponse,
	} from "@/types";
	import { notify } from "../util/notify";
	import DropDown from "../DropDown.svelte";
	import Setting from "../settings/Setting.svelte";
	import Spinner from "../Spinner.svelte";
	import ListBox from "../ListBox.svelte";

	interface Props {
		content: Media;
		onClose: (r: ArrRequestResponse | undefined) => void;
		approveMode?: boolean;
		originalRequest?: ArrRequestResponse | undefined;
	}

	let {
		content,
		onClose,
		approveMode = false,
		originalRequest = undefined,
	}: Props = $props();

	let servarrs: SonarrSettingsPublicResponseResult[] | undefined = $state();
	let selectedServarrIndex: number = $state(0);
	let selectedServerCfg: SonarrTestResponse | undefined = $state();
	let seasonItems: ListBoxItem[] = $state(
		content.seasons
			? content.seasons.flatMap((s) => {
					if (s.number == undefined) {
						return [];
					}
					return {
						id: s.number,
						value: false,
						displayValue: s.name ?? `(${s.number}) ${t("common.unknown")}`,
					};
				})
			: [],
	);
	let addRequestRunning = $state(false);

	async function getServers() {
		try {
			const r = await req.get<SonarrSettingsPublicResponseResult[]>("/arr/son");
			if (r.length > 0) {
				servarrs = r;
				selectedServarrIndex = 0;
			} else {
				notify({ text: t("request.noServersFound"), type: "error" });
			}
			processOriginalRequest();
		} catch (err) {
			console.error("Failed to get servers!", err);
			notify({ text: t("request.failedToLoadServers"), type: "error" });
		}
	}

	async function getConfig(name: string) {
		try {
			const r = await req.get<SonarrTestResponse>(`/arr/son/config/${name}`);
			selectedServerCfg = r;
		} catch (err) {
			console.error("Failed to get config!", err);
			notify({ text: t("request.failedToLoadConfig"), type: "error" });
		}
	}

	async function request() {
		let nid;
		try {
			if (!servarrs || !servarrs[selectedServarrIndex]) {
				notify({ text: t("request.mustSelectServer"), type: "error" });
				return;
			}
			if (!selectedServerCfg) {
				notify({ text: t("request.noServerConfig"), type: "error" });
				return;
			}
			addRequestRunning = true;
			nid = notify({ text: t("request.requesting"), type: "loading" });
			const server = servarrs[selectedServarrIndex];
			const rootFolder = selectedServerCfg.rootFolders?.find(
				(f) => f.id === server.rootFolder,
			);
			if (!rootFolder) {
				console.error(
					"show request.. no root folder found with id:",
					server.rootFolder,
					"rf:",
					rootFolder,
				);
				notify({ id: nid, text: t("request.noRootFolder"), type: "error" });
				return;
			}
			const resp = await req.post<ArrRequestResponse>(
				`/arr/son/request${approveMode && originalRequest ? `/approve/${originalRequest.id}` : ""}`,
				{
					serverName: server.name,
					title: content.name,
					year: content.releaseDate
						? new Date(content.releaseDate)?.getFullYear()
						: undefined,
					tmdbId: content.ids.tmdb,
					tvdbId: content.ids.tvdb,
					seriesType: content.isShowAnime ? "anime" : "standard",
					qualityProfile: server.qualityProfile,
					rootFolder: rootFolder.path,
					languageProfile: server.languageProfile,
					seasons: seasonItems.map((s) => {
						return {
							seasonNumber: s.id,
							monitored: s.value,
						};
					}),
				},
			);
			addRequestRunning = false;
			if (resp) {
				notify({ id: nid, text: t("request.complete"), type: "success" });
				onClose(resp);
			}
		} catch (err) {
			console.error("content request failed!", err);
			addRequestRunning = false;
			notify({ id: nid, text: t("request.failed"), type: "error" });
		}
	}

	function processOriginalRequest() {
		if (!originalRequest) {
			return;
		}
		try {
			if (originalRequest.requestJson) {
				const ogr = JSON.parse(originalRequest.requestJson);
				if (!ogr) {
					console.info("processOriginalRequest: No json.", ogr);
					return;
				}
				if (ogr?.seasons?.length > 0) {
					console.debug("processOriginalRequest: Found seasons.. restoring.");
					for (let i = 0; i < ogr.seasons.length; i++) {
						const s = ogr.seasons[i];
						// Default is not monitored, so no point going through the whole rigmarole to 'restore' the default value.
						if (!s.monitored) {
							continue;
						}
						const sItem = seasonItems?.find((si) => si.id === s.seasonNumber);
						if (sItem) {
							sItem.value = s.monitored;
						}
					}
				}
				if (ogr?.serverName) {
					console.debug(
						"processOriginalRequest: restoring server name:",
						ogr?.serverName,
					);
					const idx = servarrs?.findIndex((s) => s.name === ogr?.serverName);
					if (idx !== undefined && idx !== -1) {
						selectedServarrIndex = idx;
					}
				}
			} else {
				notify({
					type: "error",
					text: t("request.couldNotRestore"),
					time: 10000,
				});
			}
		} catch (err) {
			console.error("processOriginalRequest: Failed!", err);
			notify({
				text: t("request.failedProcessingOriginal"),
				type: "error",
			});
		}
	}

	$effect.pre(() => {
		if (
			typeof selectedServarrIndex !== "undefined" &&
			servarrs &&
			servarrs?.length > 0
		) {
			const s = servarrs[selectedServarrIndex];
			if (!s) {
				selectedServerCfg = undefined;
			} else {
				getConfig(s.name);
			}
		}
	});

	getServers();
</script>

<Modal
	title={approveMode ? t("request.approveRequest") : t("request.request")}
	desc={content.name}
	maxWidth="700px"
	onClose={() => onClose(undefined)}
>
	<div class="req-ctr">
		{#if servarrs}
			<div class="seasons-list">
				<ListBox
					bind:options={seasonItems}
					allCheckBox={t("request.allSeasons")}
				/>
			</div>

			{#if servarrs?.length > 1}
				<Setting title={t("request.selectServerTitle")}>
					<DropDown
						placeholder={t("request.selectServer")}
						bind:active={selectedServarrIndex}
						options={servarrs?.length > 0
							? servarrs.map((s, i) => {
									return { id: i, value: s.name };
								})
							: []}
					/>
				</Setting>
			{/if}

			<button onclick={request} disabled={addRequestRunning}>
				{approveMode ? t("request.approve") : t("request.request")}
			</button>
		{:else}
			<Spinner />
		{/if}
	</div>
</Modal>

<style lang="scss">
	.req-ctr {
		display: flex;
		flex-flow: column;
		gap: 10px;
		height: 100%;

		.seasons-list {
			max-height: 500px;
			overflow: auto;
		}

		button {
			margin-top: auto;
			margin-left: auto;
			width: max-content;
		}
	}
</style>

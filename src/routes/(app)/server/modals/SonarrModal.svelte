<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { run } from "svelte/legacy";

	import Checkbox from "@/lib/Checkbox.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { notify } from "@/lib/util/notify";
	import type {
		DropDownItem,
		SonarrSettings,
		SonarrTestResponse,
	} from "@/types";
	import { req } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";

	interface Props {
		servarr: SonarrSettings;
		isEditing: boolean;
		onClose: () => void;
	}

	let { servarr = $bindable(), isEditing, onClose }: Props = $props();

	let error: string | undefined = $state();
	let formDisabled = $state(false);

	let qualityProfiles: DropDownItem[] = $state([]);
	let languageProfiles: DropDownItem[] = $state([]);
	let rootFolders: DropDownItem[] = $state([]);

	function testIfHostAndKeySet() {
		if (servarr.host && servarr.key) {
			console.log("host and key set.. loading settings data from sonarr");
			getSettingsData();
		} else {
			qualityProfiles = [];
			languageProfiles = [];
			rootFolders = [];
			error = t("server.arr.hostAndKeyRequired");
		}
	}

	async function getSettingsData() {
		try {
			formDisabled = true;
			const res = await req.post<SonarrTestResponse>("/arr/son/test", {
				host: servarr.host,
				key: servarr.key,
			});
			qualityProfiles = res.qualityProfiles.map((d) => {
				return { id: d.id, value: d.name };
			});
			languageProfiles = res.languageProfiles.map((d) => {
				return { id: d.id, value: d.name };
			});
			rootFolders = res.rootFolders.map((d) => {
				return { id: d.id, value: d.path };
			});
			formDisabled = false;
			error = "";
		} catch (err) {
			console.error("getSettingsData failed!", err);
			error = t("server.arr.checkHostAndKey");
			formDisabled = false;
		}
	}

	function checkForm() {
		let errs: string[] = [];
		if (!servarr.name) {
			errs.push("name");
		}
		if (!servarr.host) {
			errs.push("host");
		}
		if (!servarr.key) {
			errs.push("key");
		}
		if (!servarr.qualityProfile) {
			errs.push("qualityProfile");
		}
		if (!servarr.rootFolder) {
			errs.push("rootFolder");
		}
		if (!servarr.languageProfile) {
			errs.push("languageProfile");
		}
		if (errs.length > 0) {
			error = t("password.missingParams", { params: errs.join(", ") });
		} else {
			error = "";
		}
	}

	async function save() {
		checkForm();
		// if no error set from checkForm func, continue
		if (!error) {
			console.log(servarr);
			try {
				await req.post(`/arr/son/${isEditing ? "edit" : "add"}`, servarr);
				notify({
					type: "success",
					text: isEditing
						? t("manageUsers.edit.changesSaved")
						: t("server.arr.added"),
				});
				onClose();
			} catch (err) {
				console.error("Failed to save server!", err);
				error = ReqerError.getMsg(
					err,
					isEditing
						? t("server.arr.failedToEdit")
						: t("server.arr.failedToAdd"),
				);
			}
		}
	}

	async function remove() {
		try {
			await req.post(`/arr/son/rm/${servarr.name}`);
			notify({
				type: "success",
				text: t("server.arr.removed"),
			});
			onClose();
		} catch (err) {
			console.error("Failed to remove server!", err);
			error = ReqerError.getMsg(err, t("server.arr.failedToRemove"));
		}
	}
	run(() => {
		if (isEditing) {
			getSettingsData();
		}
	});
</script>

<Modal
	title={isEditing
		? t("server.arr.edit", { name: servarr.name })
		: t("server.arr.add", { service: "Sonarr" })}
	desc={t("server.arr.setupDesc", { service: "Sonarr" })}
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}
	<SettingsList>
		{#if !isEditing}
			<Setting title={t("tags.name")} desc={t("server.arr.nameDesc")}>
				<input
					type="text"
					placeholder="https://sonarr.example.com"
					bind:value={servarr.name}
					disabled={formDisabled}
				/>
			</Setting>
		{/if}
		<Setting
			title={t("server.host", { service: "Sonarr" })}
			desc={t("server.arr.hostDescSonarr")}
		>
			<input
				type="text"
				placeholder="https://sonarr.example.com"
				bind:value={servarr.host}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title={t("server.arr.key", { service: "Sonarr" })}
			desc={t("server.arr.keyDesc", { service: "Sonarr" })}
		>
			<input
				type="text"
				placeholder="dGhhbmtzIGZvciB1c2luZyB3YXRjaGFyciA6KQ=="
				bind:value={servarr.key}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title={t("server.arr.qualityProfile")}
			desc={t("server.arr.qualityProfileDescSonarr")}
		>
			<DropDown
				placeholder={qualityProfiles?.length == 0
					? t("server.arr.addHostAndKeyFirst")
					: t("server.arr.selectQualityProfile")}
				options={qualityProfiles}
				bind:active={servarr.qualityProfile}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title={t("server.arr.rootFolder")}
			desc={t("server.arr.rootFolderDescSonarr")}
		>
			<DropDown
				placeholder={rootFolders?.length == 0
					? t("server.arr.addHostAndKeyFirst")
					: t("server.arr.selectRootFolder")}
				options={rootFolders}
				bind:active={servarr.rootFolder}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title={t("server.arr.languageProfile")}
			desc={t("server.arr.languageProfileDesc")}
		>
			<DropDown
				placeholder={languageProfiles?.length == 0
					? t("server.arr.addHostAndKeyFirst")
					: t("server.arr.selectLanguageProfile")}
				options={languageProfiles}
				bind:active={servarr.languageProfile}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title={t("server.arr.automaticSearch")}
			desc={t("server.arr.automaticSearchDescSonarr")}
			row
		>
			<Checkbox
				name="automaticSearch"
				disabled={formDisabled}
				bind:value={servarr.automaticSearch}
			/>
		</Setting>
		<div class="btns">
			{#if isEditing}
				<button class="danger" onclick={() => remove()}
					>{t("common.delete")}</button
				>
			{/if}
			<button id="test" class="secondary" onclick={() => testIfHostAndKeySet()}
				>{t("server.arr.test")}</button
			>
			<button onclick={() => save()}
				>{isEditing ? t("common.save") : t("server.arr.addServer")}</button
			>
		</div>
	</SettingsList>
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		#test {
			margin-left: auto;
		}

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;
		}
	}

	.error {
		position: sticky;
		top: 0;
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
		margin-bottom: 15px;
	}
</style>

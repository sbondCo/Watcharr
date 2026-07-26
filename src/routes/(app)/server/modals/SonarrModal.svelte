<script lang="ts">
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
			error = "host and key must be provided to test the connection";
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
			error = `Request Failed, check your Host and Key`;
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
			error = `Missing required params: ${errs.join(", ")}`;
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
					text: isEditing ? "Changes saved!" : "Server added successfully!",
				});
				onClose();
			} catch (err: any) {
				console.error("Failed to save server!", err);
				error = ReqerError.getMsg(
					err,
					`Failed to ${isEditing ? "edit" : "add"}`,
				);
			}
		}
	}

	async function remove() {
		try {
			await req.post(`/arr/son/rm/${servarr.name}`);
			notify({
				type: "success",
				text: "Removed server",
			});
			onClose();
		} catch (err: any) {
			console.error("Failed to remove server!", err);
			error = ReqerError.getMsg(err, "Failed to remove");
		}
	}
	run(() => {
		if (isEditing) {
			getSettingsData();
		}
	});
</script>

<Modal
	title={isEditing ? `Edit ${servarr.name}` : "Add Sonarr Server"}
	desc="Setup a connection to your Sonarr server"
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}
	<SettingsList>
		{#if !isEditing}
			<Setting title="Name" desc="Give your server a memorable name.">
				<input
					type="text"
					placeholder="https://sonarr.example.com"
					bind:value={servarr.name}
					disabled={formDisabled}
				/>
			</Setting>
		{/if}
		<Setting
			title="Sonarr Host"
			desc="Point to your Sonarr server to enable tv show requesting."
		>
			<input
				type="text"
				placeholder="https://sonarr.example.com"
				bind:value={servarr.host}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting title="Sonarr Key" desc="API key for your Sonarr instance.">
			<input
				type="text"
				placeholder="dGhhbmtzIGZvciB1c2luZyB3YXRjaGFyciA6KQ=="
				bind:value={servarr.key}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title="Quality Profile"
			desc="Default quality profile when adding shows."
		>
			<DropDown
				placeholder={qualityProfiles?.length == 0
					? "Add a Host and Key, then press test to view"
					: "Select a quality profile"}
				options={qualityProfiles}
				bind:active={servarr.qualityProfile}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting title="Root Folder" desc="Default root folder when adding shows.">
			<DropDown
				placeholder={rootFolders?.length == 0
					? "Add a Host and Key, then press test to view"
					: "Select a root folder"}
				options={rootFolders}
				bind:active={servarr.rootFolder}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title="Language Profile"
			desc="Default language profile when adding shows."
		>
			<DropDown
				placeholder={languageProfiles?.length == 0
					? "Add a Host and Key, then press test to view"
					: "Select a language profile"}
				options={languageProfiles}
				bind:active={servarr.languageProfile}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting
			title="Automatic Search"
			desc="Start missing episode search automatically?"
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
				<button class="danger" onclick={() => remove()}>Delete</button>
			{/if}
			<button id="test" class="secondary" onclick={() => testIfHostAndKeySet()}
				>Test</button
			>
			<button onclick={() => save()}>{isEditing ? "Save" : "Add Server"}</button
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

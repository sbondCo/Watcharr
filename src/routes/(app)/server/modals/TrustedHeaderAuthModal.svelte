<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Notice from "@/lib/Notice.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { ServerConfigByName, TrustedHeaderAuthSetting } from "@/types";
	import { onMount } from "svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let headerCfg: TrustedHeaderAuthSetting = $state({
		enabled: false,
		headerName: "",
		logoutUrl: "",
		autoLogin: false,
	});
	let formDisabled = false;
	let loadingCfg = $state(false);
	let error = "";

	async function getHeaderCfg() {
		try {
			loadingCfg = true;
			const res = await req.get<ServerConfigByName<TrustedHeaderAuthSetting>>(
				"/server/config",
				{
					params: { s: "HEADER_AUTH" },
				},
			);
			headerCfg = res.value;
			loadingCfg = false;
		} catch (err) {
			console.error("getHeaderCfg failed!", err);
			notify({
				type: "error",
				text: "Failed to get configuration from server.",
				time: 6000,
			});
		}
	}

	async function save() {
		const nid = notify({
			type: "loading",
			text: "Saving..",
		});
		try {
			await req.post("/server/config", headerCfg, {
				params: { s: "HEADER_AUTH" },
			});
			notify({
				id: nid,
				type: "success",
				text: "Changes saved!",
			});
			onClose();
		} catch (err) {
			console.error("save failed!", err);
			notify({
				id: nid,
				type: "error",
				text: "Failed to save config. Please try again!",
				time: 6000,
			});
		}
	}

	onMount(() => {
		getHeaderCfg();
	});
</script>

<Modal
	title="Trusted Header Authentication"
	desc="Configure trusted header single sign-on."
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}
	{#if loadingCfg}
		<Spinner />
	{:else}
		<SettingsList>
			<Notice
				title="This is dangerous!"
				desc="If setup incorrectly, the authorization module for your server could be easily comprimised. If configured, please ensure your Watcharr instance is only available through your proxy and not available directly."
				type="warn"
			/>
			<Setting
				title="Header Name"
				desc="Name of the header used for authentication."
			>
				<input
					type="text"
					placeholder="X-User"
					onblur={() => {}}
					disabled={formDisabled}
					bind:value={headerCfg.headerName}
				/>
			</Setting>
			<Setting
				title="Logout URL"
				desc="Where can we redirect so that the user can logout?"
			>
				<input
					type="text"
					placeholder="https://auth.example.com/logout"
					onblur={() => {}}
					disabled={formDisabled}
					bind:value={headerCfg.logoutUrl}
				/>
			</Setting>
			<Setting
				title="Auto Login"
				desc="Should we auto login users as soon as they reach our login page?"
				row
			>
				<Checkbox
					name="HeaderAuthAutoLogin"
					disabled={formDisabled}
					bind:value={headerCfg.autoLogin}
				/>
			</Setting>
			<Setting
				title="Enabled"
				desc="Enable Trusted Header Single Sign-On? Toggle on to activate above configuration."
				row
			>
				<Checkbox
					name="HeaderAuthEnabled"
					disabled={formDisabled}
					bind:value={headerCfg.enabled}
				/>
			</Setting>
			<div class="btns">
				<button onclick={() => save()}>Save</button>
			</div>
		</SettingsList>
	{/if}
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;

			&:nth-child(1) {
				margin-left: auto;
			}
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

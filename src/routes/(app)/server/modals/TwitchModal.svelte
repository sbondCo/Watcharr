<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import { notify } from "@/lib/util/notify";
	import type { TwitchSettings } from "@/types";

	interface Props {
		cfg?: TwitchSettings;
		onClose: () => void;
	}

	let { cfg = {}, onClose }: Props = $props();

	let error: string | undefined = $state();
	let formDisabled = false;

	function checkForm() {
		let errs: string[] = [];
		if (!cfg.clientId) {
			errs.push("Client ID");
		}
		if (!cfg.clientSecret) {
			errs.push("Client Secret");
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
			try {
				await req.post(`/game/config`, cfg);
				notify({
					type: "success",
					text: "Changes saved!",
				});
				onClose();
			} catch (err: any) {
				console.error("Failed to save twitch cfg!", err);
				error = ReqerError.getMsg(err, "Failed to save");
			}
		}
	}
</script>

<Modal
	title={"Twitch Config"}
	desc="Setup your twitch application to enable game support."
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}

	<SettingsList>
		<a
			href="https://watcharr.app/docs/server_config/game-support-igdb"
			target="_blank"
		>
			Learn how to configure this option at Watcharr Docs.
		</a>

		<Setting title="Client ID" desc="Twitch application Client ID.">
			<input
				type="text"
				placeholder="yrbrrakvgbce99fzjsidkfoalsk9eee"
				value={cfg.clientId}
				oninput={(e) => (cfg.clientId = e.currentTarget.value)}
				disabled={formDisabled}
			/>
		</Setting>
		<Setting title="Client Secret" desc="Twitch application Client Secret.">
			<input
				type="text"
				placeholder="yrbralsodkfishfzpajdkflqoefeee"
				value={cfg.clientSecret}
				oninput={(e) => (cfg.clientSecret = e.currentTarget.value)}
				disabled={formDisabled}
			/>
		</Setting>
		<div class="btns">
			<button onclick={() => save()}>Save</button>
		</div>
	</SettingsList>
</Modal>

<style lang="scss">
	a {
		text-decoration: underline;
		transition: opacity 100ms ease;

		&:hover {
			opacity: 0.8;
		}
	}

	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		:first-child {
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

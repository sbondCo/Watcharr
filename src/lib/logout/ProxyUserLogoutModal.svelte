<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { onMount } from "svelte";
	import Modal from "../Modal.svelte";
	import Trans from "../i18n/Trans.svelte";
	import type { TrustedHeaderAuthLogoutDetailsResponse } from "@/types";
	import Spinner from "../Spinner.svelte";
	import { clearWatcharrData } from ".";
	import { goto } from "$app/navigation";
	import { req } from "../util/api";
	import { resolve } from "$app/paths";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let loadingBtns = $state(true);
	let logoutUrl: string | undefined = $state(undefined);

	onMount(async () => {
		try {
			const r = await req.get<TrustedHeaderAuthLogoutDetailsResponse>(
				"/auth/proxy_logout_details",
			);
			logoutUrl = r?.logoutUrl;
			if (!logoutUrl?.toLowerCase()?.startsWith("http")) {
				// If no protocol in logoutUrl, set https
				logoutUrl = `https://${logoutUrl}`;
			}
		} catch (err) {
			console.error("Failed to get proxy logout details!", err);
		}
		loadingBtns = false;
	});

	function logout() {
		clearWatcharrData();
		goto(resolve("/login?noAuto=1"));
	}

	function proxyLogout() {
		if (!logoutUrl) {
			console.error(
				"proxyLogout: Not supported without a configured logoutUrl.",
			);
			return;
		}
		clearWatcharrData();
		window.location.replace(logoutUrl);
	}
</script>

<Modal title={t("logout.title")} desc={t("logout.proxyDesc")} {onClose}>
	{#if loadingBtns}
		<Spinner />
	{:else}
		<div>
			<p>
				{#if logoutUrl}
					<Trans key="logout.likelyWhatYouWant" />
				{:else}
					{t("logout.ssoLogoutNotConfigured")}
				{/if}
			</p>
			<button disabled={!logoutUrl} onclick={proxyLogout}
				>{t("logout.logOutOfSso")}</button
			>
			<p>
				<Trans key="logout.watcharrOnlyWarning" />
			</p>
			<button onclick={logout}>{t("logout.logOutOfWatcharr")}</button>
		</div>
	{/if}
</Modal>

<style lang="scss">
	div {
		display: flex;
		flex-flow: column;

		p {
			margin-bottom: 5px;
		}

		button:first-of-type {
			margin-bottom: 15px;
		}
	}
</style>

<script lang="ts">
	import { onMount } from "svelte";
	import Modal from "../Modal.svelte";
	import type { TrustedHeaderAuthLogoutDetailsResponse } from "@/types";
	import Spinner from "../Spinner.svelte";
	import { clearWatcharrData } from ".";
	import { goto } from "$app/navigation";
	import { req } from "../util/api";

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
		goto("/login?noAuto=1");
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

<Modal
	title="Logout"
	desc="You logged in via single sign-on. Did you click log out by mistake?"
	{onClose}
>
	{#if loadingBtns}
		<Spinner />
	{:else}
		<div>
			<p>
				{#if logoutUrl}
					This is likely what you want: <b
						>Fully logout of your single sign-on service and Watcharr</b
					>.
				{:else}
					Single sign-on logout has not been configured on this server. If you
					are not the server operator, let them know!
				{/if}
			</p>
			<button disabled={!logoutUrl} onclick={proxyLogout}
				>Log out of Single Sign-On Service</button
			>
			<p>
				Logging out of Watcharr will clear your local credentials and data, but <b
					>you will still be logged in to your single sign-on service</b
				>! Do <b>not</b> do this on a public machine and assume you are logged out,
				this account could still be accessible.
			</p>
			<button onclick={logout}>Log out of Watcharr</button>
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

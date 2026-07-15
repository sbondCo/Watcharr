<script lang="ts">
	import Notifications from "@/lib/notifications.svelte";
	import { onMount } from "svelte";
	import { pwaInfo } from "virtual:pwa-info";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	function resetTooltipPos() {
		const t = document.getElementById("tooltip");
		if (t) {
			t.style.top = "0";
			t.style.left = "0";
		}
	}

	onMount(() => {
		window.addEventListener("resize", resetTooltipPos);

		return () => {
			window.removeEventListener("resize", resetTooltipPos);
		};
	});
</script>

<svelte:head>
	{#if pwaInfo?.webManifest?.linkTag}
		<!-- eslint-disable-next-line -->
		{@html pwaInfo.webManifest.linkTag}
	{/if}
</svelte:head>

<div id="tooltip"></div>

<Notifications />

{@render children?.()}

<style lang="scss">
	@use "../styles/norm.scss";
</style>

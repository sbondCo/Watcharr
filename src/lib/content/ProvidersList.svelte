<script lang="ts">
	import type { MediaProvider } from "@/types";
	import ProviderIcon from "./ProviderIcon.svelte";

	interface Props {
		providers: MediaProvider[];
		fullListLink?: string;
		fullListLinkText?: string;
	}

	let { providers, fullListLink, fullListLinkText }: Props = $props();
</script>

{#if providers?.length > 0}
	<div class="streaming-providers">
		{#each providers as provider (provider.name)}
			<ProviderIcon i={provider.name} href={provider.link} wh={40} />
		{/each}
		{#if fullListLink}
			<!-- The fullListLink is important for TMDB data, we always show it
		 as "JustWatch" (set in component prop) because that data requires
		 attribution! but also it helps support tmdb in some way. -->
			<a href={fullListLink} rel="external" target="_blank">
				{fullListLinkText}
			</a>
		{/if}
	</div>
{/if}

<style lang="scss">
	.streaming-providers {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 7px 15px;
		margin-top: auto;
		padding-top: 10px;

		a {
			color: white;
			opacity: 0.7;

			&:hover {
				opacity: 1;
			}
		}
	}
</style>

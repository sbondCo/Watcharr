<script lang="ts">
	import { MediaVideoType, type MediaVideo } from "@/types";
	import VideoEmbedModal from "./VideoEmbedModal.svelte";

	interface Props {
		videos?: MediaVideo[];
	}

	let { videos }: Props = $props();

	let trailer: string | undefined = $derived.by(() => {
		if (videos && videos?.length > 0) {
			const t = videos
				.filter((v) => v.type === MediaVideoType.trailer)
				.sort((a, b) => {
					if (a.best === b.best) return 0;
					return a.best ? -1 : 1;
				});
			if (t[0]?.id) {
				return `https://www.youtube-nocookie.com/embed/${t[0]?.id}`;
			}
		}
	});
	let trailerShown = $state(false);
</script>

{#if trailer}
	<button onclick={() => (trailerShown = !trailerShown)}>View Trailer</button>
	{#if trailerShown}
		<VideoEmbedModal embed={trailer} closed={() => (trailerShown = false)} />
	{/if}
{/if}

<style lang="scss">
	button {
		max-width: fit-content;
	}
</style>

<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { MediaVideoType, type MediaVideo } from "@/types";
	import VideoEmbedModal from "./VideoEmbedModal.svelte";

	interface Props {
		videos?: MediaVideo[];
	}

	let { videos }: Props = $props();

	let trailer: string | undefined = $derived.by(() => {
		if (videos && videos?.length > 0) {
			const trailers = videos
				.filter((v) => v.type === MediaVideoType.trailer)
				.sort((a, b) => {
					if (a.best === b.best) return 0;
					return a.best ? -1 : 1;
				});
			if (trailers[0]?.id) {
				return `https://www.youtube-nocookie.com/embed/${trailers[0]?.id}`;
			}
		}
	});
	let trailerShown = $state(false);
</script>

{#if trailer}
	<button onclick={() => (trailerShown = !trailerShown)}
		>{t("content.viewTrailer")}</button
	>
	{#if trailerShown}
		<VideoEmbedModal embed={trailer} closed={() => (trailerShown = false)} />
	{/if}
{/if}

<style lang="scss">
	button {
		max-width: fit-content;
	}
</style>

<script lang="ts">
	import type { Tag, Watched } from "@/types";
	import tooltip from "../actions/tooltip";
	import Icon from "../Icon.svelte";
	import TagMenu from "./TagMenu.svelte";
	import { tagWatched, untagWatched } from "./api";
	import { onMount } from "svelte";

	interface Props {
		watchedItem: Watched;
	}

	let { watchedItem }: Props = $props();

	let menuOpen = $state(false);
	let btnEl: HTMLButtonElement | undefined = $state();

	onMount(() => {
		const onScroll = () => {
			menuOpen = false;
		};

		window.addEventListener("scroll", onScroll);

		return () => {
			window.removeEventListener("scroll", onScroll);
		};
	});

	function onTagClick(tag: Tag, remove: boolean) {
		console.debug("Tag: Adding content to tag. Remove?:", remove);
		if (remove) {
			untagWatched(watchedItem.id, tag).then((s) => {
				if (!s) return;
				if (watchedItem.tags) {
					watchedItem.tags = watchedItem.tags.filter((t) => t.id !== tag.id);
				}
			});
			return;
		}
		tagWatched(watchedItem.id, tag).then((s) => {
			if (!s) return;
			if (!watchedItem.tags) {
				watchedItem.tags = [tag];
			} else {
				watchedItem.tags.push(tag);
			}
		});
	}
</script>

<div>
	<button
		onclick={() => (menuOpen = !menuOpen)}
		use:tooltip={{
			text: `Add to a Tag`,
			pos: "bot",
		}}
		bind:this={btnEl}
	>
		<Icon i="tag" wh={19} />
	</button>

	{#if menuOpen}
		<TagMenu
			titleText="Add To Tag"
			selectedTags={watchedItem.tags}
			{onTagClick}
			arrowColor="white"
			trackEl={btnEl}
		/>
	{/if}
</div>

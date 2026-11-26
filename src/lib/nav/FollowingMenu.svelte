<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";
	import { resolve } from "$app/paths";

	interface Props extends Pick<MenuConfig, "trackEl"> {
		close: () => void;
	}

	let { close, trackEl }: Props = $props();
</script>

<Menu width="180px" {trackEl}>
	{#if store.follows?.length > 0}
		<h4 class="norm sm-caps">following</h4>
		<div class="list">
			{#each store.follows as f (f.followedUser.id)}
				<a
					href={resolve(
						`/lists/${f.followedUser.id}/${f.followedUser.username}`,
					)}
					onclick={() => close()}
				>
					{f.followedUser.username}
				</a>
			{/each}
		</div>
	{:else}
		<span style="margin-top: 0;">You are not following anyone.</span>
	{/if}
</Menu>

<style lang="scss">
	h4 {
		position: sticky;
		top: -10px;
		background-color: $bg-color;
	}

	.list {
		list-style: none;
		display: flex;
		flex-flow: column;
		width: 100%;
		height: 100%;

		a {
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
		}
	}
</style>

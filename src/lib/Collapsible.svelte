<script lang="ts">
	import Icon from "./Icon.svelte";

	interface Props {
		title: string;
		children?: import("svelte").Snippet;
	}

	let { title, children }: Props = $props();

	let collapsed = $state(true);
</script>

<button class="plain" onclick={() => (collapsed = !collapsed)}>
	<Icon i="chevron" facing={collapsed ? "right" : "down"} wh={19} />
	{title}
</button>

{#if !collapsed}
	{@render children?.()}
{/if}

<style lang="scss">
	button.plain {
		display: flex;
		align-items: center;
		justify-content: start;
		gap: 3px;
	}
</style>

<script lang="ts">
	interface Props {
		title: string | undefined;
		children?: import("svelte").Snippet;
	}

	let { title, children }: Props = $props();
</script>

<div>
	{#if title}
		<h2>{title}</h2>
	{/if}
	<ul>
		{@render children?.()}
	</ul>
</div>

<style lang="scss">
	div {
		width: 100%;
		overflow-x: auto;
	}

	h2 {
		font-size: 30px;
		font-weight: bold;
		margin-left: 30px;
		position: sticky;
		left: 0;
	}

	ul {
		display: flex;
		flex-flow: row;
		gap: 10px;
		list-style: none;
		padding: 15px 30px;
		justify-content: flex-start;
	}
</style>

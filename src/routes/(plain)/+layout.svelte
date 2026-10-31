<script lang="ts">
	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();
</script>

<nav>
	<h1>Watcharr</h1>
</nav>

{@render children?.()}

<style lang="scss">
	nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin: 10px 20px 28px 20px;
		position: relative;

		h1 {
			color: transparent;
			-webkit-text-stroke: 1.5px $text-color;
			font-size: 35px;
			user-select: none;
		}
	}
</style>

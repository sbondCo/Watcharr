<script lang="ts">
	import type { Snippet } from "svelte";

	type AcceptedColor = "success" | "error";

	interface Props {
		body: string;
		bg?: AcceptedColor;
		icon?: Snippet;
	}

	let { body, icon, bg }: Props = $props();
</script>

<div data-bg={bg}>
	{#if icon}
		<span class="icon-wrap">
			{@render icon()}
		</span>
	{/if}
	<span>{body}</span>
</div>

<style lang="scss">
	div {
		display: flex;
		align-items: center;
		gap: 3px;
		border-radius: 10px;
		padding: 3px 7px;
		user-select: none;

		background-color: $accent-color;

		&[data-bg="success"] {
			background-color: $success;
			color: white;
		}

		&[data-bg="error"] {
			background-color: $error;
		}

		.icon-wrap {
			display: flex;
		}
	}
</style>

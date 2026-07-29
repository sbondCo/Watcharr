<script lang="ts">
	import type { Tag } from "@/types";

	interface Props {
		tag: Tag;
		onClick?: () => void | undefined;
	}

	let { tag, onClick = undefined! }: Props = $props();
</script>

<button
	class="plain"
	style:color={tag.color}
	style:background={tag.bgColor}
	onclick={() => {
		if (typeof onClick === "function") {
			onClick();
		}
	}}
>
	{tag.name}
</button>

<style lang="scss">
	button {
		text-transform: capitalize;
		position: relative;
		width: max-content;
		border-radius: 8px;
		padding: 5px 8px;
		text-wrap: wrap;
		word-break: break-word;
		transition: opacity 150ms ease-in-out;

		&:hover {
			opacity: 0.8;
		}
	}
</style>

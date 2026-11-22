<script lang="ts">
	interface Props {
		count: number;
	}

	let { count }: Props = $props();
</script>

<div>{count}</div>

<style lang="scss">
	div {
		font-family:
			"Shrikhand",
			system-ui,
			-apple-system,
			BlinkMacSystemFont;
		font-size: 13px;
		color: black;
		background-color: $gold;
		padding: 3px 6px;
		border-radius: 50px;
		min-width: 25px;
		min-height: 25px;
		height: min-content;
		display: flex;
		align-items: center;
		justify-content: center;
	}
</style>

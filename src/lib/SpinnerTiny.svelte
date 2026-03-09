<script lang="ts">
	interface Props {
		style?: string;
	}

	let { style = "" }: Props = $props();
</script>

<div class="loader" {style}></div>

<style lang="scss">
	.loader {
		width: 16px;
		height: 12px;
		border: 2px solid $text-color;
		border-bottom-color: transparent;
		border-radius: 50%;
		display: inline-block;
		box-sizing: border-box;
		animation: rotation 1s linear infinite;
	}

	@keyframes rotation {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>

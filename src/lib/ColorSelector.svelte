<!-- Uses Native Color Input -->

<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	interface Props {
		value?: string;
		style?: string;
	}

	let { value = $bindable("#000000"), style = "" }: Props = $props();
</script>

<button
	aria-label={t("common.openColorPicker")}
	class="plain"
	style={`background-color: ${value};${style}`}
>
	<input type="color" bind:value />
</button>

<style lang="scss">
	button {
		width: 100%;
		height: 100%;
		border-radius: 10px;
		outline: 1px solid $accent-color;
		cursor: pointer;
	}

	input {
		opacity: 0;
		cursor: pointer;
	}
</style>

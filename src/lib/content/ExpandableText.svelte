<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	interface Props {
		text?: string;
		title?: string;
		style?: string;
	}

	const { text, title, style }: Props = $props();

	const splitText:
		| {
				// The preview we show.
				preview: string;
				rest?: string;
		  }
		| undefined = $derived.by(() => {
		if (!text) {
			return;
		}
		let index = text.indexOf("\n");
		if (index != -1) {
			// If index found..
			return {
				preview: text.slice(0, index),
				rest: text.slice(index + 2),
			};
		}
		return {
			preview: text,
		};
	});
	let bioExtended = $state(false);
</script>

<div {style}>
	{#if splitText}
		{#if title}
			<span>{title}</span>
		{/if}
		<p>
			{splitText.preview}
			{#if bioExtended}
				<br /><br />{splitText.rest}
			{/if}
			{#if splitText.rest}
				<button class="plain" onclick={() => (bioExtended = !bioExtended)}>
					{bioExtended ? t("common.less") : t("common.more")}
				</button>
			{/if}
		</p>
	{/if}
</div>

<style lang="scss">
	div {
		display: flex;
		flex-flow: column;
		gap: 3px;

		span {
			font-weight: bold;
			font-size: 14px;
		}

		p {
			font-size: 14px;
			white-space: pre-line;
		}

		button {
			font-weight: bold;
			color: white;
		}
	}
</style>

<!--
	@component
	Render a translation that contains simple markup, without using `{@html}`
	(so translations can never inject html).

	Supported in translations (not nested):
	- `<b>`, `<strong>`, `<i>`, `<em>` and `<code>`, rendered as those elements.
	- Any other `<name>text</name>`, rendered with the snippet passed as a prop
	  with the same name (eg a link):

	```svelte
	<Trans key="handbook.example">
		{#snippet link(text)}<a href="https://watcharr.app">{text}</a>{/snippet}
	</Trans>
	```
-->
<script lang="ts">
	import type { Snippet } from "svelte";
	import type { TOptions } from "i18next";
	import { t } from "./index.svelte";

	interface Props {
		key: string;
		values?: TOptions;
		[tag: string]: Snippet<[string]> | string | TOptions | undefined;
	}

	let { key, values, ...tags }: Props = $props();

	const elements = ["b", "strong", "i", "em", "code"];

	let parts = $derived.by(() => {
		const str = t(key, values);
		const result: { text: string; tag?: string }[] = [];
		const re = /<(\w+)>([\s\S]*?)<\/\1>/g;
		let last = 0;
		for (const m of str.matchAll(re)) {
			if (m.index > last) result.push({ text: str.slice(last, m.index) });
			result.push({ text: m[2], tag: m[1] });
			last = m.index + m[0].length;
		}
		if (last < str.length) result.push({ text: str.slice(last) });
		return result;
	});

	function getSnippet(tag: string) {
		const s = tags[tag];
		return typeof s === "function" ? (s as Snippet<[string]>) : undefined;
	}
</script>

{#each parts as part, i (i)}{#if !part.tag}{part.text}{:else if getSnippet(part.tag)}{@render getSnippet(
			part.tag,
		)?.(part.text)}{:else if elements.includes(part.tag)}<svelte:element
			this={part.tag}>{part.text}</svelte:element
		>{:else}{part.text}{/if}{/each}

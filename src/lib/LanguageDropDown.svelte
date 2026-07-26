<script lang="ts">
	import axios from "axios";
	import DropDown from "./DropDown.svelte";
	import type { DropDownItem } from "@/types";
	import Error from "./Error.svelte";

	interface Props {
		selectedLang?: string;
		disabled?: boolean;
		onChange: (lang: string) => void;
	}

	let {
		selectedLang = $bindable("en-US"),
		disabled = false,
		onChange,
	}: Props = $props();

	let mappedLangs: DropDownItem[] = $state<DropDownItem[]>([]);

	async function getLanguages() {
		const l = (await axios.get(`/content/languages`)).data as {
			code: string;
			name: string;
		}[];
		mappedLangs = l.map((ll) => ({ id: ll.code, value: ll.name }) as DropDownItem);
	}
</script>

{#await getLanguages() then}
	<DropDown
		placeholder="Select a language"
		bind:active={selectedLang}
		options={mappedLangs}
		onChange={() => onChange(selectedLang)}
		isDropDownItem={true}
		{disabled}
	/>
{:catch err}
	<Error error={err} pretty="Failed to load languages!" />
{/await}

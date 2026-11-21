<script lang="ts">
	import DropDown from "./DropDown.svelte";
	import type { DropDownItem } from "@/types";
	import { availableLanguages, t } from "./i18n/index.svelte";

	interface Props {
		/**
		 * Selected language code, empty for automatic (browser language).
		 */
		selectedLanguage?: string;
		disabled?: boolean;
		onChange: (language: string) => void;
	}

	let {
		selectedLanguage = $bindable(""),
		disabled = false,
		onChange,
	}: Props = $props();

	let options: DropDownItem[] = $derived([
		{ id: "", value: t("settings.language.automatic") },
		...availableLanguages.map((l) => ({ id: l.code, value: l.name })),
	]);
</script>

<DropDown
	placeholder={t("settings.language.select")}
	bind:active={selectedLanguage}
	{options}
	onChange={() => onChange(selectedLanguage)}
	isDropDownItem={true}
	{disabled}
/>

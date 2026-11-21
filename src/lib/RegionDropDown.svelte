<script lang="ts">
	import { defaultLanguage, getLanguage, t } from "@/lib/i18n/index.svelte";
	import DropDown from "./DropDown.svelte";
	import type { DropDownItem, TMDBRegions } from "@/types";
	import Error from "./Error.svelte";
	import { req } from "./util/api";

	interface Props {
		selectedCountry?: string;
		disabled?: boolean;
		onChange: (country: string) => void;
	}

	let {
		selectedCountry = $bindable("US"),
		disabled = false,
		onChange,
	}: Props = $props();

	let regions: TMDBRegions["results"] = $state([]);

	// Show country names in the users language when possible
	// (TMDB only gives us english names).
	let mappedCountries: DropDownItem[] = $derived.by(() => {
		const lang = getLanguage();
		let names: Intl.DisplayNames | undefined;
		if (lang !== defaultLanguage) {
			try {
				names = new Intl.DisplayNames([lang], { type: "region" });
			} catch (err) {
				console.warn("RegionDropDown: Intl.DisplayNames unsupported", err);
			}
		}
		return regions
			.map((cc) => {
				let value = cc.english_name;
				try {
					value = names?.of(cc.iso_3166_1) ?? cc.english_name;
				} catch {
					// Invalid region code, keep the english name.
				}
				return { id: cc.iso_3166_1, value } as DropDownItem;
			})
			.sort((a, b) => a.value.localeCompare(b.value, lang));
	});

	async function getCountries() {
		const c = await req.get<TMDBRegions>(`/content/regions`);
		regions = c.results;
	}
</script>

{#await getCountries() then}
	<DropDown
		placeholder={t("settings.country.select")}
		bind:active={selectedCountry}
		options={mappedCountries}
		onChange={() => onChange(selectedCountry)}
		isDropDownItem={true}
		{disabled}
	/>
{:catch err}
	<Error error={err} pretty={t("settings.country.failedToLoad")} />
{/await}

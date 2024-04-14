<script lang="ts">
  import axios from "axios";
  import DropDown from "./DropDown.svelte";
  import type { DropDownItem, TMDBRegions } from "@/types";
  import Error from "./Error.svelte";

  export let selectedCountry: string = "US";
  export let disabled = false;
  export let onChange: (country: string) => void;

  let mappedCountries: DropDownItem[];

  async function getCountries() {
    const c = (await axios.get(`/content/regions`)).data as TMDBRegions;
    mappedCountries = c.results
      .map((cc) => {
        return {
          id: cc.iso_3166_1,
          value: cc.english_name
        } as DropDownItem;
      })
      .sort((a, b) => a.value.localeCompare(b.value));
  }
</script>

{#await getCountries() then}
  <DropDown
    placeholder="Select a country"
    bind:active={selectedCountry}
    options={mappedCountries}
    onChange={() => onChange(selectedCountry)}
    isDropDownItem={true}
    {disabled}
  />
{:catch err}
  <Error error={err} pretty="Failed to load countries!" />
{/await}

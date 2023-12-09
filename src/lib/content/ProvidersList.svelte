<script lang="ts">
  import type { TMDBContentWatchProviders } from "@/types";
  import ProviderIcon from "./ProviderIcon.svelte";

  export let providers: TMDBContentWatchProviders;
</script>

{#if providers?.flatrate?.length > 0}
  <div class="streaming-providers">
    {#each providers.flatrate as provider}
      <ProviderIcon i={provider.provider_name} wh={40} />
    {/each}
  </div>
{/if}

<style lang="scss">
  .streaming-providers {
    display: flex;
    align-items: flex-end;
    flex-wrap: wrap;
    gap: 15px;
    margin-top: auto;
    padding-top: 10px;
  }
</style>

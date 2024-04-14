<script lang="ts">
  import type { TMDBContentWatchProviders } from "@/types";
  import ProviderIcon from "./ProviderIcon.svelte";

  export let providers: TMDBContentWatchProviders;
</script>

{#if providers?.flatrate?.length > 0 || providers?.free?.length > 0}
  <div class="streaming-providers">
    {#if providers?.flatrate?.length > 0}
      {#each providers.flatrate as provider}
        <ProviderIcon i={provider.provider_name} wh={40} />
      {/each}
    {/if}
    {#if providers?.free?.length > 0}
      {#each providers.free as provider}
        <ProviderIcon i={provider.provider_name} wh={40} />
      {/each}
    {/if}
  </div>
{/if}

<style lang="scss">
  .streaming-providers {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 15px;
    margin-top: auto;
    padding-top: 10px;
  }
</style>

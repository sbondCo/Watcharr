<script>
  import Icon from "@/lib/Icon.svelte";
  import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
  import { unNotify } from "@/lib/util/notify";
  import { notifications } from "@/store";
  import { pwaInfo } from 'virtual:pwa-info'; 

  $: webManifestLink = pwaInfo ? pwaInfo.webManifest.linkTag : '' 
  $: notifs = $notifications;

  console.log(
    `%cWATCHARR v${__WATCHARR_VERSION__}`,
    "background: white;color: black;font-size: large;padding: 3px 5px;"
  );
</script>

<svelte:head> 
 	{@html webManifestLink} 
</svelte:head>
<div id="tooltip" />
<div id="notifications">
  {#each notifs as n}
    <div class={`${n.type} notif`}>
      {#if n.type === "loading"}
        <SpinnerTiny />
      {/if}
      <span>{@html n.text}</span>
      <button
        class="plain"
        on:click={() => {
          unNotify(n.id);
        }}
      >
        <Icon i="close" />
      </button>
    </div>
  {/each}
</div>
<slot />

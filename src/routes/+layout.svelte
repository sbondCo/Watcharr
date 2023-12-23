<script>
  import Icon from "@/lib/Icon.svelte";
  import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
  import { unNotify } from "@/lib/util/notify";
  import { notifications } from "@/store";
  import { onMount } from "svelte";
  import { pwaInfo } from "virtual:pwa-info";

  $: notifs = $notifications;

  console.log(
    `%cWATCHARR v${__WATCHARR_VERSION__}`,
    "background: white;color: black;font-size: large;padding: 3px 5px;"
  );

  function resetTooltipPos() {
    const t = document.getElementById("tooltip");
    if (t) {
      t.style.top = "0";
      t.style.left = "0";
    }
  }

  onMount(() => {
    window.addEventListener("resize", resetTooltipPos);

    return () => {
      window.removeEventListener("resize", resetTooltipPos);
    };
  });
</script>

<svelte:head>
  {#if pwaInfo?.webManifest?.linkTag}
    <!-- eslint-disable-next-line -->
    {@html pwaInfo.webManifest.linkTag}
  {/if}
</svelte:head>

<div id="tooltip" />
<div id="notifications">
  {#each notifs as n}
    <div class={`${n.type} notif`}>
      {#if n.type === "loading"}
        <SpinnerTiny />
      {/if}
      <!-- only comes from our strings (which may have html) -->
      <!-- eslint-disable-next-line -->
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

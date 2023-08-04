<script>
  // import "@/norm.scss";
  import Icon from "@/lib/Icon.svelte";
  import { unNotify } from "@/lib/util/notify";
  import { notifications } from "@/store";

  $: notifs = $notifications;

  console.log(
    `%cWATCHARR v${__WATCHARR_VERSION__}`,
    "background: white;color: black;font-size: large;padding: 3px 5px;"
  );
</script>

<div id="tooltip" />
<div id="notifications">
  {#each notifs as n}
    <div class={n.type}>
      <span>{n.text}</span>
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

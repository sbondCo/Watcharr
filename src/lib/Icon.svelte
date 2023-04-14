<script lang="ts" context="module">
  export type Icon = "check" | "clock" | "calendar" | "jellyfin";
</script>

<script lang="ts">
  export let i: Icon;
  export let wh = 24;
</script>

{#if i === "check"}
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path
      fill="none"
      stroke="currentColor"
      stroke-linecap="square"
      stroke-miterlimit="10"
      stroke-width="44"
      d="M416 128L192 384l-96-96"
    />
  </svg>
{:else if i === "clock"}
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path
      d="M256 48C141.13 48 48 141.13 48 256c0 114.69 93.32 208 208 208 114.86 0 208-93.14 208-208 0-114.69-93.31-208-208-208zm108 240H244a4 4 0 01-4-4V116a4 4 0 014-4h24a4 4 0 014 4v140h92a4 4 0 014 4v24a4 4 0 01-4 4z"
    />
  </svg>
{:else if i === "calendar"}
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path
      d="M32 456a24 24 0 0024 24h400a24 24 0 0024-24V176H32zm80-238.86a9.14 9.14 0 019.14-9.14h109.72a9.14 9.14 0 019.14 9.14v109.72a9.14 9.14 0 01-9.14 9.14H121.14a9.14 9.14 0 01-9.14-9.14zM456 64h-55.92V32h-48v32H159.92V32h-48v32H56a23.8 23.8 0 00-24 23.77V144h448V87.77A23.8 23.8 0 00456 64z"
    />
  </svg>
{:else if i === "jellyfin"}
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width={wh} height={wh}>
    <path
      d="M12 .002C8.826.002-1.398 18.537.16 21.666c1.56 3.129 22.14 3.094 23.682 0C25.384 18.573 15.177 0 12 0zm7.76 18.949c-1.008 2.028-14.493 2.05-15.514 0C3.224 16.9 9.92 4.755 12.003 4.755c2.081 0 8.77 12.166 7.759 14.196zM12 9.198c-1.054 0-4.446 6.15-3.93 7.189.518 1.04 7.348 1.027 7.86 0 .511-1.027-2.874-7.19-3.93-7.19z"
    />
  </svg>
{/if}

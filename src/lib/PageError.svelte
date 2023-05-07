<script lang="ts">
  export let pretty: string;
  export let error: any;
</script>

<div>
  <div>
    <strong>{pretty}</strong>
    {#if error?.message}
      <p>{error.message}</p>
      {#if error.response}
        <p>{error.response.data.error}</p>
      {/if}
    {:else}
      <p>{JSON.stringify(error)}</p>
    {/if}
  </div>
</div>

<style lang="scss">
  div {
    display: flex;
    justify-content: center;
    width: 100%;

    div {
      display: flex;
      justify-content: center;
      flex-flow: column;
      max-width: 500px;
      gap: 15px;

      p {
        color: white;
        background-color: rgb(66, 60, 60);
        padding: 10px;
        border-radius: 8px;
        text-transform: capitalize;
      }
    }
  }
</style>

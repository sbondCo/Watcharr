<script lang="ts">
  export let pretty: string;
  export let error: Error;
</script>

<div>
  <div>
    <strong>{pretty}</strong>
    <p>{error.message}</p>
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
      margin: 0 20px;

      strong {
        margin-bottom: 10px;
      }

      p {
        color: white;
        background-color: rgb(66, 60, 60);
        padding: 10px;
        border-radius: 8px;
      }
    }
  }
</style>

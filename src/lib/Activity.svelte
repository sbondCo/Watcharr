<script lang="ts">
  export let tvId: number;
</script>

<div class="activity">
  <h2>Activity</h2>
  <ul>
    <li>
      {new Date().toLocaleDateString()} Status changed to hold
    </li>
    <li>{new Date().toLocaleDateString()} Rating changed to 5</li>
    <li>{new Date().toLocaleDateString()} Added show as watching</li>
  </ul>
</div>

<style lang="scss">
  .activity {
    width: 100%;

    ul {
      display: flex;
      flex-flow: column;
      gap: 8px;
      margin-top: 8px;
      margin-left: calc(30px + 8px);
      list-style: none;

      li {
        width: max-content;
        padding: 10px 12px;
        color: white;
        background-color: rgb(46, 46, 46);
        border-radius: 8px;
      }
    }
  }

  h2 {
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
    font-size: 30px;
    font-weight: bold;
    margin-left: 30px;
  }
</style>

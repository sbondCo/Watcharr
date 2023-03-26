<script lang="ts">
  import type { Content } from "@/types";

  // export let content: Content;
  export let poster: string;
</script>

<li>
  <div>
    <img src={poster} alt="poster" />
    <!-- <span>{content.title}</span> -->
  </div>
</li>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    background-color: rgb(48, 45, 45);
    overflow: hidden;
    flex: 1 1;
    border-radius: 5px;
    width: 170px;
    // aspect-ratio: 2/3;

    span {
      color: white;
      padding: 5px 8px;
    }
  }
</style>

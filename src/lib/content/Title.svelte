<script lang="ts">
  export let homepage: string;
  export let title: string;
  export let releaseDate: string;
  export let voteAverage: number;
  export let voteCount: number;

  const vote = Math.round(voteAverage * 10) / 10;
</script>

<span class="title-container">
  <span class="title">
    <a href={homepage} target="_blank">{title}</a>
    <span>{new Date(Date.parse(releaseDate)).getFullYear()}</span>
  </span>
  <span class="rating" title={`TMDB Rating: ${vote} out of 10 (based on ${voteCount} votes)`}>
    <span>*</span>
    {vote}
  </span>
</span>

<style lang="scss">
  .title-container {
    display: flex;
    gap: 10px;

    .title {
      a {
        color: white;
        text-decoration: none;
        font-size: 30px;
        font-weight: bold;
        padding-right: 3px;
      }

      span {
        font-size: 20px;
        color: rgba($color: #fff, $alpha: 0.7);
      }
    }

    .rating {
      display: flex;
      align-items: start;
      justify-content: center;
      gap: 5px;
      color: green;
      margin-left: auto;
      font-size: 22px;
      color: gold;
      font-weight: bolder;

      span {
        font-family: "Rampart One";
        -webkit-text-stroke: 1px gold;
        font-size: 40px;
        line-height: 0.7;
        margin-top: 7px;
      }
    }
  }
</style>

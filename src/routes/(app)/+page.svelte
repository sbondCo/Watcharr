<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { updateWatched } from "@/lib/api";
  import { watchedList } from "@/store";

  $: watched = $watchedList;
</script>

<svelte:head>
  <title>Watched List</title>
</svelte:head>

<PosterList>
  {#if watched?.length > 0}
    {#each watched as w (w.id)}
      <a data-sveltekit-preload-data="tap" href={`/${w.content.type}/${w.content.id}`}>
        <Poster
          poster={w.content.poster_path
            ? "http://localhost:3080/img" + w.content.poster_path
            : undefined}
          title={w.content.title}
          desc={w.content.overview}
          rating={w.rating}
          status={w.status}
          onBtnClicked={(type) => {
            updateWatched(w.id, type)
              ?.then(() => {
                w.status = type;
                $watchedList = watched;
              })
              .catch((err) => {
                console.error(err);
              });
          }}
          onRatingChanged={(rating) => {
            updateWatched(w.id, undefined, rating)
              ?.then(() => {
                w.rating = rating;
                $watchedList = watched;
              })
              .catch((err) => {
                console.error(err);
              });
          }}
        />
      </a>
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>

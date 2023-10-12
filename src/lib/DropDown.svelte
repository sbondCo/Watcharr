<script lang="ts">
  import Icon from "./Icon.svelte";

  export let options: string[];
  export let active: string | undefined;
  export let placeholder: string;
  export let blendIn: boolean = false;
  export let disabled = false;

  let availableOptions = options.filter((o) => o !== active);
  let open = false;
</script>

<div
  class={[
    open ? "is-open" : "",
    !active ? "placeholder-shown" : "",
    blendIn ? "blend-in" : ""
  ].join(" ")}
>
  <button on:click={() => (open = !open)} {disabled}>
    {active ? active : placeholder}
    <Icon i="chevron" facing={open ? "up" : "down"} />
  </button>
  <ul>
    {#each availableOptions as o}
      <li>
        <button
          class="plain"
          on:click={() => {
            active = o;
            open = false;
          }}
        >
          {o}
        </button>
      </li>
    {/each}
  </ul>
</div>

<style lang="scss">
  div {
    position: relative;

    &.placeholder-shown {
      & > button {
        color: #8e8e8e;

        &:hover:not(:disabled) {
          color: $bg-color;
        }
      }
    }

    &.blend-in {
      & > button {
        border-color: transparent;
        background-color: transparent;

        &:hover:not(:disabled),
        &:focus-visible:not(:disabled) {
          background-color: $text-color;
        }
      }
    }

    &.is-open {
      ul {
        display: flex;
      }

      & > button {
        border-color: $text-color;
        border-bottom-left-radius: 0;
        border-bottom-right-radius: 0;
      }
    }

    button {
      gap: 3px;
      text-transform: capitalize;

      :global(svg) {
        margin-left: auto;
      }
    }

    ul {
      display: none;
      flex-flow: column;
      position: absolute;
      list-style: none;
      width: 100%;
      font-size: 14px;
      border: 2px solid $text-color;
      border-top: 0;
      border-bottom-left-radius: 5px;
      border-bottom-right-radius: 5px;
      background-color: $bg-color;
      z-index: 99;

      li {
        width: 100%;

        button {
          padding: 5px 10px;
          width: 100%;
          text-align: start;
          text-transform: capitalize;
          transition: background-color 100ms ease;

          &:hover:not(:disabled),
          &:focus-visible:not(:disabled) {
            background-color: $text-color;
            color: $bg-color;
            fill: $bg-color;
            opacity: 1;
          }
        }
      }
    }
  }
</style>
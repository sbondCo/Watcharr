<script lang="ts">
  import type { DropDownItem } from "@/types";
  import Icon from "./Icon.svelte";
  import { onMount } from "svelte";

  export let options: string[] | DropDownItem[];
  // If we are using DropDownItems[] as options.
  export let isDropDownItem = false;
  export let active: string | number | undefined = undefined;
  export let placeholder: string;
  export let blendIn: boolean = false;
  export let disabled = false;
  export let onChange: () => void = () => {};

  let activeValue: string;
  let open = false;
  let ulElement: HTMLUListElement;

  function handleKeyPress(event: KeyboardEvent) {
    if (!open || disabled) return; // Don't handle if closed or disabled

    const pressedLetter = event.key.toLowerCase();

    // Filter options that start with the pressed letter (ignoring case)
    const filteredOptions = options.filter((o) =>
      typeof o === "string"
        ? o.toLowerCase().startsWith(pressedLetter)
        : o.value.toLowerCase().startsWith(pressedLetter)
    );

    // If there are filtered options, select the first one
    let f: string | number | undefined;
    if (filteredOptions.length > 0) {
      f = typeof filteredOptions[0] === "string" ? filteredOptions[0] : filteredOptions[0].id;
    }

    // Find first button with text content starting with letter pressed and scroll it into view.
    const btns = ulElement?.querySelectorAll("button");
    for (let i = 0; i < btns.length; i++) {
      const btn = btns[i];
      if (btn.textContent?.toLowerCase()?.startsWith(pressedLetter)) {
        btn.scrollIntoView({ behavior: "smooth", block: "start" });
        break;
      }
    }
  }

  $: {
    if (typeof active === "string" && !isDropDownItem) {
      activeValue = active;
    } else {
      const v = options.find((o) => (typeof o !== "string" ? o.id === active : false));
      if (v && typeof v !== "string") activeValue = v.value;
    }
  }

  onMount(() => {
    document.addEventListener("keypress", handleKeyPress);

    return () => {
      document.removeEventListener("keypress", handleKeyPress);
    };
  });
</script>

<div
  class={[
    open ? "is-open" : "",
    typeof active === "undefined" ? "placeholder-shown" : "",
    blendIn ? "blend-in" : ""
  ].join(" ")}
>
  <button on:click={() => (open = !open)} {disabled}>
    {activeValue ? activeValue : placeholder}
    <Icon i="chevron" facing={open ? "up" : "down"} />
  </button>
  <ul bind:this={ulElement}>
    {#each options.filter((o) => (typeof o === "string" ? o !== active : o.id !== active)) as o}
      <li>
        <button
          class="plain"
          on:click={() => {
            active = typeof o == "string" ? o : o.id;
            open = false;
            onChange();
          }}
        >
          {typeof o == "string" ? o : o.value}
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
      max-height: 20vh;
      overflow-y: auto;

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

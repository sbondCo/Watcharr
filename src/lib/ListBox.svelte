<script lang="ts">
  import type { ListBoxItem } from "@/types";
  import Checkbox from "./Checkbox.svelte";

  export let options: ListBoxItem[];

  /**
   * If a checkbox to select all checkboxes is wanted,
   * provide the text to display besides it.
   */
  export let allCheckBox: string | undefined = undefined;

  let allCheckBoxValue: boolean = false;
</script>

<div>
  {#if allCheckBox}
    <div>
      <Checkbox
        name={allCheckBox}
        bind:value={allCheckBoxValue}
        toggled={(on) => {
          for (let i = 0; i < options.length; i++) {
            const e = options[i];
            e.value = on;
          }
          options = options;
        }}
      />
      <span>{allCheckBox}</span>
    </div>
  {/if}
  {#each options as o}
    <div>
      <Checkbox
        name={o.displayValue}
        bind:value={o.value}
        toggled={(on) => {
          if (!on) {
            allCheckBoxValue = false;
          } else if (!allCheckBoxValue) {
            if (!options.some((o) => !o.value)) {
              allCheckBoxValue = true;
            }
          }
        }}
      />
      <span>{o.displayValue}</span>
    </div>
  {/each}
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    gap: 10px;

    & > div {
      display: flex;
      flex-flow: row;
      gap: 10px;
    }
  }
</style>

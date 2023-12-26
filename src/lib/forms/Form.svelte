<script lang="ts">
  import type { HTMLInputTypeAttribute } from "svelte/elements";
  // export let title: string;
  // export let desc: string = "";
  // export let row: boolean = false;
  export let labelFor: string;
  export let labelText: string;
  export let inputType: HTMLInputTypeAttribute;
  export let inputName: string;
  export let inputPlaceholder: string;
</script>

<div class="setting-ctr">
  <div class="row">
    <div>
      <label for={labelFor}>{labelText}</label>
    </div>
    <div>
      <input type={inputType} name={inputName} placeholder={inputPlaceholder} />
    </div>
  </div>
</div>

<style>
  .setting-ctr {
    display: flex;
    flex-direction: row;
    margin-bottom: 1rem;
  }

  .row {
    display: flex;
    flex-direction: column;
    /*justify-content: space-between;*/
    /*align-items: center;*/
  }

  label {
    font-size: 1.2rem;
    font-weight: 600;
    margin-right: 1rem;
  }

  input {
    font-size: 1.2rem;
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid #ccc;
    outline: none;
  }
</style>

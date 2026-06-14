<script lang="ts">
  import { onMount } from "svelte";

  export let modelPath: string;
  export let vaePath: string;

  let models: string[] = [];
  let vaes: string[] = [];

  onMount(async () => {
    try {
      models = await window.runtime.Call("ListModels") || [];
    } catch (e) {
      console.error("ListModels failed:", e);
    }
  });

  function basename(path: string): string {
    return path.split("/").pop() || path;
  }
</script>

<div class="field">
  <label for="model">Model</label>
  <select id="model" bind:value={modelPath}>
    <option value="">-- Select model --</option>
    {#each models as m}
      <option value={m}>{basename(m)}</option>
    {/each}
  </select>
</div>

<div class="field">
  <label for="vae">VAE (optional)</label>
  <select id="vae" bind:value={vaePath}>
    <option value="">-- No VAE --</option>
    {#each vaes as v}
      <option value={v}>{basename(v)}</option>
    {/each}
  </select>
</div>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  label {
    font-size: 12px;
    font-weight: 500;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  select {
    background: #25252b;
    border: 1px solid #333;
    color: #e0e0e0;
    padding: 6px 10px;
    border-radius: 6px;
    font-size: 13px;
  }

  select:focus {
    outline: none;
    border-color: #7c3aed;
  }
</style>

<script lang="ts">
  import { onMount } from "svelte";
  import { ListModels } from "../../wailsjs/go/orchestrator/Manager";

  export let modelPath: string = "";
  export let vaePath: string = "";

  let models: string[] = [];
  let vaes: string[] = [];
  let loading = true;
  let errorMsg = "";

  onMount(async () => {
    try {
      const result = await ListModels();
      console.log("ListModels result:", result);
      models = result || [];
    } catch (e: any) {
      errorMsg = String(e?.message || e);
      console.error("ListModels error:", e);
    } finally {
      loading = false;
    }
  });

  function basename(path: string): string {
    const parts = path.split("/");
    return parts[parts.length - 1] || path;
  }
</script>

{#if errorMsg}
  <div class="error">Error: {errorMsg}</div>
{/if}

<div class="field">
  <label for="model">Model</label>
  <select id="model" bind:value={modelPath}>
    <option value="">
      {loading ? "Loading..." : models.length === 0 ? "No models found in ~/.comfygo/models/" : "-- Select model --"}
    </option>
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
  .error {
    background: #3a1a1a;
    color: #f87171;
    padding: 8px;
    border-radius: 6px;
    font-size: 12px;
    margin-bottom: 8px;
  }

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

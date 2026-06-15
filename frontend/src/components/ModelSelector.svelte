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
    background: rgba(239,68,68,0.15);
    color: var(--red);
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
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  select {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-primary);
    padding: 6px 10px;
    border-radius: 6px;
    font-size: 13px;
    -webkit-appearance: none;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2364748B' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 8px center;
    padding-right: 28px;
    cursor: pointer;
  }

  select option {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  select:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }
</style>

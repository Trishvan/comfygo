<script lang="ts">
  let { modelPath, vaePath } = $props();

  let models = $state<string[]>([]);
  let vaes = $state<string[]>([]);

  async function listFiles(dir: string): Promise<string[]> {
    try {
      return await window.runtime.Call("ListModels", dir) || [];
    } catch {
      return [];
    }
  }

  $effect(() => {
    const dir = localStorage.getItem("comfygo-model-dir") || "";
    if (dir) {
      listFiles(dir).then((m) => { models = m; });
    }
  });
</script>

<div class="field">
  <label for="model">Model</label>
  <select id="model" bind:value={modelPath}>
    <option value="">-- Select model --</option>
    {#each models as m}
      <option value={m}>{m}</option>
    {/each}
  </select>
</div>

<div class="field">
  <label for="vae">VAE (optional)</label>
  <select id="vae" bind:value={vaePath}>
    <option value="">-- No VAE --</option>
    {#each vaes as v}
      <option value={v}>{v}</option>
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

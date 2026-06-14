<script lang="ts">
  import PromptInput from "./PromptInput.svelte";
  import ModelSelector from "./ModelSelector.svelte";

  let { state, onGenerate, onCancel } = $props();

  let prompt = $state("");
  let negativePrompt = $state("");
  let modelPath = $state("");
  let vaePath = $state("");
  let steps = $state(20);
  let cfgScale = $state(7.0);
  let seed = $state(-1);
  let width = $state(512);
  let height = $state(512);

  function submit() {
    onGenerate({
      prompt,
      negativePrompt,
      modelPath,
      vaePath,
      steps,
      cfgScale,
      seed,
      width,
      height,
      samplerName: "euler_a",
    });
  }

  let isBusy = $derived(state === "loading" || state === "generating");
</script>

<form onsubmit={(e) => { e.preventDefault(); submit(); }}>
  <ModelSelector bind:modelPath bind:vaePath />

  <PromptInput label="Prompt" bind:value={prompt} />
  <PromptInput label="Negative Prompt" bind:value={negativePrompt} />

  <div class="field">
    <label for="steps">Steps: {steps}</label>
    <input id="steps" type="range" min="1" max="50" bind:value={steps} disabled={isBusy} />
  </div>

  <div class="field">
    <label for="cfg">CFG Scale: {cfgScale.toFixed(1)}</label>
    <input id="cfg" type="range" min="1" max="20" step="0.5" bind:value={cfgScale} disabled={isBusy} />
  </div>

  <div class="field">
    <label for="seed">Seed (-1 = random)</label>
    <input id="seed" type="number" bind:value={seed} disabled={isBusy} />
  </div>

  <div class="field-row">
    <div class="field">
      <label for="width">Width</label>
      <select id="width" bind:value={width} disabled={isBusy}>
        <option value={512}>512</option>
        <option value={640}>640</option>
        <option value={768}>768</option>
      </select>
    </div>
    <div class="field">
      <label for="height">Height</label>
      <select id="height" bind:value={height} disabled={isBusy}>
        <option value={512}>512</option>
        <option value={640}>640</option>
        <option value={768}>768</option>
      </select>
    </div>
  </div>

  <div class="actions">
    {#if isBusy}
      <button type="button" class="btn btn-cancel" onclick={onCancel}>Cancel</button>
    {:else}
      <button type="submit" class="btn btn-generate">Generate</button>
    {/if}
  </div>
</form>

<style>
  form {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field label {
    font-size: 12px;
    font-weight: 500;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .field input[type="range"] {
    width: 100%;
    accent-color: #a78bfa;
  }

  .field input[type="number"],
  .field select {
    background: #25252b;
    border: 1px solid #333;
    color: #e0e0e0;
    padding: 6px 10px;
    border-radius: 6px;
    font-size: 14px;
  }

  .field-row {
    display: flex;
    gap: 12px;
  }

  .field-row .field {
    flex: 1;
  }

  .actions {
    margin-top: 8px;
  }

  .btn {
    width: 100%;
    padding: 10px 16px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }

  .btn-generate {
    background: #7c3aed;
    color: #fff;
  }
  .btn-generate:hover {
    background: #6d28d9;
  }

  .btn-cancel {
    background: #dc2626;
    color: #fff;
  }
  .btn-cancel:hover {
    background: #b91c1c;
  }
</style>

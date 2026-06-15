<script lang="ts">
  import { onMount } from "svelte";
  import PromptInput from "./PromptInput.svelte";
  import ModelSelector from "./ModelSelector.svelte";

  export let state: string;
  export let onGenerate: (params: any) => void;
  export let onCancel: () => void;

  let prompt = "";
  let negativePrompt = "";
  let modelPath = "";
  let vaePath = "";
  let samplerName = "euler_a";
  let steps = 20;
  let cfgScale = 7.0;
  let seed = -1;
  let width = 512;
  let height = 512;

  let activeTab = "txt2img";

  $: isBusy = state === "loading" || state === "generating";

  onMount(() => {
    const saved = localStorage.getItem("comfygo_form");
    if (saved) {
      try {
        const f = JSON.parse(saved);
        prompt = f.prompt ?? "";
        negativePrompt = f.negativePrompt ?? "";
        modelPath = f.modelPath ?? "";
        vaePath = f.vaePath ?? "";
        samplerName = f.samplerName ?? "euler_a";
        steps = f.steps ?? 20;
        cfgScale = f.cfgScale ?? 7.0;
        seed = f.seed ?? -1;
        width = f.width ?? 512;
        height = f.height ?? 512;
      } catch {}
    }

    window.addEventListener("beforeunload", saveForm);
    return () => window.removeEventListener("beforeunload", saveForm);
  });

  function saveForm() {
    localStorage.setItem("comfygo_form", JSON.stringify({
      prompt, negativePrompt, modelPath, vaePath, samplerName,
      steps, cfgScale, seed, width, height
    }));
  }

  $: saveForm();

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
      samplerName,
    });
  }
</script>

<div class="inspector">
  <div class="tabs">
    <button class="tab" class:active={activeTab === "txt2img"} on:click={() => (activeTab = "txt2img")}>
      Text to Image
    </button>
    <button class="tab" class:active={activeTab === "img2img"} on:click={() => (activeTab = "img2img")} disabled>
      Image to Image
    </button>
    <button class="tab" class:active={activeTab === "extras"} on:click={() => (activeTab = "extras")} disabled>
      Extras
    </button>
  </div>

  {#if activeTab === "txt2img"}
    <form class="params" on:submit|preventDefault={submit}>
      <ModelSelector bind:modelPath bind:vaePath />

      <div class="field">
        <label for="sampler">Sampler</label>
        <select id="sampler" bind:value={samplerName} disabled={isBusy}>
          <option value="euler_a">Euler A</option>
        </select>
      </div>

      <div class="field">
        <label for="steps">Steps: {steps}</label>
        <input
          id="steps"
          type="range"
          min="1"
          max="50"
          bind:value={steps}
          disabled={isBusy}
        />
      </div>

      <div class="field">
        <label for="cfg">CFG Scale: {cfgScale.toFixed(1)}</label>
        <input
          id="cfg"
          type="range"
          min="1"
          max="20"
          step="0.5"
          bind:value={cfgScale}
          disabled={isBusy}
        />
      </div>

      <div class="field">
        <label for="seed">Seed (-1 = random)</label>
        <input
          id="seed"
          type="number"
          bind:value={seed}
          disabled={isBusy}
        />
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

      <PromptInput label="Prompt" bind:value={prompt} />
      <PromptInput label="Negative Prompt" bind:value={negativePrompt} />

      <div class="actions">
        {#if isBusy}
          <button type="button" class="btn btn-cancel" on:click={onCancel}>
            Cancel
          </button>
        {:else}
          <button type="submit" class="btn btn-generate">Generate</button>
        {/if}
      </div>
    </form>
  {:else if activeTab === "img2img"}
    <div class="tab-placeholder">Image to Image coming soon</div>
  {:else}
    <div class="tab-placeholder">Extras coming soon</div>
  {/if}
</div>

<style>
  .inspector {
    grid-area: inspector;
    background: var(--bg-elevated);
    border-left: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-height: 0;
  }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }
  .tab {
    flex: 1;
    padding: 8px 4px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
    text-align: center;
    white-space: nowrap;
  }
  .tab:hover:not(:disabled) {
    color: var(--text-secondary);
  }
  .tab.active {
    color: var(--accent);
    border-bottom: 2px solid var(--accent);
  }
  .tab:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .params {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .field label {
    font-size: 11px;
    font-weight: 500;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .field input[type="range"] {
    width: 100%;
    accent-color: var(--accent);
  }
  .field input[type="number"],
  .field select {
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    color: var(--text-primary);
    padding: 5px 8px;
    border-radius: 5px;
    font-size: 13px;
  }
  .field select {
    -webkit-appearance: none;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2364748B' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 8px center;
    padding-right: 28px;
    cursor: pointer;
  }
  .field select option {
    background: var(--bg-secondary);
    color: var(--text-primary);
  }
  .field input[type="number"]:focus,
  .field select:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 2px var(--accent-glow);
  }
  .field-row {
    display: flex;
    gap: 8px;
  }
  .field-row .field {
    flex: 1;
  }

  .actions {
    margin-top: 4px;
  }
  .btn {
    width: 100%;
    padding: 9px 16px;
    border: none;
    border-radius: 7px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s;
  }
  .btn-generate {
    background: var(--btn-gradient);
    color: #fff;
  }
  .btn-generate:hover {
    opacity: 0.9;
  }
  .btn-cancel {
    background: var(--red);
    color: #fff;
  }
  .btn-cancel:hover {
    opacity: 0.9;
  }

  .tab-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: var(--text-muted);
    font-size: 13px;
  }
</style>

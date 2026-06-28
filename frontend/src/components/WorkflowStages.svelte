<script lang="ts">
  export let state: string;
  export let step = 0;
  export let total = 0;

  let pipelineStage = -1; // 0=Load, 1=Generate, 2=Save, 3=Done
  let showDone = false;

  $: {
    if (state === "loading") {
      pipelineStage = 0;
      showDone = false;
    } else if (state === "generating") {
      pipelineStage = 1;
      showDone = false;
    } else if (state === "complete") {
      pipelineStage = 2;
      showDone = false;
      // After a brief "Save" display, move to "Done"
      setTimeout(() => {
        pipelineStage = 3;
        showDone = true;
      }, 800);
    } else if (state === "error") {
      pipelineStage = -1;
      showDone = false;
    } else {
      pipelineStage = -1;
      showDone = false;
    }
  }

  const stages = [
    { id: "load", label: "Load Model", color: "#3B82F6" },
    { id: "generate", label: "Generate", color: "#7C3AED" },
    { id: "save", label: "Save", color: "#F59E0B" },
    { id: "done", label: "Done", color: "#22C55E" },
  ];
</script>

<div class="workflow">
  {#each stages as stage, i}
    <div
      class="stage"
      class:completed={i < pipelineStage || (i === 3 && showDone)}
      class:active={i === pipelineStage && !(i === 3 && !showDone)}
      class:error={state === "error"}
      style="--stage-color: {stage.color}"
    >
      <div class="stage-dot">
        {#if i === 0 && pipelineStage === 0 && state === "loading"}
          <div class="spinner-ring"></div>
        {/if}
      </div>
      <span class="stage-label">
        {#if i === 1 && state === "generating" && total > 0}
          Step {step}/{total}
        {:else if i === 2 && pipelineStage === 2}
          Encoding...
        {:else if i === 2 && showDone}
          Saved
        {:else}
          {stage.label}
        {/if}
      </span>
    </div>
    {#if i < stages.length - 1}
      <div class="arrow" class:completed={i < pipelineStage - 1 || (i === 2 && showDone)}>
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M5 12h14M13 5l7 7-7 7" />
        </svg>
      </div>
    {/if}
  {/each}
</div>

<style>
  .workflow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0;
    padding: 8px 16px;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-subtle);
    overflow-x: auto;
  }

  .stage {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 6px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    transition: all 0.2s;
    white-space: nowrap;
  }
  .stage.active {
    border-color: var(--stage-color);
    background: color-mix(in srgb, var(--stage-color) 15%, var(--bg-elevated));
  }
  .stage.completed {
    border-color: var(--stage-color);
    background: color-mix(in srgb, var(--stage-color) 10%, var(--bg-elevated));
  }
  .stage.error {
    border-color: var(--red);
    background: color-mix(in srgb, var(--red) 15%, var(--bg-elevated));
  }
  .stage.error .stage-label {
    color: var(--red);
  }

  .stage-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--stage-color);
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .stage.completed .stage-dot {
    box-shadow: 0 0 6px var(--stage-color);
  }
  .stage.active .stage-dot {
    box-shadow: 0 0 10px var(--stage-color);
    animation: pulse 1.5s infinite;
  }

  .spinner-ring {
    width: 12px;
    height: 12px;
    border: 2px solid rgba(255,255,255,0.2);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .stage-label {
    font-size: 11px;
    font-weight: 500;
    color: var(--text-secondary);
  }
  .stage.active .stage-label {
    color: var(--stage-color);
  }
  .stage.completed .stage-label {
    color: var(--stage-color);
  }

  .arrow {
    display: flex;
    align-items: center;
    padding: 0 4px;
    color: var(--text-muted);
  }
  .arrow.completed {
    color: var(--text-secondary);
  }
</style>

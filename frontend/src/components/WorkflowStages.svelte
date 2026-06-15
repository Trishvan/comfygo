<script lang="ts">
  export let state: string;
  export let progress: number;

  const stages = [
    { id: "prompt", label: "Prompt", color: "#7C3AED" },
    { id: "checkpoint", label: "Checkpoint", color: "#3B82F6" },
    { id: "lora", label: "LoRA", color: "#06B6D4" },
    { id: "sampler", label: "Sampler", color: "#F59E0B" },
    { id: "upscale", label: "Upscale", color: "#A855F7" },
    { id: "save", label: "Save", color: "#22C55E" },
  ];

  $: activeIndex = state === "generating" ? clamp(Math.floor(progress * stages.length), 0, stages.length - 1)
    : state === "complete" ? stages.length - 1
    : state === "error" ? stages.length - 1
    : -1;

  function clamp(v: number, min: number, max: number) {
    return Math.max(min, Math.min(max, v));
  }
</script>

<div class="workflow">
  {#each stages as stage, i}
    <div
      class="stage"
      class:completed={i < activeIndex}
      class:active={i === activeIndex}
      style="--stage-color: {stage.color}"
    >
      <div class="stage-dot"></div>
      <span class="stage-label">{stage.label}</span>
    </div>
    {#if i < stages.length - 1}
      <div class="arrow" class:completed={i < activeIndex}>
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

  .stage-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--stage-color);
    transition: all 0.2s;
  }
  .stage.completed .stage-dot {
    box-shadow: 0 0 6px var(--stage-color);
  }
  .stage.active .stage-dot {
    box-shadow: 0 0 10px var(--stage-color);
    animation: pulse 1.5s infinite;
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

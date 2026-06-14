<script lang="ts">
  export let state: string;
  export let progress: number;

  $: label = state === "idle" ? "Ready"
    : state === "loading" ? "Loading model..."
    : state === "generating" ? `Generating... ${(progress * 100).toFixed(0)}%`
    : state === "complete" ? "Complete"
    : "Error";

  $: barWidth = state === "generating" ? `${progress * 100}%`
    : state === "complete" ? "100%"
    : "0%";
</script>

<div class="status-bar">
  <span class="status-label">{label}</span>
  <div class="progress-track">
    <div class="progress-fill" style="width: {barWidth}"></div>
  </div>
</div>

<style>
  .status-bar {
    width: 100%;
    max-width: 600px;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 0;
  }

  .status-label {
    font-size: 12px;
    color: #888;
    white-space: nowrap;
    min-width: 100px;
  }

  .progress-track {
    flex: 1;
    height: 4px;
    background: #2a2a30;
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: #a78bfa;
    border-radius: 2px;
    transition: width 0.15s ease;
  }
</style>

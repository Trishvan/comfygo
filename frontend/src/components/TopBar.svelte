<script lang="ts">
  export let state: string;
  export let ramUsedGB = 0;
  export let ramTotalGB = 0;
  export let vramUsedGB = 0;
  export let vramTotalGB = 0;

  $: ramStr = `${ramUsedGB.toFixed(1)} / ${ramTotalGB.toFixed(1)} GB`;
  $: vramStr = vramTotalGB > 0 ? `${vramUsedGB.toFixed(1)} / ${vramTotalGB.toFixed(1)} GB` : "-- / -- GB";
</script>

<div class="topbar">
  <div class="topbar-left">
    <span class="app-name">ComfyGo</span>
    <span class="separator">|</span>
    <span class="project-name">Untitled</span>
  </div>
  <div class="topbar-right">
    <div class="stat">
      <span class="stat-label">RAM</span>
      <span class="stat-value">{ramStr}</span>
    </div>
    <div class="stat">
      <span class="stat-label">VRAM</span>
      <span class="stat-value">{vramStr}</span>
    </div>
    <div class="stat backend">
      <span
        class="status-dot"
        class:green={state === "idle" || state === "complete"}
        class:amber={state === "loading"}
        class:red={state === "error"}
      ></span>
      <span class="stat-label">Backend</span>
      <span class="stat-value">stable-diffusion.cpp</span>
    </div>
  </div>
</div>

<style>
  .topbar {
    grid-area: topbar;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-subtle);
    z-index: 10;
    height: 38px;
  }
  .topbar-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .app-name {
    font-weight: 700;
    font-size: 13px;
    color: var(--accent);
  }
  .separator {
    color: var(--text-muted);
  }
  .project-name {
    font-size: 13px;
    color: var(--text-secondary);
  }
  .topbar-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .stat {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
  }
  .stat-label {
    color: var(--text-muted);
  }
  .stat-value {
    color: var(--text-secondary);
    font-weight: 500;
  }
  .backend {
    gap: 6px;
  }
  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    transition: background 0.2s;
  }
  .green {
    background: var(--green);
  }
  .amber {
    background: var(--amber);
  }
  .red {
    background: var(--red);
  }
</style>

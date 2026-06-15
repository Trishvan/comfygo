<script lang="ts">
  export let state: string;
  export let ramUsedGB = 0;
  export let ramTotalGB = 0;
  export let ramPercent = 0;
  export let vramUsedGB = 0;
  export let vramTotalGB = 0;
  export let vramPercent = 0;

  let activeTab = "queue";

  interface QueueItem {
    id: number;
    prompt: string;
    status: "running" | "queued" | "completed" | "failed";
    progress: number;
    model: string;
    timestamp: Date;
  }

  let queueItems: QueueItem[] = [];
  let jobCounter = 0;

  export function addJob(prompt: string, model: string) {
    jobCounter++;
    const item: QueueItem = {
      id: jobCounter,
      prompt: prompt.slice(0, 40) + (prompt.length > 40 ? "..." : ""),
      status: "queued",
      progress: 0,
      model,
      timestamp: new Date(),
    };
    queueItems = [item, ...queueItems];
    if (queueItems.length > 20) queueItems = queueItems.slice(0, 20);
  }

  export function updateRunningJob(progress: number) {
    const running = queueItems.find((q) => q.status === "running");
    if (running) {
      running.progress = progress;
      queueItems = [...queueItems];
    }
  }

  export function completeRunningJob(success: boolean) {
    const running = queueItems.find((q) => q.status === "running");
    if (running) {
      running.status = success ? "completed" : "failed";
      running.progress = success ? 1 : 0;
      queueItems = [...queueItems];
    }
  }

  export function setRunningJob() {
    const queued = queueItems.find((q) => q.status === "queued");
    if (queued) {
      queued.status = "running";
      queueItems = [...queueItems];
    }
  }

  $: timeStr = (d: Date) =>
    d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

  let logs: string[] = [];
  let logInput = "";

  export function addLog(msg: string) {
    logs = [`[${new Date().toLocaleTimeString()}] ${msg}`, ...logs];
    if (logs.length > 100) logs = logs.slice(0, 100);
  }
</script>

<div class="bottom-panel">
  <div class="tabs">
    <button
      class="tab"
      class:active={activeTab === "queue"}
      on:click={() => (activeTab = "queue")}
    >
      Queue
    </button>
    <button
      class="tab"
      class:active={activeTab === "logs"}
      on:click={() => (activeTab = "logs")}
    >
      Logs
    </button>
    <button
      class="tab"
      class:active={activeTab === "system"}
      on:click={() => (activeTab = "system")}
    >
      System Monitor
    </button>
  </div>

  <div class="tab-content">
    {#if activeTab === "queue"}
      <div class="queue-filters">
        <span class="filter active">All</span>
        <span class="filter">Running</span>
        <span class="filter">Queued</span>
        <span class="filter">Completed</span>
      </div>
      <div class="queue-list">
        {#if queueItems.length === 0}
          <div class="empty">No jobs in queue</div>
        {:else}
          {#each queueItems as item (item.id)}
            <div class="queue-item" class:running={item.status === "running"}>
              <div class="qi-header">
                <span
                  class="qi-status"
                  class:running={item.status === "running"}
                  class:completed={item.status === "completed"}
                  class:failed={item.status === "failed"}
                ></span>
                <span class="qi-prompt">{item.prompt}</span>
                <span class="qi-time">{timeStr(item.timestamp)}</span>
              </div>
              {#if item.status === "running" || item.status === "completed"}
                <div class="qi-progress">
                  <div
                    class="qi-bar"
                    style="width: {item.progress * 100}%"
                    class:done={item.status === "completed"}
                  ></div>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
    {:else if activeTab === "logs"}
      <div class="log-list">
        {#if logs.length === 0}
          <div class="empty">No logs</div>
        {:else}
          {#each logs as log}
            <div class="log-line">{log}</div>
          {/each}
        {/if}
      </div>
    {:else}
      <div class="system-monitor">
        <div class="sys-row">
          <span class="sys-label">RAM</span>
          <div class="sys-bar-track">
            <div class="sys-bar" style="width: {ramPercent}%"></div>
          </div>
          <span class="sys-value">{ramUsedGB.toFixed(1)} / {ramTotalGB.toFixed(1)} GB</span>
        </div>
        <div class="sys-row">
          <span class="sys-label">VRAM</span>
          <div class="sys-bar-track">
            <div class="sys-bar" style="width: {vramPercent}%"></div>
          </div>
          <span class="sys-value">{vramTotalGB > 0 ? `${vramUsedGB.toFixed(1)} / ${vramTotalGB.toFixed(1)} GB` : "-- / -- GB"}</span>
        </div>
        <div class="sys-row">
          <span class="sys-label">Backend</span>
          <span class="sys-status" class:green={state === "idle" || state === "complete"} class:amber={state === "loading"} class:red={state === "error"}>
            {state === "idle" ? "Ready" : state === "loading" ? "Loading" : state === "generating" ? "Generating" : state === "complete" ? "Ready" : "Error"}
          </span>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .bottom-panel {
    grid-area: bottom;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-subtle);
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
    padding: 6px 16px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }
  .tab:hover {
    color: var(--text-secondary);
  }
  .tab.active {
    color: var(--accent);
    border-bottom: 2px solid var(--accent);
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: 8px 12px;
    min-height: 0;
  }

  .queue-filters {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
  }
  .filter {
    font-size: 11px;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px 8px;
    border-radius: 4px;
  }
  .filter:hover {
    color: var(--text-secondary);
  }
  .filter.active {
    color: var(--accent);
    background: var(--accent-glow);
  }

  .queue-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .queue-item {
    padding: 6px 8px;
    border-radius: 5px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
  }
  .queue-item.running {
    border-color: var(--accent);
  }

  .qi-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .qi-status {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
    background: var(--text-muted);
  }
  .qi-status.running {
    background: var(--accent);
    animation: pulse 1.5s infinite;
  }
  .qi-status.completed {
    background: var(--green);
  }
  .qi-status.failed {
    background: var(--red);
  }
  .qi-prompt {
    font-size: 12px;
    color: var(--text-secondary);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .qi-time {
    font-size: 10px;
    color: var(--text-muted);
    flex-shrink: 0;
  }

  .qi-progress {
    margin-top: 4px;
    height: 3px;
    background: var(--border-subtle);
    border-radius: 2px;
    overflow: hidden;
  }
  .qi-bar {
    height: 100%;
    background: var(--accent);
    border-radius: 2px;
    transition: width 0.2s;
  }
  .qi-bar.done {
    background: var(--green);
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }

  .empty {
    color: var(--text-muted);
    font-size: 12px;
    padding: 12px;
    text-align: center;
  }

  .log-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .log-line {
    font-size: 11px;
    font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
    color: var(--text-secondary);
    padding: 2px 0;
  }

  .system-monitor {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 4px 0;
  }
  .sys-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .sys-label {
    font-size: 11px;
    color: var(--text-muted);
    width: 50px;
    flex-shrink: 0;
  }
  .sys-bar-track {
    flex: 1;
    height: 6px;
    background: var(--border-subtle);
    border-radius: 3px;
    overflow: hidden;
  }
  .sys-bar {
    height: 100%;
    background: var(--accent);
    border-radius: 3px;
    transition: width 0.3s;
  }
  .sys-value {
    font-size: 11px;
    color: var(--text-secondary);
    width: 70px;
    text-align: right;
    flex-shrink: 0;
  }
  .sys-status {
    font-size: 11px;
    font-weight: 500;
  }
  .sys-status.green {
    color: var(--green);
  }
  .sys-status.amber {
    color: var(--amber);
  }
  .sys-status.red {
    color: var(--red);
  }
</style>

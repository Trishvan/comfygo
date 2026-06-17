<script lang="ts">
  import { CancelJob, CancelRunningJob, ClearCompleted, ReorderQueue, RetryJob } from "../../wailsjs/go/orchestrator/Manager";

  export let state: string;
  export let queue: any[] = [];
  export let ramUsedGB = 0;
  export let ramTotalGB = 0;
  export let ramPercent = 0;
  export let vramUsedGB = 0;
  export let vramTotalGB = 0;
  export let vramPercent = 0;

  let activeTab = "queue";
  let filter: string | null = null;

  let logs: string[] = [];

  export function addLog(msg: string) {
    logs = [`[${new Date().toLocaleTimeString()}] ${msg}`, ...logs];
    if (logs.length > 100) logs = logs.slice(0, 100);
  }

  async function handleCancelJob(id: number, status: string) {
    if (status === "running") {
      await CancelRunningJob();
    } else {
      await CancelJob(id);
    }
  }

  async function handleClearCompleted() {
    await ClearCompleted();
  }

  async function handleReorder(from: number, to: number) {
    await ReorderQueue(from, to);
  }

  async function handleRetryJob(id: number) {
    await RetryJob(id);
  }

  function statusLabel(s: string): string {
    switch (s) {
      case "queued": return "Queued";
      case "running": return "Running";
      case "completed": return "Done";
      case "failed": return "Failed";
      case "cancelled": return "Cancelled";
      default: return s;
    }
  }

  function promptShort(prompt: string): string {
    return prompt.slice(0, 40) + (prompt.length > 40 ? "..." : "");
  }

  function fileName(filePath: string): string {
    const parts = (filePath || "").split("/");
    return parts[parts.length - 1] || "";
  }

  $: filteredQueue = !Array.isArray(queue) ? [] : (filter ? queue.filter((item: any) => item.status === filter) : queue);

  function timeStr(d: string) {
    if (!d) return "--";
    const dt = new Date(d);
    if (isNaN(dt.getTime())) return "--";
    return dt.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  }
</script>

<div class="bottom-panel">
  <div class="tabs">
    <button
      class="tab"
      class:active={activeTab === "queue"}
      on:click={() => (activeTab = "queue")}
    >
      Queue ({queue.length})
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
      <div class="queue-toolbar">
        <div class="queue-filters">
          <span class="filter" class:active={filter === null} on:click={() => (filter = null)}>All</span>
          <span class="filter" class:active={filter === "running"} on:click={() => (filter = "running")}>Running</span>
          <span class="filter" class:active={filter === "queued"} on:click={() => (filter = "queued")}>Queued</span>
          <span class="filter" class:active={filter === "completed"} on:click={() => (filter = "completed")}>Done</span>
        </div>
        <button class="clear-btn" on:click={handleClearCompleted}>Clear Done</button>
      </div>
      <div class="queue-list">
        {#if filteredQueue.length === 0}
          <div class="empty">No {filter || "jobs"} in queue</div>
        {:else}
          {#each filteredQueue as item, i (item.id)}
            <div
              class="queue-item"
              class:running={item.status === "running"}
              class:failed={item.status === "failed" || item.status === "cancelled"}
              class:done={item.status === "completed"}
            >
              <div class="qi-header">
                <span class="qi-status" class:running={item.status === "running"} class:done={item.status === "completed"} class:failed={item.status === "failed" || item.status === "cancelled"}></span>
                <span class="qi-prompt" title={item.params?.prompt}>{promptShort(item.params?.prompt || "Untitled")}</span>
                <span class="qi-model">{fileName(item.params?.modelPath)}</span>
                <span class="qi-time">{timeStr(item.createdAt)}</span>
                <span class="qi-state">{statusLabel(item.status)}</span>
              </div>
              {#if item.status === "running" || item.status === "queued"}
                <div class="qi-progress">
                  <div class="qi-bar" style="width: {item.progress * 100}%"></div>
                </div>
              {:else if item.status === "completed"}
                <div class="qi-progress">
                  <div class="qi-bar done" style="width: 100%"></div>
                </div>
              {/if}
              {#if item.error}
                <div class="qi-error">{item.error}</div>
              {/if}
              <div class="qi-actions">
                {#if item.status === "queued"}
                  <button class="qi-btn" on:click={() => handleCancelJob(item.id, item.status)}>Cancel</button>
                  {#if i > 0}
                    <button class="qi-btn" on:click={() => handleReorder(i, i - 1)}>Up</button>
                  {/if}
                  {#if i < filteredQueue.length - 1}
                    <button class="qi-btn" on:click={() => handleReorder(i, i + 1)}>Down</button>
                  {/if}
                {:else if item.status === "running"}
                  <button class="qi-btn cancel" on:click={() => handleCancelJob(item.id, item.status)}>Cancel</button>
                {:else}
                  <button class="qi-btn rerun" on:click={() => handleRetryJob(item.id)}>Re-run</button>
                {/if}
              </div>
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

  .queue-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }
  .queue-filters {
    display: flex;
    gap: 8px;
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
  .clear-btn {
    background: none;
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 4px;
    cursor: pointer;
  }
  .clear-btn:hover {
    color: var(--text-secondary);
    border-color: var(--text-muted);
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
  .queue-item.failed {
    border-color: var(--red);
  }
  .queue-item.done {
    border-color: var(--green);
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
  .qi-status.done {
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
  .qi-model {
    font-size: 10px;
    color: var(--text-muted);
    flex-shrink: 0;
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .qi-time {
    font-size: 10px;
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .qi-state {
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
  .qi-error {
    font-size: 10px;
    color: var(--red);
    margin-top: 4px;
  }

  .qi-actions {
    display: flex;
    gap: 4px;
    margin-top: 4px;
  }
  .qi-btn {
    background: none;
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 3px;
    cursor: pointer;
  }
  .qi-btn:hover {
    color: var(--text-secondary);
  }
  .qi-btn.cancel {
    color: var(--red);
  }
  .qi-btn.cancel:hover {
    border-color: var(--red);
  }
  .qi-btn.rerun {
    color: var(--green);
  }
  .qi-btn.rerun:hover {
    border-color: var(--green);
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

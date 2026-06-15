<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { EventsOn, WindowSetSize, WindowSetPosition } from "../wailsjs/runtime/runtime";
  import { Generate, Cancel, GetImageData, GetSystemStats } from "../wailsjs/go/orchestrator/Manager";
  import TopBar from "./components/TopBar.svelte";
  import LeftNav from "./components/LeftNav.svelte";
  import PreviewPanel from "./components/PreviewPanel.svelte";
  import InspectorPanel from "./components/InspectorPanel.svelte";
  import WorkflowStages from "./components/WorkflowStages.svelte";
  import BottomPanel from "./components/BottomPanel.svelte";

  let state = "idle";
  let progress = 0;
  let imageUrl = "";
  let imageWidth = 0;
  let imageHeight = 0;

  let ramUsedGB = 0;
  let ramTotalGB = 0;
  let ramPercent = 0;
  let vramUsedGB = 0;
  let vramTotalGB = 0;
  let vramPercent = 0;

  let previewPanel: PreviewPanel;
  let bottomPanel: BottomPanel;

  let statsTimer: number | undefined;

  async function pollStats() {
    try {
      const s = await GetSystemStats();
      ramUsedGB = s.ramUsedGB;
      ramTotalGB = s.ramTotalGB;
      ramPercent = s.ramPercent;
      vramUsedGB = s.vramUsedGB;
      vramTotalGB = s.vramTotalGB;
      vramPercent = s.vramPercent;
    } catch {}
  }

  async function handleGenerate(params: any) {
    imageUrl = "";
    bottomPanel?.addJob(params.prompt || "Untitled", params.modelPath || "unknown");
    bottomPanel?.addLog("Generation started");
    bottomPanel?.setRunningJob();
    Generate(params);
  }

  function handleCancel() {
    Cancel();
    bottomPanel?.addLog("Generation cancelled");
  }

  function onStateChange(s: string) {
    state = s;
    if (s === "generating") {
      bottomPanel?.setRunningJob();
    }
  }

  function onProgress(step: number, total: number) {
    progress = step / total;
    bottomPanel?.updateRunningJob(step / total);
  }

  async function onComplete(meta: { width: number; height: number }) {
    imageWidth = meta.width;
    imageHeight = meta.height;
    state = "complete";

    const b64 = await GetImageData();
    if (b64) {
      imageUrl = `data:image/png;base64,${b64}`;
      previewPanel?.addThumbnail(imageUrl);
    }
    bottomPanel?.completeRunningJob(true);
    bottomPanel?.addLog("Generation completed");
  }

  function onError(msg: string) {
    state = "error";
    console.error(msg);
    bottomPanel?.completeRunningJob(false);
    bottomPanel?.addLog(`Error: ${msg}`);
  }

  onMount(() => {
    const unsub1 = EventsOn("state-change", onStateChange);
    const unsub2 = EventsOn("progress", onProgress);
    const unsub3 = EventsOn("generation-complete", onComplete);
    const unsub4 = EventsOn("error", onError);

    // Position at (0,0) and size to fill the available screen area.
    // More reliable than WindowMaximise on Linux with webkit2_41.
    setTimeout(async () => {
      try {
        await WindowSetPosition(0, 0);
        await WindowSetSize(screen.availWidth, screen.availHeight);
      } catch {}
    }, 200);

    // Poll system stats every 3 seconds
    pollStats();
    statsTimer = window.setInterval(pollStats, 3000);

    return () => {
      unsub1(); unsub2(); unsub3(); unsub4();
      if (statsTimer !== undefined) clearInterval(statsTimer);
    };
  });
</script>

<div class="app-shell">
  <LeftNav />
  <TopBar {state} {ramUsedGB} {ramTotalGB} {vramUsedGB} {vramTotalGB} />
  <div class="main-area">
    <PreviewPanel
      {imageUrl}
      {imageWidth}
      {imageHeight}
      {progress}
      {state}
      bind:this={previewPanel}
    />
    <WorkflowStages {state} {progress} />
  </div>
  <InspectorPanel {state} onGenerate={handleGenerate} onCancel={handleCancel} />
  <BottomPanel {state} {ramUsedGB} {ramTotalGB} {ramPercent} {vramUsedGB} {vramTotalGB} {vramPercent} bind:this={bottomPanel} />
</div>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(:root) {
    --bg-primary: #080B14;
    --bg-secondary: #0F172A;
    --bg-elevated: #131C31;
    --accent: #7C3AED;
    --accent-hover: #8B5CF6;
    --accent-glow: rgba(124,58,237,0.25);
    --blue: #3B82F6;
    --green: #22C55E;
    --amber: #F59E0B;
    --red: #EF4444;
    --text-primary: #F8FAFC;
    --text-secondary: #94A3B8;
    --text-muted: #64748B;
    --border-subtle: rgba(255,255,255,0.08);
    --btn-gradient: linear-gradient(135deg, #7C3AED 0%, #3B82F6 100%);
  }

  :global(html),
  :global(body) {
    width: 100%;
    height: 100%;
    margin: 0;
    padding: 0;
  }
  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .app-shell {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    overflow: hidden;
    display: grid;
    grid-template-columns: 48px 1fr 320px;
    grid-template-rows: 38px 1fr 180px;
    grid-template-areas:
      "nav topbar topbar"
      "nav main inspector"
      "nav bottom bottom";
  }

  .main-area {
    grid-area: main;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-height: 0;
  }
</style>

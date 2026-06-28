<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { EventsOn, WindowSetSize, WindowSetPosition } from "../wailsjs/runtime/runtime";
  import {
    EnqueueJob,
    CancelRunningJob,
    ClearCompleted,
    GetSystemStats,
  } from "../wailsjs/go/orchestrator/Manager";
  import TopBar from "./components/TopBar.svelte";
  import LeftNav from "./components/LeftNav.svelte";
  import PreviewPanel from "./components/PreviewPanel.svelte";
  import InspectorPanel from "./components/InspectorPanel.svelte";
  import WorkflowStages from "./components/WorkflowStages.svelte";
  import BottomPanel from "./components/BottomPanel.svelte";
  import GalleryView from "./components/GalleryView.svelte";

  let state = "idle";
  let progress = 0;
  let step = 0;
  let total = 0;
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
  let queue: any[] = [];

  // View switching: nav controls which view is shown in main area
  let activeNav = "home";
  let currentView = "preview"; // "preview" | "gallery"

  // History viewing — separate from current generation
  let viewingUrl = "";
  let viewingWidth = 0;
  let viewingHeight = 0;

  // Toast notification
  let toastMessage = "";
  let toastVisible = false;
  let toastTimer: number | undefined;

  function showToast(msg: string) {
    toastMessage = msg;
    toastVisible = true;
    if (toastTimer !== undefined) clearTimeout(toastTimer);
    toastTimer = window.setTimeout(() => {
      toastVisible = false;
    }, 4000);
  }

  function onNavChange(nav: string) {
    activeNav = nav;
    if (nav === "outputs") {
      currentView = "gallery";
    } else {
      currentView = "preview";
    }
  }

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
    viewingUrl = "";
    try {
      await EnqueueJob(params);
      bottomPanel?.addLog("Queued generation");
    } catch (e) {
      console.error("EnqueueJob failed:", e);
      bottomPanel?.addLog(`Error: ${e}`);
    }
  }

  function handleCancel() {
    CancelRunningJob();
  }

  function onStateChange(s: string) {
    state = s;
  }

  function onProgress(pStep: number, pTotal: number) {
    step = pStep;
    total = pTotal;
    progress = pTotal > 0 ? pStep / pTotal : 0;
  }

  async function onComplete(meta: { width: number; height: number }) {
    imageWidth = meta.width;
    imageHeight = meta.height;
    state = "complete";
    step = total;
    progress = 1;

    const { GetImageData } = await import("../wailsjs/go/orchestrator/Manager");
    const b64 = await GetImageData();
    if (b64) {
      const url = `data:image/png;base64,${b64}`;
      imageUrl = url;
      // Only auto-show if not viewing a past output
      if (!viewingUrl) {
        previewPanel?.showImage(url, meta.width, meta.height);
      }
      previewPanel?.addThumbnail(url);
    }
    bottomPanel?.addLog("Generation completed");

    // Toast if user is in gallery view
    if (currentView === "gallery") {
      showToast("Generation complete! Check the gallery.");
    }
  }

  function onError(msg: string) {
    state = "error";
    console.error(msg);
    bottomPanel?.addLog(`Error: ${msg}`);
  }

  function onQueueUpdate(items: any) {
    if (Array.isArray(items)) {
      queue = items;
    }
  }

  // Called from GalleryView when user wants to view a past output in preview
  function onSelectFromGallery(url: string, w: number, h: number) {
    viewingUrl = url;
    viewingWidth = w;
    viewingHeight = h;
    currentView = "preview";
    activeNav = "home";
    previewPanel?.showImage(url, w, h);
  }

  function onBackToCurrent() {
    viewingUrl = "";
    if (imageUrl) {
      previewPanel?.showImage(imageUrl, imageWidth, imageHeight);
    }
  }

  onMount(async () => {
    const unsub1 = EventsOn("state-change", onStateChange);
    const unsub2 = EventsOn("progress", onProgress);
    const unsub3 = EventsOn("generation-complete", onComplete);
    const unsub4 = EventsOn("error", onError);
    const unsub5 = EventsOn("queue-update", onQueueUpdate);

    // Fetch initial queue state
    try {
      const { GetQueue } = await import("../wailsjs/go/orchestrator/Manager");
      const items = await GetQueue();
      if (Array.isArray(items)) {
        queue = items;
      }
    } catch {}

    setTimeout(async () => {
      try {
        await WindowSetPosition(0, 0);
        await WindowSetSize(screen.availWidth, screen.availHeight);
      } catch {}
    }, 200);

    pollStats();
    statsTimer = window.setInterval(pollStats, 3000);

    return () => {
      unsub1(); unsub2(); unsub3(); unsub4(); unsub5();
      if (statsTimer !== undefined) clearInterval(statsTimer);
      if (toastTimer !== undefined) clearTimeout(toastTimer);
    };
  });
</script>

<div class="app-shell">
  <LeftNav {activeNav} onNavChange={onNavChange} />
  <TopBar {state} {ramUsedGB} {ramTotalGB} {vramUsedGB} {vramTotalGB} />
  <div class="main-area">
    {#if currentView === "gallery"}
      <GalleryView
        onSelectImage={onSelectFromGallery}
        onBack={() => { currentView = "preview"; activeNav = "home"; }}
        generating={state === "loading" || state === "generating"}
      />
    {:else}
      <PreviewPanel
        {imageUrl}
        {imageWidth}
        {imageHeight}
        {progress}
        {state}
        {viewingUrl}
        {viewingWidth}
        {viewingHeight}
        onBackToCurrent={onBackToCurrent}
        bind:this={previewPanel}
      />
    {/if}
    <WorkflowStages {state} {step} {total} />
  </div>
  <InspectorPanel {state} onGenerate={handleGenerate} onCancel={handleCancel} />
  <BottomPanel {state} {queue} {ramUsedGB} {ramTotalGB} {ramPercent} {vramUsedGB} {vramTotalGB} {vramPercent} bind:this={bottomPanel} />
</div>

{#if toastVisible}
  <div class="toast">
    <span>{toastMessage}</span>
  </div>
{/if}

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

  .toast {
    position: fixed;
    bottom: 190px;
    right: 330px;
    padding: 10px 18px;
    background: var(--bg-elevated);
    border: 1px solid var(--green);
    border-radius: 8px;
    color: var(--green);
    font-size: 13px;
    font-weight: 500;
    z-index: 1000;
    box-shadow: 0 4px 20px rgba(0,0,0,0.4);
    animation: toastIn 0.3s ease;
  }
  @keyframes toastIn {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>

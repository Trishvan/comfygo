<script lang="ts">
  import { onMount } from "svelte";
  import { EventsOn } from "../wailsjs/runtime/runtime";
  import { Generate, Cancel, GetImageData } from "../wailsjs/go/orchestrator/Manager";
  import GenerationForm from "./components/GenerationForm.svelte";
  import PreviewPanel from "./components/PreviewPanel.svelte";
  import StatusBar from "./components/StatusBar.svelte";

  let state = "idle";
  let progress = 0;
  let imageUrl = "";
  let imageWidth = 0;
  let imageHeight = 0;

  async function handleGenerate(params: any) {
    imageUrl = "";
    Generate(params);
  }

  function handleCancel() {
    Cancel();
  }

  function onStateChange(s: string) {
    state = s;
  }

  function onProgress(step: number, total: number) {
    progress = step / total;
  }

  async function onComplete(meta: { width: number; height: number }) {
    imageWidth = meta.width;
    imageHeight = meta.height;
    state = "complete";

    const b64 = await GetImageData();
    if (b64) {
      imageUrl = `data:image/png;base64,${b64}`;
    }
  }

  function onError(msg: string) {
    state = "error";
    console.error(msg);
  }

  onMount(() => {
    const unsub1 = EventsOn("state-change", onStateChange);
    const unsub2 = EventsOn("progress", onProgress);
    const unsub3 = EventsOn("generation-complete", onComplete);
    const unsub4 = EventsOn("error", onError);
    return () => { unsub1(); unsub2(); unsub3(); unsub4(); };
  });
</script>

<div class="app-shell">
  <aside class="sidebar">
    <h1 class="logo">ComfyGo</h1>
    <GenerationForm {state} onGenerate={handleGenerate} onCancel={handleCancel} />
  </aside>
  <main class="main-area">
    <PreviewPanel {imageUrl} {imageWidth} {imageHeight} {progress} {state} />
    <StatusBar {state} {progress} />
  </main>
</div>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: #0f0f11;
    color: #e0e0e0;
    overflow: hidden;
    height: 100vh;
  }

  .app-shell {
    display: flex;
    height: 100vh;
  }

  .sidebar {
    width: 320px;
    min-width: 320px;
    background: #1a1a1e;
    border-right: 1px solid #2a2a30;
    display: flex;
    flex-direction: column;
    padding: 20px;
    overflow-y: auto;
  }

  .logo {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.5px;
    color: #a78bfa;
    margin-bottom: 24px;
  }

  .main-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px;
    position: relative;
  }
</style>

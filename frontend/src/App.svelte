<script lang="ts">
  import GenerationForm from "./components/GenerationForm.svelte";
  import PreviewPanel from "./components/PreviewPanel.svelte";
  import StatusBar from "./components/StatusBar.svelte";

  let state = "idle";
  let progress = 0;
  let imageUrl = "";
  let imageWidth = 0;
  let imageHeight = 0;
  let genId = 0;

  function handleGenerate(params: any) {
    imageUrl = "";
    window.runtime.Call("Generate", params);
  }

  function handleCancel() {
    window.runtime.Call("Cancel");
  }

  function onStateChange(s: string) {
    state = s;
  }

  function onProgress(step: number, total: number) {
    progress = step / total;
  }

  function onComplete(meta: { width: number; height: number }) {
    genId++;
    imageUrl = `wails://app/render.png?id=${genId}`;
    imageWidth = meta.width;
    imageHeight = meta.height;
    state = "complete";
  }

  function onError(msg: string) {
    state = "error";
    console.error(msg);
  }

  window.runtime.EventsOn("state-change", onStateChange);
  window.runtime.EventsOn("progress", onProgress);
  window.runtime.EventsOn("generation-complete", onComplete);
  window.runtime.EventsOn("error", onError);
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

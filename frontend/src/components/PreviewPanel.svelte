<script lang="ts">
  export let imageUrl: string;
  export let imageWidth: number;
  export let imageHeight: number;
  export let progress: number;
  export let state: string;
  export let viewingUrl = "";
  export let viewingWidth = 0;
  export let viewingHeight = 0;
  export let onBackToCurrent: () => void;

  let zoomMode: "fit" | "fill" | "one" = "fit";
  let thumbnails: string[] = [];
  let currentDisplayUrl = "";
  let currentDisplayW = 0;
  let currentDisplayH = 0;

  export function showImage(url: string, w: number, h: number) {
    currentDisplayUrl = url;
    currentDisplayW = w;
    currentDisplayH = h;
  }

  export function addThumbnail(url: string) {
    thumbnails = [url, ...thumbnails];
    if (thumbnails.length > 10) thumbnails = thumbnails.slice(0, 10);
  }

  $: displayUrl = viewingUrl || currentDisplayUrl;
  $: displayW = viewingWidth || currentDisplayW;
  $: displayH = viewingHeight || currentDisplayH;
  $: isShowingHistory = !!viewingUrl;

  $: progressLabel = state === "generating"
    ? `Generating... ${(progress * 100).toFixed(0)}%`
    : state === "loading"
    ? "Loading model..."
    : state === "complete"
    ? "Complete"
    : state === "error"
    ? "Error"
    : "";

  $: resolutionLabel = displayW > 0 ? `${displayW} x ${displayH}` : "";

  function cycleZoom() {
    const modes: Array<"fit" | "fill" | "one"> = ["fit", "fill", "one"];
    const idx = modes.indexOf(zoomMode);
    zoomMode = modes[(idx + 1) % modes.length];
  }
</script>

<div class="preview-area">
  {#if isShowingHistory}
    <div class="history-banner">
      <span>Viewing saved output</span>
      {#if state === "loading" || state === "generating"}
        <span class="gen-active">⚡ Generating in background</span>
      {/if}
      <button class="back-btn" on:click={onBackToCurrent}>
        Back to Current
      </button>
    </div>
  {/if}

  <div class="preview-canvas">
    {#if state === "generating" && !isShowingHistory}
      <div class="placeholder">
        <div class="spinner"></div>
        <p>{progressLabel}</p>
      </div>
    {:else if displayUrl}
      <img
        src={displayUrl}
        alt="Generated image"
        class="preview-img"
        class:fit={zoomMode === "fit"}
        class:fill={zoomMode === "fill"}
        class:one={zoomMode === "one"}
      />
    {:else if state === "error"}
      <div class="placeholder error">
        <p>Generation failed</p>
      </div>
    {:else}
      <div class="placeholder">
        <p>Configure parameters and press Generate</p>
      </div>
    {/if}
  </div>

  {#if state === "generating" && !isShowingHistory}
    <div class="progress-row">
      <div class="progress-track">
        <div class="progress-fill" style="width: {progress * 100}%"></div>
      </div>
      <span class="progress-label">{progressLabel}</span>
    </div>
  {/if}

  <div class="preview-tools">
    <div class="tools-left">
      {#if resolutionLabel}
        <span class="resolution">{resolutionLabel}</span>
      {/if}
      {#if isShowingHistory}
        <span class="history-tag">History</span>
      {/if}
    </div>
    <div class="tools-right">
      <button class="tool-btn" title="Fit" class:active={zoomMode === "fit"} on:click={() => (zoomMode = "fit")}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M8 3H5a2 2 0 00-2 2v3m18 0V5a2 2 0 00-2-2h-3m0 18h3a2 2 0 002-2v-3M3 16v3a2 2 0 002 2h3" />
        </svg>
      </button>
      <button class="tool-btn" title="Fill" class:active={zoomMode === "fill"} on:click={() => (zoomMode = "fill")}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 20h16M4 4h16" />
        </svg>
      </button>
      <button class="tool-btn" title="1:1" class:active={zoomMode === "one"} on:click={() => (zoomMode = "one")}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2v20M2 12h20" />
        </svg>
      </button>
    </div>
  </div>

  {#if thumbnails.length > 0 && !isShowingHistory}
    <div class="thumbnails">
      {#each thumbnails as thumb, i}
        <button
          class="thumb-btn"
          class:active={i === 0 && state === "complete"}
          on:click={() => showImage(thumb, displayW, displayH)}
        >
          <img src={thumb} alt="Thumbnail {i}" />
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .preview-area {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .history-banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 12px;
    background: color-mix(in srgb, var(--blue) 15%, var(--bg-elevated));
    border-bottom: 1px solid var(--border-subtle);
    font-size: 11px;
    color: var(--text-secondary);
    flex-shrink: 0;
  }
  .gen-active {
    color: var(--accent);
    font-weight: 500;
  }
  .back-btn {
    margin-left: auto;
    padding: 3px 10px;
    border: 1px solid var(--border-subtle);
    border-radius: 4px;
    background: var(--bg-elevated);
    color: var(--text-secondary);
    font-size: 10px;
    cursor: pointer;
    transition: all 0.15s;
  }
  .back-btn:hover {
    color: var(--text-primary);
    border-color: var(--text-muted);
  }

  .preview-canvas {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    position: relative;
    background: var(--bg-primary);
    border-radius: 0;
    min-height: 0;
  }

  .placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    color: var(--text-muted);
    font-size: 14px;
  }
  .placeholder.error {
    color: var(--red);
  }

  .preview-img {
    border-radius: 0;
    transition: all 0.2s;
  }
  .preview-img.fit {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
  }
  .preview-img.fill {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .preview-img.one {
    image-rendering: pixelated;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--border-subtle);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .progress-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 16px;
    background: var(--bg-elevated);
    border-top: 1px solid var(--border-subtle);
  }
  .progress-track {
    flex: 1;
    height: 4px;
    background: var(--border-subtle);
    border-radius: 2px;
    overflow: hidden;
  }
  .progress-fill {
    height: 100%;
    background: var(--btn-gradient);
    border-radius: 2px;
    transition: width 0.15s ease;
  }
  .progress-label {
    font-size: 11px;
    color: var(--text-secondary);
    white-space: nowrap;
  }

  .preview-tools {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 12px;
    background: var(--bg-elevated);
    border-top: 1px solid var(--border-subtle);
  }
  .tools-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .resolution {
    font-size: 11px;
    color: var(--text-muted);
  }
  .history-tag {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 3px;
    background: color-mix(in srgb, var(--blue) 20%, var(--bg-elevated));
    color: var(--blue);
  }
  .tools-right {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .tool-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
  }
  .tool-btn:hover {
    background: var(--border-subtle);
    color: var(--text-secondary);
  }
  .tool-btn.active {
    background: var(--accent-glow);
    color: var(--accent);
  }

  .thumbnails {
    display: flex;
    gap: 6px;
    padding: 6px 12px;
    background: var(--bg-elevated);
    border-top: 1px solid var(--border-subtle);
    overflow-x: auto;
  }
  .thumb-btn {
    width: 48px;
    height: 48px;
    border-radius: 4px;
    overflow: hidden;
    border: 2px solid transparent;
    cursor: pointer;
    padding: 0;
    flex-shrink: 0;
    transition: border-color 0.15s;
  }
  .thumb-btn:hover {
    border-color: var(--text-muted);
  }
  .thumb-btn.active {
    border-color: var(--accent);
  }
  .thumb-btn img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>

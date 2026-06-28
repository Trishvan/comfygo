<script lang="ts">
  import { onMount } from "svelte";
  import { ListOutputs, GetOutputImage, GetHistory } from "../../wailsjs/go/orchestrator/Manager";

  export let onSelectImage: (url: string, width: number, height: number) => void;
  export let onBack: () => void;
  export let generating = false;

  let files: string[] = [];
  let selectedFile: string | null = null;
  let selectedUrl = "";
  let selectedEntry: any = null;
  let history: any[] = [];
  let loading = true;

  async function loadOutputs() {
    loading = true;
    try {
      files = await ListOutputs();
      history = await GetHistory();
    } catch {}
    loading = false;
  }

  function findHistory(filename: string): any {
    if (!filename) return null;
    // filename is like "20260101_120000_512x512.png"
    // History entries have the full path in `filename` field
    return history.find((h: any) => h.filename && h.filename.endsWith(filename));
  }

  async function selectFile(name: string) {
    selectedFile = name;
    try {
      const b64 = await GetOutputImage(name);
      if (b64) {
        selectedUrl = `data:image/png;base64,${b64}`;
        const entry = findHistory(name);
        selectedEntry = entry;
      }
    } catch {}
  }

  function openInPreview(name: string) {
    selectFile(name);
    // The onSelectImage will be handled by parent after we set selectedUrl
  }

  function viewInPreview() {
    if (selectedUrl && selectedFile) {
      const w = selectedEntry?.width || 0;
      const h = selectedEntry?.height || 0;
      onSelectImage(selectedUrl, w, h);
    }
  }

  onMount(() => {
    loadOutputs();
  });
</script>

<div class="gallery">
  <div class="gallery-header">
    <button class="back-btn" on:click={onBack}>
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M19 12H5M12 19l-7-7 7-7" />
      </svg>
      Back to Preview
    </button>
    <h2 class="gallery-title">Outputs</h2>
    {#if generating}
      <span class="gen-badge">⚡ Generating in background</span>
    {/if}
  </div>

  <div class="gallery-content">
    {#if loading}
      <div class="gallery-empty">
        <div class="spinner"></div>
        <p>Loading outputs...</p>
      </div>
    {:else if files.length === 0}
      <div class="gallery-empty">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" style="color: var(--text-muted);">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
        <p>No generated outputs yet</p>
      </div>
    {:else}
      <div class="gallery-grid">
        {#each files as name}
          {#await GetOutputImage(name) then b64}
            <button
              class="thumb-card"
              class:selected={selectedFile === name}
              on:click={() => selectFile(name)}
              ondblclick={() => viewInPreview()}
            >
              <img src="data:image/png;base64,{b64}" alt={name} loading="lazy" />
              <span class="thumb-name">{name}</span>
            </button>
          {:catch}
            <div class="thumb-card error-card">
              <span>Failed to load</span>
            </div>
          {/await}
        {/each}
      </div>

      {#if selectedUrl}
        <div class="detail-pane">
          <div class="detail-image">
            <img src={selectedUrl} alt={selectedFile || ""} />
          </div>
          <div class="detail-info">
            <div class="detail-row">
              <span class="detail-label">File</span>
              <span class="detail-value">{selectedFile}</span>
            </div>
            {#if selectedEntry}
              <div class="detail-row">
                <span class="detail-label">Prompt</span>
                <span class="detail-value detail-prompt">{selectedEntry.params?.prompt || "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Negative</span>
                <span class="detail-value">{selectedEntry.params?.negativePrompt || "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Resolution</span>
                <span class="detail-value">{selectedEntry.width || selectedEntry.params?.width || "--"} x {selectedEntry.height || selectedEntry.params?.height || "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Steps</span>
                <span class="detail-value">{selectedEntry.params?.steps || "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">CFG</span>
                <span class="detail-value">{selectedEntry.params?.cfgScale ?? "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Seed</span>
                <span class="detail-value">{selectedEntry.params?.seed ?? "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Sampler</span>
                <span class="detail-value">{selectedEntry.params?.samplerName || "--"}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Status</span>
                <span class="detail-value" class:green={selectedEntry.status === "completed"} class:red={selectedEntry.status === "failed" || selectedEntry.status === "cancelled"}>
                  {selectedEntry.status || "--"}
                </span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Time</span>
                <span class="detail-value">{selectedEntry.createdAt ? new Date(selectedEntry.createdAt).toLocaleString() : "--"}</span>
              </div>
            {/if}
            <button class="view-btn" on:click={viewInPreview}>
              View in Preview Panel
            </button>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .gallery {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .gallery-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 16px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .back-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    background: var(--bg-elevated);
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }
  .back-btn:hover {
    color: var(--text-primary);
    border-color: var(--text-muted);
  }

  .gallery-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    flex: 1;
  }

  .gen-badge {
    font-size: 11px;
    padding: 4px 10px;
    border-radius: 12px;
    background: color-mix(in srgb, var(--accent) 20%, var(--bg-elevated));
    color: var(--accent);
    border: 1px solid color-mix(in srgb, var(--accent) 30%, transparent);
  }

  .gallery-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    min-height: 0;
  }

  .gallery-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    height: 100%;
    color: var(--text-muted);
    font-size: 14px;
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 2px solid var(--border-subtle);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .gallery-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 12px;
    margin-bottom: 20px;
  }

  .thumb-card {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
    overflow: hidden;
    border: 2px solid var(--border-subtle);
    background: var(--bg-elevated);
    cursor: pointer;
    transition: all 0.15s;
    padding: 0;
  }
  .thumb-card:hover {
    border-color: var(--text-muted);
  }
  .thumb-card.selected {
    border-color: var(--accent);
    box-shadow: 0 0 12px var(--accent-glow);
  }
  .thumb-card img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    display: block;
  }
  .thumb-name {
    font-size: 10px;
    color: var(--text-muted);
    padding: 4px 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .error-card {
    padding: 24px;
    text-align: center;
    color: var(--red);
    font-size: 12px;
  }

  .detail-pane {
    display: flex;
    gap: 20px;
    padding: 16px;
    background: var(--bg-elevated);
    border-radius: 8px;
    border: 1px solid var(--border-subtle);
  }
  .detail-image {
    flex-shrink: 0;
    width: 200px;
  }
  .detail-image img {
    width: 100%;
    border-radius: 4px;
  }
  .detail-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .detail-row {
    display: flex;
    gap: 8px;
    font-size: 12px;
  }
  .detail-label {
    color: var(--text-muted);
    width: 80px;
    flex-shrink: 0;
  }
  .detail-value {
    color: var(--text-secondary);
    min-width: 0;
    word-break: break-word;
  }
  .detail-value.green {
    color: var(--green);
  }
  .detail-value.red {
    color: var(--red);
  }
  .detail-prompt {
    max-height: 60px;
    overflow-y: auto;
  }

  .view-btn {
    margin-top: 8px;
    padding: 6px 16px;
    border: none;
    border-radius: 6px;
    background: var(--btn-gradient);
    color: white;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    align-self: flex-start;
  }
  .view-btn:hover {
    opacity: 0.9;
  }
</style>

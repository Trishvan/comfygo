<script lang="ts">
  let { imageUrl, imageWidth, imageHeight, progress, state } = $props();
</script>

<div class="preview-container">
  {#if state === "generating"}
    <div class="placeholder">
      <div class="spinner"></div>
      <p>Generating...</p>
    </div>
  {:else if imageUrl}
    <img src={imageUrl} alt="Generated image" class="preview-img" />
    <div class="meta" class:visible={imageWidth > 0}>
      {imageWidth} x {imageHeight}
    </div>
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

<style>
  .preview-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
    width: 100%;
  }

  .placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    color: #555;
    font-size: 14px;
  }

  .placeholder.error {
    color: #f87171;
  }

  .preview-img {
    max-width: 100%;
    max-height: calc(100vh - 100px);
    border-radius: 8px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.4);
  }

  .meta {
    margin-top: 8px;
    font-size: 12px;
    color: #666;
    opacity: 0;
    transition: opacity 0.2s;
  }
  .meta.visible {
    opacity: 1;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #333;
    border-top-color: #a78bfa;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>

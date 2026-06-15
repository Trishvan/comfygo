<script lang="ts">
  export let label: string;
  export let value: string;
</script>

<div class="field">
  <label for="prompt-{label}">{label}</label>
  <textarea id="prompt-{label}" bind:value rows="3" placeholder="Enter {label.toLowerCase()}..."></textarea>
</div>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  label {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  textarea {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-primary);
    padding: 8px 10px;
    border-radius: 6px;
    font-size: 13px;
    font-family: inherit;
    resize: vertical;
    width: 100%;
  }

  textarea:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }
</style>

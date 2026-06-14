<script lang="ts">
  let { label, value } = $props();
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
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  textarea {
    background: #25252b;
    border: 1px solid #333;
    color: #e0e0e0;
    padding: 8px 10px;
    border-radius: 6px;
    font-size: 13px;
    font-family: inherit;
    resize: vertical;
    width: 100%;
  }

  textarea:focus {
    outline: none;
    border-color: #7c3aed;
  }
</style>

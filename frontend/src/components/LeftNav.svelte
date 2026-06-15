<script lang="ts">
  export let activeNav = "home";

  const navItems = [
    { id: "home", label: "Home", path: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" },
    { id: "projects", label: "Projects", path: "M2 7a2 2 0 012-2h4l2 2h6a2 2 0 012 2v1H2V7zm0 3h16v5a2 2 0 01-2 2H4a2 2 0 01-2-2v-5z" },
    { id: "workflows", label: "Workflows", path: "M8 4h8M8 4a2 2 0 11-4 0 2 2 0 014 0zm8 0a2 2 0 104 0 2 2 0 00-4 0zM8 12h8M8 12a2 2 0 11-4 0 2 2 0 014 0zm8 0a2 2 0 104 0 2 2 0 00-4 0zM8 20h8M8 20a2 2 0 11-4 0 2 2 0 014 0zm8 0a2 2 0 104 0 2 2 0 00-4 0z" },
    { id: "assets", label: "Assets", path: "M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" },
    { id: "models", label: "Models", path: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" },
    { id: "loras", label: "LoRAs", path: "M4 7l8 4 8-4-8-4-8 4zm0 5l8 4 8-4m0 4l-8 4-8-4" },
    { id: "controlnet", label: "ControlNet", path: "M8 16a4 4 0 100-8 4 4 0 000 8zm8-4a2 2 0 100-4 2 2 0 000 4zm0 8a2 2 0 100-4 2 2 0 000 4zM12 12l2-2m0 4l-2 2" },
    { id: "vae", label: "VAE", path: "M15 12a3 3 0 11-6 0 3 3 0 016 0z M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" },
    { id: "outputs", label: "Outputs", path: "M4 16l4-4 4 4m-2 2l2-2 4 4m-12-2a2 2 0 100 4 2 2 0 000-4zm10-10a2 2 0 100 4 2 2 0 000-4z M4 4h16a2 2 0 012 2v12a2 2 0 01-2 2H4a2 2 0 01-2-2V6a2 2 0 012-2z" },
    { id: "jobs", label: "Jobs", path: "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" },
    { id: "settings", label: "Settings", path: "M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z" },
  ];
</script>

<nav class="left-nav">
  {#each navItems as item}
    <button
      class="nav-btn"
      class:active={activeNav === item.id}
      title={item.label}
      on:click={() => (activeNav = item.id)}
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d={item.path} />
      </svg>
    </button>
  {/each}
</nav>

<style>
  .left-nav {
    grid-area: nav;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: 8px 0;
    background: var(--bg-primary);
    border-right: 1px solid var(--border-subtle);
    z-index: 10;
    overflow-y: auto;
  }

  .nav-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
  }
  .nav-btn:hover {
    background: var(--bg-elevated);
    color: var(--text-secondary);
  }
  .nav-btn.active {
    background: var(--accent-glow);
    color: var(--accent);
  }
</style>

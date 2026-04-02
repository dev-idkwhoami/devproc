<script lang="ts">
  import { getAppState, selectProject, addProject, removeProject } from './state.svelte.ts';

  let showDropdown = $state(false);

  const state = $derived(getAppState());

  const activeProjectData = $derived(
    state.projects.find((p) => p.name === state.activeProject)
  );

  async function handleSelect(name: string) {
    showDropdown = false;
    await selectProject(name);
  }

  async function handleAddProject() {
    showDropdown = false;
    await addProject();
  }

  async function handleDeselect() {
    showDropdown = false;
    await selectProject('');
  }

  function handleRemove(e: MouseEvent, name: string) {
    e.stopPropagation();
    removeProject(name);
  }
</script>

<div class="project-selector">
  <div class="label">PROJECT</div>
  <button class="selector" onclick={() => (showDropdown = !showDropdown)}>
    <span>{state.activeProject || 'No project selected'}</span>
    <span class="arrow">{showDropdown ? '▴' : '▾'}</span>
  </button>
  {#if activeProjectData}
    <div class="path">{activeProjectData.path}</div>
  {/if}

  {#if showDropdown}
    <div class="dropdown">
      {#if state.activeProject}
        <button class="dropdown-item deselect" onclick={handleDeselect}>
          ✕ Deselect project
        </button>
      {/if}
      {#each state.projects as project}
        <button
          class="dropdown-item"
          class:active={project.name === state.activeProject}
          onclick={() => handleSelect(project.name)}
        >
          <span>{project.name}</span>
          <span
            class="remove"
            role="button"
            tabindex="-1"
            onclick={(e) => handleRemove(e, project.name)}
            onkeydown={() => {}}
          >✕</span>
        </button>
      {/each}
      <button class="dropdown-item add" onclick={handleAddProject}>
        + Add project...
      </button>
    </div>
  {/if}
</div>

<style>
  .project-selector {
    padding: 10px 12px;
    border-bottom: 2px solid var(--border);
    position: relative;
  }
  .selector {
    width: 100%;
    background: var(--surface);
    border: 1px solid #444;
    color: var(--text);
    padding: 6px 10px;
    font-size: 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    text-transform: none;
    letter-spacing: 0;
  }
  .arrow {
    color: var(--text-muted);
  }
  .path {
    font-size: 9px;
    color: var(--text-muted);
    margin-top: 3px;
  }
  .dropdown {
    position: absolute;
    left: 12px;
    right: 12px;
    top: 100%;
    background: var(--surface);
    border: 1px solid #444;
    z-index: 10;
    max-height: 200px;
    overflow-y: auto;
  }
  .dropdown-item {
    width: 100%;
    background: transparent;
    color: var(--text);
    padding: 6px 10px;
    font-size: 11px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    text-transform: none;
    letter-spacing: 0;
  }
  .dropdown-item:hover {
    background: #1a1a1a;
  }
  .dropdown-item.active {
    color: var(--green);
  }
  .dropdown-item.add {
    color: var(--text-dim);
    border-top: 1px solid var(--border);
  }
  .dropdown-item.deselect {
    color: var(--text-muted);
    border-bottom: 1px solid var(--border);
  }
  .remove {
    color: var(--text-muted);
    font-size: 9px;
  }
  .remove:hover {
    color: var(--red);
  }
</style>

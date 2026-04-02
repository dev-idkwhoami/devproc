<script lang="ts">
  import { getAppState } from './state.svelte.ts';
  import CommandItem from './CommandItem.svelte';

  const state = $derived(getAppState());
</script>

<div class="command-list">
  <div class="label">COMMANDS</div>

  {#if !state.activeProject}
    <div class="empty">Select a project to see commands</div>
  {:else if state.commands.length === 0}
    <div class="empty">No commands configured</div>
  {:else}
    {#each state.commands as command (command.name)}
      <CommandItem {command} projectName={state.activeProject} />
    {/each}
  {/if}
</div>

<style>
  .command-list {
    padding: 8px 12px;
    flex: 1;
    overflow-y: auto;
  }
  .empty {
    color: var(--text-muted);
    font-size: 10px;
    padding: 12px 0;
    text-align: center;
  }
</style>

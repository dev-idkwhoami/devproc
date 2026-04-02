<script lang="ts">
  import { getAppState, startAll, killAll } from './state.svelte.ts';

  interface Props {
    onAddCommand: () => void;
  }

  let { onAddCommand }: Props = $props();

  const state = $derived(getAppState());
</script>

<div class="bottom-bar">
  <button class="btn-add" onclick={onAddCommand} disabled={!state.activeProject}>
    + ADD CMD
  </button>
  <div class="actions">
    <button class="btn-start-all" onclick={() => startAll()} disabled={!state.activeProject}>
      START ALL
    </button>
    <button class="btn-kill-all" onclick={() => killAll()} disabled={!state.activeProject}>
      KILL ALL
    </button>
  </div>
</div>

<style>
  .bottom-bar {
    border-top: 2px solid var(--border);
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .btn-add {
    background: transparent;
    color: var(--text-dim);
    border: 1px solid #444;
  }
  .btn-add:hover:not(:disabled) {
    color: var(--text);
    border-color: var(--text-dim);
  }
  .actions {
    display: flex;
    gap: 6px;
  }
  .btn-start-all {
    background: var(--green);
    color: black;
  }
  .btn-kill-all {
    background: var(--red);
    color: white;
  }
  button:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
</style>

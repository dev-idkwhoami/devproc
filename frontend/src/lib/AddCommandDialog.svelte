<script lang="ts">
  import { getAppState, addCommand } from './state.svelte.ts';

  interface Props {
    onclose: () => void;
  }

  let { onclose }: Props = $props();

  const state = $derived(getAppState());

  let cmdName = $state('');
  let cmdStr = $state('');

  function selectHistory(cmd: string) {
    cmdStr = cmd;
  }

  async function handleSave() {
    if (!cmdName.trim() || !cmdStr.trim()) return;
    await addCommand(state.activeProject, cmdName.trim(), cmdStr.trim());
    onclose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose();
    if (e.key === 'Enter') handleSave();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="overlay" onclick={onclose} role="dialog">
  <div class="dialog" onclick={(e) => e.stopPropagation()} role="document">
    <div class="dialog-header">
      <span>ADD COMMAND</span>
      <button class="close" onclick={onclose}>✕</button>
    </div>

    <div class="dialog-body">
      <div class="field">
        <div class="label">NAME</div>
        <input
          type="text"
          bind:value={cmdName}
          placeholder="e.g. Vite"
          class="input"
        />
      </div>

      <div class="field">
        <div class="label">COMMAND</div>
        <input
          type="text"
          bind:value={cmdStr}
          placeholder="e.g. npm run dev"
          class="input"
        />
      </div>

      {#if state.commandHistory.length > 0}
        <div class="field">
          <div class="label">HISTORY</div>
          <div class="history">
            {#each state.commandHistory as cmd}
              <button class="history-item" onclick={() => selectHistory(cmd)}>
                {cmd}
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <div class="dialog-footer">
      <button class="btn-cancel" onclick={onclose}>CANCEL</button>
      <button class="btn-save" onclick={handleSave}>SAVE</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .dialog {
    background: var(--bg);
    border: 2px solid var(--border);
    width: 290px;
  }
  .dialog-header {
    background: #1a1a1a;
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border);
    font-size: 10px;
    font-weight: bold;
    letter-spacing: 2px;
    color: var(--text-dim);
  }
  .close {
    background: transparent;
    color: var(--text-muted);
    font-size: 10px;
    padding: 2px 6px;
  }
  .dialog-body {
    padding: 12px;
  }
  .field {
    margin-bottom: 10px;
  }
  .input {
    width: 100%;
    background: var(--surface);
    border: 1px solid #444;
    color: var(--text);
    padding: 6px 10px;
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
  }
  .input:focus {
    border-color: var(--text-dim);
  }
  .history {
    max-height: 100px;
    overflow-y: auto;
    border: 1px solid #444;
  }
  .history-item {
    width: 100%;
    background: transparent;
    color: var(--text);
    padding: 4px 10px;
    font-size: 11px;
    text-align: left;
    text-transform: none;
    letter-spacing: 0;
  }
  .history-item:hover {
    background: #1a1a1a;
  }
  .dialog-footer {
    padding: 8px 12px;
    border-top: 2px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: 6px;
  }
  .btn-cancel {
    background: transparent;
    color: var(--text-dim);
    border: 1px solid #444;
  }
  .btn-save {
    background: var(--green);
    color: black;
  }
</style>

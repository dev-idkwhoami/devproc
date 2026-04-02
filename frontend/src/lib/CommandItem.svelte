<script lang="ts">
  import { startCommand, killCommand, removeCommand } from './state.svelte.ts';
  import type { CommandStatus } from './state.svelte.ts';

  interface Props {
    command: CommandStatus;
    projectName: string;
  }

  let { command, projectName }: Props = $props();

  let showDelete = $state(false);
  let confirmDelete = $state(false);

  function handleAction() {
    if (command.status === 'running') {
      killCommand(command.name);
    } else {
      startCommand(command.name);
    }
  }

  function handleDelete() {
    if (confirmDelete) {
      removeCommand(projectName, command.name);
      confirmDelete = false;
    } else {
      confirmDelete = true;
      setTimeout(() => { confirmDelete = false; }, 3000);
    }
  }

  const dotClass = $derived(
    command.status === 'running' ? 'dot green' :
    command.status === 'errored' ? 'dot red' : 'dot grey'
  );

  const buttonLabel = $derived(
    command.status === 'running' ? 'KILL' :
    command.status === 'errored' ? 'RESTART' : 'START'
  );

  const buttonClass = $derived(
    command.status === 'running' ? 'btn-kill' :
    command.status === 'errored' ? 'btn-restart' : 'btn-start'
  );

  const statusText = $derived(
    command.status === 'running' ? command.cmd :
    command.statusMsg || command.cmd
  );

  const statusClass = $derived(
    command.status === 'errored' ? 'status errored' : 'status'
  );
</script>

<div
  class="command-item"
  onmouseenter={() => (showDelete = true)}
  onmouseleave={() => (showDelete = false)}
  role="listitem"
>
  <div class="left">
    <span class={dotClass}></span>
    <div class="info">
      <div class="name">{command.name}</div>
      <div class={statusClass}>{statusText}</div>
    </div>
  </div>
  <div class="right">
    {#if confirmDelete}
      <button class="btn-confirm-delete" onclick={handleDelete}>SURE?</button>
    {:else if showDelete}
      <button class="btn-delete" onclick={handleDelete}>✕</button>
    {/if}
    <button class={buttonClass} onclick={handleAction}>{buttonLabel}</button>
  </div>
</div>

<style>
  .command-item {
    background: var(--surface);
    border: 1px solid #2a2a2a;
    padding: 8px 10px;
    margin-bottom: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .dot.green {
    background: var(--green);
    box-shadow: 0 0 6px var(--green);
  }
  .dot.red {
    background: var(--red);
    box-shadow: 0 0 6px var(--red);
  }
  .dot.grey {
    background: #666;
  }
  .info {
    min-width: 0;
  }
  .name {
    font-size: 12px;
    font-weight: bold;
  }
  .status {
    font-size: 9px;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .status.errored {
    color: var(--red);
  }
  .right {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }
  .btn-kill {
    background: var(--red);
    color: white;
  }
  .btn-start {
    background: var(--green);
    color: black;
  }
  .btn-restart {
    background: var(--yellow);
    color: black;
  }
  .btn-delete {
    background: transparent;
    color: var(--text-muted);
    padding: 2px 6px;
    font-size: 9px;
  }
  .btn-delete:hover {
    color: var(--red);
  }
  .btn-confirm-delete {
    background: var(--red);
    color: white;
    padding: 2px 6px;
    font-size: 9px;
  }
</style>

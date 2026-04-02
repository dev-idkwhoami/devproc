<script lang="ts">
  import { HideWindow } from '../../wailsjs/go/main/App';
</script>

<div class="titlebar" role="banner">
  <span class="title">DEVPROC</span>
  <button class="close-btn" onclick={() => HideWindow()}>✕</button>
</div>

<style>
  .titlebar {
    background: #1a1a1a;
    padding: 8px 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border);
    --wails-draggable: drag;
    cursor: grab;
  }
  .titlebar:active {
    cursor: grabbing;
  }
  .title {
    font-size: 11px;
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 2px;
    color: var(--text-dim);
  }
  .close-btn {
    background: transparent;
    color: var(--text-muted);
    padding: 2px 6px;
    font-size: 10px;
    --wails-draggable: no-drag;
  }
  .close-btn:hover {
    color: var(--red);
  }
</style>

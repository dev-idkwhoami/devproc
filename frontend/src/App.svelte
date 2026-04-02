<script lang="ts">
  import { onMount } from 'svelte';
  import { loadState, initEventListeners } from './lib/state.svelte.ts';
  import TitleBar from './lib/TitleBar.svelte';
  import ProjectSelector from './lib/ProjectSelector.svelte';
  import CommandList from './lib/CommandList.svelte';
  import BottomBar from './lib/BottomBar.svelte';
  import AddCommandDialog from './lib/AddCommandDialog.svelte';

  let showAddDialog = $state(false);

  onMount(() => {
    loadState();
    initEventListeners();
  });
</script>

<div class="app">
  <TitleBar />
  <ProjectSelector />
  <CommandList />
  <BottomBar onAddCommand={() => (showAddDialog = true)} />

  {#if showAddDialog}
    <AddCommandDialog onclose={() => (showAddDialog = false)} />
  {/if}
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 320px;
  }
</style>

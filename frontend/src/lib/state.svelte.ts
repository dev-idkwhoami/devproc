import { GetState, SelectProject, AddProject, RemoveProject, AddCommand, RemoveCommand, StartCommand, KillCommand, StartAll, KillAll } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export interface CommandStatus {
  name: string;
  cmd: string;
  status: 'stopped' | 'running' | 'errored';
  statusMsg: string;
}

export interface Project {
  name: string;
  path: string;
  commands: { name: string; cmd: string }[];
}

export interface AppState {
  activeProject: string;
  projects: Project[];
  commands: CommandStatus[];
  commandHistory: string[];
}

let appState = $state<AppState>({
  activeProject: '',
  projects: [],
  commands: [],
  commandHistory: [],
});

export function getAppState(): AppState {
  return appState;
}

export async function loadState() {
  const state = await GetState();
  appState.activeProject = state.activeProject;
  appState.projects = state.projects;
  appState.commands = state.commands;
  appState.commandHistory = state.commandHistory;
}

export async function selectProject(name: string) {
  await SelectProject(name);
  await loadState();
}

export async function addProject() {
  await AddProject();
  await loadState();
}

export async function removeProject(name: string) {
  await RemoveProject(name);
}

export async function addCommand(projectName: string, cmdName: string, cmd: string) {
  await AddCommand(projectName, cmdName, cmd);
}

export async function removeCommand(projectName: string, cmdName: string) {
  await RemoveCommand(projectName, cmdName);
}

export async function startCommand(name: string) {
  await StartCommand(name);
}

export async function killCommand(name: string) {
  await KillCommand(name);
}

export async function startAll() {
  await StartAll();
}

export async function killAll() {
  await KillAll();
}

export function initEventListeners() {
  EventsOn('state-update', (state: AppState) => {
    appState.activeProject = state.activeProject;
    appState.projects = state.projects;
    appState.commands = state.commands;
    appState.commandHistory = state.commandHistory;
  });
}

<script lang="ts" setup>
import { fs as models } from '../../wailsjs/go/models';
import FolderTree from './FolderTree.vue';

defineProps<{
  isTreeLoading: boolean;
  treeError: string;
  folderTreeRoot: models.Folder | null;
  currentFolder: string;
}>();

const emit = defineEmits<{
  (e: 'selectFolder'): void;
  (e: 'folderSelected', path: string): void;
  (e: 'loadSubfolders', folder: models.Folder): void;
}>();
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <button @click="emit('selectFolder')" title="Select Root Folder for Tree">Browse...</button>
    </div>
    <div v-if="isTreeLoading" class="loading-tree">Loading Tree...</div>
    <div v-else-if="treeError" class="tree-error">{{ treeError }}</div>
    <div v-else-if="folderTreeRoot" class="folder-tree-container">
      <FolderTree 
        :folder="folderTreeRoot" 
        :selectedPath="currentFolder"
        @folder-selected="emit('folderSelected', $event)"
        @load-subfolders="emit('loadSubfolders', $event)" />
    </div>
    <div v-else class="no-tree">
      Click 'Browse' to select a root folder.
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 350px;
  min-width: 200px;
  max-width: 50%;
  resize: horizontal;
  overflow: auto;
  flex-shrink: 0;
  background-color: var(--background);
  border-right: 1px solid var(--current-line);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
    padding: 10px;
    border-bottom: 1px solid var(--current-line);
}

.sidebar-header button {
    width: 100%;
}

.loading-tree,
.tree-error,
.no-tree {
    padding: 15px;
    text-align: center;
    color: var(--comment);
}

.tree-error {
    color: var(--red);
}

.folder-tree-container {
    flex-grow: 1;
    padding: 5px 0;
    overflow-y: auto;
}

button {
  padding: 6px 12px;
  background-color: var(--purple);
  border: none;
  border-radius: 4px;
  color: var(--foreground);
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s ease;
}

button:hover:not(:disabled) {
  background-color: var(--pink);
}
</style>

<script lang="ts" setup>
import { ref, defineProps, defineEmits, computed, watch } from 'vue';
// Import the main namespace from models
import { main } from '../../wailsjs/go/models';

const props = defineProps<{
  // Use the namespaced type
  folder: main.Folder;
  level?: number; // For indentation
  selectedPath?: string | null;
}>();

const emit = defineEmits<{
  (e: 'folder-selected', path: string): void
}>();

const isOpen = ref(true); // Folders start open by default, could be changed
const indent = computed(() => (props.level || 0) * 15); // Indentation level in pixels

function toggleFolder() {
  isOpen.value = !isOpen.value;
}

function selectFolder(path: string, event: MouseEvent) {
  event.stopPropagation(); // Prevent toggle when selecting
  emit('folder-selected', path);
}

// Check if the current folder or any child is selected
const isSelectedOrParentOfSelected = computed(() => {
    if (!props.selectedPath) return false;
    if (props.folder.path === props.selectedPath) return true;
    // Check if selected path starts with this folder's path + separator
    // This ensures '/base/folder' doesn't match '/base/folder-other'
    // Handle both Windows and Unix separators
    const separator = props.folder.path.includes('\\') ? '\\' : '/';
    const pathWithSeparator = props.folder.path.endsWith(separator) ? props.folder.path : props.folder.path + separator;
    return props.selectedPath.startsWith(pathWithSeparator);
});

// Keep the folder open if it or a child is the selected path
// Watch for changes in selectedPath to potentially re-open folders
watch(() => props.selectedPath, (newPath, oldPath) => {
    if (isSelectedOrParentOfSelected.value) {
        isOpen.value = true;
    }
}, { immediate: true }); // Run immediately on component mount

</script>

<template>
  <div class="folder-node">
    <div 
      class="folder-item" 
      :style="{ paddingLeft: indent + 'px' }" 
      :class="{ selected: folder.path === selectedPath }"
      @click="selectFolder(folder.path, $event)"
    >
      <!-- Toggle Icon -->
      <span v-if="folder.children && folder.children.length > 0" 
            class="toggle-icon" 
            @click.stop="toggleFolder">{{ isOpen ? '&#9660;' : '&#9654;' }}</span> <!-- Down/Right arrow -->
      <span v-else class="toggle-icon spacer"></span> <!-- Placeholder for alignment -->
      <!-- Folder Name -->
      <span class="folder-name">{{ folder.name }}</span>
    </div>
    <div v-if="isOpen && folder.children && folder.children.length > 0" class="children">
      <FolderTree 
        v-for="child in folder.children" 
        :key="child.path" 
        :folder="child" 
        :level="(level || 0) + 1"
        :selectedPath="selectedPath"
        @folder-selected="(path) => emit('folder-selected', path)" 
      />
    </div>
  </div>
</template>

<style scoped>
/* Inherit variables from parent or define locally if needed */

.folder-node {
  /* Spacing between nodes if needed */
   text-align: left; /* Ensure content defaults to left alignment */
}

.folder-item {
  padding: 4px 5px;
  cursor: pointer;
  display: flex;
  align-items: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  /* color: inherit; */ /* Inherit from parent by default */
  /* padding-left is set dynamically via :style */
  /* text-align: left; */ /* Flex handles horizontal alignment, default is start */
}

.folder-item:hover {
  background-color: var(--current-line); /* Use variable */
}

.folder-item.selected {
    background-color: var(--purple); /* Use variable for selected */
    color: var(--foreground);
    font-weight: bold;
}

.toggle-icon {
  display: inline-block;
  width: 15px; /* Ensure space for icon */
  text-align: center;
  margin-right: 5px;
  font-size: 0.8em;
  color: var(--comment); /* Use variable */
  flex-shrink: 0; /* Prevent icon/spacer from shrinking */
}

.toggle-icon.spacer {
    /* Just takes up space */
    visibility: hidden;
}

.folder-name {
  /* Style for folder name */
  overflow: hidden; /* Handle overflow within the name span */
  text-overflow: ellipsis; /* Add ellipsis if name is too long */
  flex-grow: 1; /* Allow name to take up remaining space */
  min-width: 0; /* Important for text-overflow in flex items */
  /* text-align: left; */ /* Inherits from parent */
}

.children {
  /* Style for the container of child nodes */
   text-align: left; /* Ensure children container aligns content left */
}
</style>

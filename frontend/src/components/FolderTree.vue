<script lang="ts" setup>
import { ref, defineProps, defineEmits, computed, watch } from 'vue';
// Import the fs namespace from models
import { fs as models } from '../../wailsjs/go/models';

const props = defineProps<{
  folder: models.Folder;
  level?: number;
  selectedPath?: string | null;
}>();

const emit = defineEmits<{
  (e: 'folder-selected', path: string): void,
  (e: 'load-subfolders', folder: models.Folder): void
}>();

const isOpen = ref(false); // Changed to false by default for lazy loading
const isSubLoading = ref(false);
const indent = computed(() => (props.level || 0) * 15);

async function toggleFolder() {
  if (!isOpen.value && props.folder.hasChildren && (!props.folder.children || props.folder.children.length === 0)) {
    isSubLoading.value = true;
    emit('load-subfolders', props.folder);
    // We don't await here directly as the parent handles it and updates the object reference
    // But we might need a small delay or watch to set loading to false
    // For simplicity, let's assume it's fast enough or handled by the parent
    setTimeout(() => { isSubLoading.value = false; }, 500);
  }
  isOpen.value = !isOpen.value;
}

function selectFolder(path: string, event: MouseEvent) {
  event.stopPropagation();
  emit('folder-selected', path);
}

const isSelectedOrParentOfSelected = computed(() => {
    if (!props.selectedPath) return false;
    if (props.folder.path === props.selectedPath) return true;
    const separator = props.folder.path.includes('\\') ? '\\' : '/';
    const pathWithSeparator = props.folder.path.endsWith(separator) ? props.folder.path : props.folder.path + separator;
    return props.selectedPath.startsWith(pathWithSeparator);
});

watch(() => props.selectedPath, (newPath) => {
    if (isSelectedOrParentOfSelected.value && !isOpen.value) {
        toggleFolder();
    }
}, { immediate: true });

</script>

<template>
  <div class="folder-node">
    <div 
      class="folder-item" 
      :style="{ paddingLeft: indent + 'px' }" 
      :class="{ selected: folder.path === selectedPath }"
      @click="selectFolder(folder.path, $event)"
    >
      <span v-if="folder.hasChildren" 
            class="toggle-icon" 
            @click.stop="toggleFolder">
            <template v-if="isSubLoading">...</template>
            <template v-else>{{ isOpen ? '&#9660;' : '&#9654;' }}</template>
      </span>
      <span v-else class="toggle-icon spacer"></span>
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
        @load-subfolders="(f) => emit('load-subfolders', f)"
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

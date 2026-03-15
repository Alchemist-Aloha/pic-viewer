<script lang="ts" setup>
import { SlideshowMode } from '../composables/useSlideshow';

defineProps<{
  currentFolder: string;
  currentFilename: string;
  slideshowActive: boolean;
  slideshowMode: SlideshowMode;
  slideshowDelay: number;
  isLoading: boolean;
  hasImages: boolean;
  hasLeafFolders: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:slideshowMode', value: SlideshowMode): void;
  (e: 'update:slideshowDelay', value: number): void;
  (e: 'toggleSlideshow'): void;
}>();
</script>

<template>
  <div class="controls">
    <span v-if="currentFolder" class="folder-path">Current: {{ currentFolder }}</span>
    <span v-else>No folder selected</span>
    <span v-if="currentFilename" class="file-name" :title="currentFilename">File: {{ currentFilename }}</span>

    <div class="slideshow-controls">
      <button @click="emit('toggleSlideshow')"
              :disabled="isLoading || (!hasImages && !hasLeafFolders)">
        {{ slideshowActive ? 'Stop' : 'Start' }}
      </button>
      <select :value="slideshowMode" 
              @change="emit('update:slideshowMode', ($event.target as HTMLSelectElement).value as SlideshowMode)"
              :disabled="slideshowActive || isLoading">
        <option value="sequence">Sequence</option>
        <option value="random-folder" :disabled="!hasImages">Random (Folder)</option>
        <option value="random-all" :disabled="!hasLeafFolders">Random (All)</option>
      </select>
      <label>
        Delay (ms):
        <input type="number" 
               :value="slideshowDelay"
               @input="emit('update:slideshowDelay', +($event.target as HTMLInputElement).value)"
               min="500" step="100" 
               :disabled="slideshowActive || isLoading" 
               class="delay-input">
      </label>
    </div>
  </div>
</template>

<style scoped>
.controls {
  padding: 10px;
  background-color: var(--current-line);
  border-bottom: 1px solid var(--background);
  display: flex;
  align-items: center;
  gap: 15px;
  flex-shrink: 0;
}

.folder-path {
  font-size: 0.9em;
  color: var(--foreground);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-grow: 1;
  min-width: 0;
}

.file-name {
  font-size: 0.9em;
  color: var(--cyan);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}

.slideshow-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 5px;
  border-left: 1px solid var(--background);
}

.slideshow-controls label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.9em;
}

.delay-input {
  width: 60px;
  padding: 4px;
  background-color: var(--background);
  color: var(--foreground);
  border: 1px solid var(--comment);
  border-radius: 3px;
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

button:disabled {
  background-color: var(--comment);
  color: var(--background);
  cursor: not-allowed;
}

select {
  padding: 6px 8px;
  background-color: var(--purple);
  color: var(--foreground);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

select:disabled {
  background-color: var(--comment);
  color: var(--background);
  cursor: not-allowed;
}
</style>

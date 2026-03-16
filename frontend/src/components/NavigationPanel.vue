<script lang="ts" setup>
defineProps<{
  currentIndex: number;
  totalImages: number;
  isLoading: boolean;
  isTreeLoading: boolean;
  lastVisitedFolder: string;
  hasNextLeafFolder: boolean;
  hasPreviousLeafFolder: boolean;
  rootPath: string;
}>();

const emit = defineEmits<{
  (e: 'prevImage'): void;
  (e: 'nextImage'): void;
  (e: 'goToLastVisitedFolder'): void;
  (e: 'goToPrevLeafFolder'): void;
  (e: 'goToNextLeafFolder'): void;
  (e: 'goToRandomFolder'): void;
}>();
</script>

<template>
  <div class="navigation" v-if="totalImages > 0 || hasNextLeafFolder || hasPreviousLeafFolder">
    <button @click="emit('prevImage')" :disabled="isLoading || (currentIndex <= 0 && !hasPreviousLeafFolder)">Previous</button>
    <span v-if="totalImages > 0">{{ currentIndex + 1 }} / {{ totalImages }}</span>
    <button 
      @click="emit('nextImage')" 
      :disabled="isLoading || (currentIndex >= totalImages - 1 && !hasNextLeafFolder)"
      title="Go to the next image or next folder">
      Next
    </button>
    <button @click="emit('goToLastVisitedFolder')" :disabled="isTreeLoading || !lastVisitedFolder" title="Go to the previously visited folder">
      Last Visited
    </button>
    <button @click="emit('goToPrevLeafFolder')" :disabled="isTreeLoading || !hasPreviousLeafFolder" title="Go to the previous leaf folder in the tree">
      Prev Folder
    </button>
    <button @click="emit('goToNextLeafFolder')" :disabled="isTreeLoading || !hasNextLeafFolder" title="Go to the next leaf folder in the tree">
      Next Folder
    </button>
    <button @click="emit('goToRandomFolder')" :disabled="isTreeLoading || !rootPath" title="Go to a random folder with images">
      Random Folder
    </button>
  </div>
</template>

<style scoped>
.navigation {
  padding: 10px;
  background-color: var(--current-line);
  border-top: 1px solid var(--background);
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  flex-wrap: wrap;
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
</style>

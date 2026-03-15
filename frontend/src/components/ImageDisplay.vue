<script lang="ts" setup>
import { CSSProperties } from 'vue';

defineProps<{
  currentImageSrc: string;
  imageStyle: CSSProperties;
  isLoading: boolean;
  currentFolder: string;
  hasImages: boolean;
}>();

const emit = defineEmits<{
  (e: 'wheel', event: WheelEvent): void;
}>();
</script>

<template>
  <div class="image-viewer" @wheel="emit('wheel', $event)">
    <div v-if="isLoading" class="loading">Loading Image...</div>
    <img v-else-if="currentImageSrc"
         :src="currentImageSrc"
         alt="Current Image"
         :style="imageStyle" />
    <div v-else-if="currentFolder && !hasImages && !isLoading" class="no-images">
      No images found in this folder.
    </div>
    <div v-else class="no-folder">
      Select a folder from the tree.
    </div>
  </div>
</template>

<style scoped>
.image-viewer {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: auto;
  position: relative;
  cursor: grab;
  min-height: 0;
  background-color: var(--background);
}
.image-viewer:active {
    cursor: grabbing;
}

.loading,
.no-images,
.no-folder {
  font-size: 1.2em;
  color: var(--comment);
  padding: 20px;
}
</style>

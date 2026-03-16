<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { SelectFolder, ListImages, ReadImage, FindNextFolder, FindPrevFolder, FindRandomFolderWithImages } from '../../wailsjs/go/main/App';
import { LogError } from '../../wailsjs/runtime/runtime';

import Sidebar from './Sidebar.vue';
import ControlPanel from './ControlPanel.vue';
import ImageDisplay from './ImageDisplay.vue';
import NavigationPanel from './NavigationPanel.vue';

import { useImageZoom } from '../composables/useImageZoom';
import { useFolderTree } from '../composables/useFolderTree';
import { useImageGallery } from '../composables/useImageGallery';
import { useSlideshow, SlideshowMode } from '../composables/useSlideshow';

const currentFolder = ref<string>("");
const rootPath = ref<string>("");
const lastVisitedFolder = ref<string>("");

const { 
  zoomIn, zoomOut, resetZoom, imageStyle 
} = useImageZoom();

const { 
  folderTreeRoot, isTreeLoading, treeError, 
  loadFolderTree, loadSubfolders, clearTree 
} = useFolderTree();

const { 
  images, currentIndex, currentImageSrc, isLoading, currentFilename, 
  hasNextLeafFolder, hasPreviousLeafFolder, preloadedImageSrc, preloadedIndex, 
  preloadedFolder, preloadNextImage, clearPreload, displayCurrentImage, loadImagesForPath 
} = useImageGallery(currentFolder, rootPath);

const { 
  slideshowActive, slideshowDelay, slideshowMode, toggleSlideshow, stopSlideshow 
} = useSlideshow(
  nextImage,
  displayRandomImageInFolder,
  async () => {}, // displayRandomImageInAllFolders disabled, made async for TS
  canStartSlideshow
);

function canStartSlideshow(mode: SlideshowMode): boolean {
  if (mode === 'sequence' && (images.value.length > 0 || hasNextLeafFolder.value)) return true;
  if (mode === 'random-folder' && images.value.length > 0) return true;
  return false;
}

async function selectFolder() {
  clearTree();
  images.value = [];
  currentIndex.value = -1;
  currentImageSrc.value = "";
  currentFolder.value = "";
  rootPath.value = "";

  try {
    const selectedPath = await SelectFolder();
    if (selectedPath) {
      rootPath.value = selectedPath;
      // Load images for the root folder itself
      await loadImagesForPath(selectedPath, resetZoom);
      // Load only immediate children for the tree
      await loadFolderTree(selectedPath);
    }
  } catch (err) {
    LogError("Error selecting folder: " + err);
  }
}

async function handleFolderSelected(selectedPath: string) {
  if (!selectedPath || selectedPath === currentFolder.value) return;

  if (currentFolder.value) {
    lastVisitedFolder.value = currentFolder.value;
  }

  await loadImagesForPath(selectedPath, resetZoom);
}

function handleLoadSubfolders(folder: any) {
    loadSubfolders(folder);
}

function goToLastVisitedFolder() {
  if (lastVisitedFolder.value) handleFolderSelected(lastVisitedFolder.value);
}

function displayRandomImageInFolder() {
  if (images.value.length < 1) return;
  let randomIndex;
  if (images.value.length === 1) {
    randomIndex = 0;
  } else {
    do {
      randomIndex = Math.floor(Math.random() * images.value.length);
    } while (randomIndex === currentIndex.value);
  }
  currentIndex.value = randomIndex;
  displayCurrentImage(resetZoom);
}

async function nextImage() {
  if (images.value.length === 0 && !hasNextLeafFolder.value) return;

  let targetIdx = currentIndex.value + 1;
  
  if (targetIdx >= images.value.length) {
      // Try to go to next folder
      try {
          const nextFolder = await FindNextFolder(currentFolder.value, rootPath.value);
          if (nextFolder) {
              await handleFolderSelected(nextFolder);
              return;
          }
      } catch (err) {
          LogError("Error finding next folder: " + err);
      }
      stopSlideshow();
      return;
  }

  if (slideshowMode.value === 'sequence' && targetIdx === preloadedIndex.value && currentFolder.value === preloadedFolder.value && preloadedImageSrc.value) {
    currentImageSrc.value = preloadedImageSrc.value;
    currentIndex.value = targetIdx;
    resetZoom();
    clearPreload();
    preloadNextImage(slideshowActive.value, slideshowMode.value);
  } else {
    currentIndex.value = targetIdx;
    await displayCurrentImage(resetZoom);
    preloadNextImage(slideshowActive.value, slideshowMode.value);
  }
}

async function prevImage() {
  if (images.value.length === 0 && !hasPreviousLeafFolder.value) return;
  clearPreload();

  let targetIdx = currentIndex.value - 1;

  if (targetIdx < 0) {
      try {
          const prevFolder = await FindPrevFolder(currentFolder.value, rootPath.value);
          if (prevFolder) {
              await handleFolderSelected(prevFolder);
              // Set to last image of prev folder
              if (images.value.length > 0) {
                  currentIndex.value = images.value.length - 1;
                  await displayCurrentImage(resetZoom);
              }
              return;
          }
      } catch (err) {
          LogError("Error finding prev folder: " + err);
      }
      return;
  }

  currentIndex.value = targetIdx;
  await displayCurrentImage(resetZoom);
}

function handleWheel(event: WheelEvent) {
  event.preventDefault();
  if (event.ctrlKey) {
    if (!currentImageSrc.value) return;
    event.deltaY < 0 ? zoomIn() : zoomOut();
  } else {
    if (event.deltaY < 0) {
      if (!isLoading.value && (currentIndex.value > 0 || hasPreviousLeafFolder.value)) prevImage();
    } else if (event.deltaY > 0) {
      if (!isLoading.value && (currentIndex.value < images.value.length - 1 || hasNextLeafFolder.value)) nextImage();
    }
  }
}

function handleKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowRight': nextImage(); break;
    case 'ArrowLeft': prevImage(); break;
    case ' ': toggleSlideshow(); event.preventDefault(); break;
    case '+': case '=': zoomIn(); event.preventDefault(); break;
    case '-': zoomOut(); event.preventDefault(); break;
  }
}

async function goToRandomFolder() {
  if (!rootPath.value) return;
  try {
    const randomFolder = await FindRandomFolderWithImages(rootPath.value);
    if (randomFolder) handleFolderSelected(randomFolder);
  } catch (err) {
    LogError("Error going to random folder: " + err);
  }
}

async function goToNextLeafFolder() {
  try {
      const next = await FindNextFolder(currentFolder.value, rootPath.value);
      if (next) handleFolderSelected(next);
  } catch (err) {
      LogError("Error going to next folder: " + err);
  }
}

async function goToPrevLeafFolder() {
  try {
      const prev = await FindPrevFolder(currentFolder.value, rootPath.value);
      if (prev) handleFolderSelected(prev);
  } catch (err) {
      LogError("Error going to prev folder: " + err);
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  stopSlideshow();
  window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <div class="layout-container">
    <Sidebar 
      :isTreeLoading="isTreeLoading"
      :treeError="treeError"
      :folderTreeRoot="folderTreeRoot"
      :currentFolder="currentFolder"
      @selectFolder="selectFolder"
      @folderSelected="handleFolderSelected"
      @loadSubfolders="handleLoadSubfolders"
    />

    <div class="main-content">
      <div class="container">
        <ControlPanel 
          v-model:slideshowMode="slideshowMode"
          v-model:slideshowDelay="slideshowDelay"
          :currentFolder="currentFolder"
          :currentFilename="currentFilename"
          :slideshowActive="slideshowActive"
          :isLoading="isLoading"
          :hasImages="images.length > 0"
          :hasLeafFolders="false"
          @toggleSlideshow="toggleSlideshow"
        />

        <ImageDisplay 
          :currentImageSrc="currentImageSrc"
          :imageStyle="imageStyle"
          :isLoading="isLoading"
          :currentFolder="currentFolder"
          :hasImages="images.length > 0"
          @wheel="handleWheel"
        />

        <NavigationPanel 
          :currentIndex="currentIndex"
          :totalImages="images.length"
          :isLoading="isLoading"
          :isTreeLoading="isTreeLoading"
          :lastVisitedFolder="lastVisitedFolder"
          :hasNextLeafFolder="hasNextLeafFolder"
          :hasPreviousLeafFolder="hasPreviousLeafFolder"
          :rootPath="rootPath"
          @prevImage="prevImage"
          @nextImage="nextImage"
          @goToLastVisitedFolder="goToLastVisitedFolder"
          @goToPrevLeafFolder="goToPrevLeafFolder"
          @goToNextLeafFolder="goToNextLeafFolder"
          @goToRandomFolder="goToRandomFolder"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: var(--background);
  color: var(--foreground);
}

.main-content {
  flex-grow: 1;
  display: flex;
  overflow: hidden;
}

.container {
  display: flex;
  flex-direction: column;
  width: 100%;
  background-color: var(--background);
  color: var(--foreground);
  font-family: sans-serif;
}
</style>

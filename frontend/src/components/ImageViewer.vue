<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { SelectFolder, ListImages, ReadImage } from '../../wailsjs/go/main/App';
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
const lastVisitedFolder = ref<string>("");

const { 
  zoomIn, zoomOut, resetZoom, imageStyle 
} = useImageZoom();

const { 
  folderTreeRoot, flatFolderList, leafFolderList, leafFolderSet, flatFolderIndexMap, leafFolderIndexMap, isTreeLoading, treeError,
  loadFolderTree, clearTree 
} = useFolderTree();

const { 
  images, currentIndex, currentImageSrc, isLoading, currentFilename, 
  hasNextLeafFolder, hasPreviousLeafFolder, preloadedImageSrc, preloadedIndex, 
  preloadedFolder, preloadNextImage, clearPreload, displayCurrentImage, loadImagesForPath 
} = useImageGallery(currentFolder, flatFolderList, leafFolderList, leafFolderSet, flatFolderIndexMap, leafFolderIndexMap);

const { 
  slideshowActive, slideshowDelay, slideshowMode, toggleSlideshow, stopSlideshow 
} = useSlideshow(
  nextImage,
  displayRandomImageInFolder,
  displayRandomImageInAllFolders,
  canStartSlideshow
);

function canStartSlideshow(mode: SlideshowMode): boolean {
  if (mode === 'sequence' && images.value.length < 1 && flatFolderList.value.length < 2) return false;
  if (mode === 'random-folder' && images.value.length < 1) return false;
  if (mode === 'random-all' && leafFolderList.value.length < 1) return false;
  return true;
}

async function selectFolder() {
  clearTree();
  images.value = [];
  currentIndex.value = -1;
  currentImageSrc.value = "";
  currentFolder.value = "";

  try {
    const selectedPath = await SelectFolder();
    if (selectedPath) {
      await loadImagesForPath(selectedPath, resetZoom);
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

async function displayRandomImageInAllFolders() {
  if (leafFolderList.value.length === 0) return;
  try {
    const randomFolderPath = leafFolderList.value[Math.floor(Math.random() * leafFolderList.value.length)];
    const randomFolderImages = await ListImages(randomFolderPath);

    if (randomFolderImages.length > 0) {
      const randomImageIndex = Math.floor(Math.random() * randomFolderImages.length);
      currentFolder.value = randomFolderPath;
      images.value = randomFolderImages;
      currentIndex.value = randomImageIndex;
      currentImageSrc.value = await ReadImage(randomFolderImages[randomImageIndex]);
      resetZoom();
    }
  } catch (err) {
    LogError(`Error during random-all slideshow: ${err}`);
  }
}

async function nextImage() {
  if (images.value.length === 0 && flatFolderList.value.length === 0) return;

  let targetIdx = currentIndex.value + 1;
  let targetFolder = currentFolder.value;
  let isMovingFolder = false;

  if (images.value.length === 0 || targetIdx >= images.value.length) {
    const currentFolderIndexInFlatList = flatFolderIndexMap.value.get(currentFolder.value) ?? -1;
    let nextLeafFolderFound = false;

    if (currentFolderIndexInFlatList !== -1) {
      for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
        const potentialNextFolder = flatFolderList.value[i];
        if (leafFolderSet.value.has(potentialNextFolder)) {
          targetFolder = potentialNextFolder;
          targetIdx = 0;
          isMovingFolder = true;
          nextLeafFolderFound = true;
          break;
        }
      }
    }

    if (!nextLeafFolderFound) {
      stopSlideshow();
      return;
    }
  }

  if (isMovingFolder && targetFolder === preloadedFolder.value && targetIdx === preloadedIndex.value && preloadedImageSrc.value) {
    currentImageSrc.value = preloadedImageSrc.value;
    currentFolder.value = targetFolder;
    images.value = await ListImages(targetFolder);
    currentIndex.value = targetIdx;
    resetZoom();
    clearPreload();
    preloadNextImage(slideshowActive.value, slideshowMode.value);
  } else if (isMovingFolder) {
    await handleFolderSelected(targetFolder);
  } else if (!isMovingFolder && slideshowMode.value === 'sequence' && targetIdx === preloadedIndex.value && targetFolder === preloadedFolder.value && preloadedImageSrc.value) {
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
    const currentLeafIndex = leafFolderIndexMap.value.get(currentFolder.value) ?? -1;
    if (currentLeafIndex > 0) {
      const prevLeafFolderPath = leafFolderList.value[currentLeafIndex - 1];
      await loadImagesForPath(prevLeafFolderPath, resetZoom);
      if (images.value.length > 0) {
        currentIndex.value = images.value.length - 1;
        await displayCurrentImage(resetZoom);
      }
      return;
    }
    if (images.value.length > 0) {
      targetIdx = images.value.length - 1;
    } else {
      return;
    }
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

function goToRandomFolder() {
  if (leafFolderList.value.length === 0 || isTreeLoading.value) return;
  let randomFolder;
  if (leafFolderList.value.length > 1) {
    do {
      randomFolder = leafFolderList.value[Math.floor(Math.random() * leafFolderList.value.length)];
    } while (randomFolder === currentFolder.value);
  } else {
    randomFolder = leafFolderList.value[0];
  }
  if (randomFolder) handleFolderSelected(randomFolder);
}

function goToNextLeafFolder() {
  const idx = leafFolderIndexMap.value.get(currentFolder.value) ?? -1;
  if (idx !== -1 && idx + 1 < leafFolderList.value.length) {
    handleFolderSelected(leafFolderList.value[idx + 1]);
  } else if (idx === -1 && leafFolderList.value.length > 0) {
    handleFolderSelected(leafFolderList.value[0]);
  }
}

function goToPrevLeafFolder() {
  const idx = leafFolderIndexMap.value.get(currentFolder.value) ?? -1;
  if (idx > 0) {
    handleFolderSelected(leafFolderList.value[idx - 1]);
  } else if (idx === -1 && leafFolderList.value.length > 0) {
    handleFolderSelected(leafFolderList.value[leafFolderList.value.length - 1]);
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
          :hasLeafFolders="leafFolderList.length > 0"
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
          :leafFolderListLength="leafFolderList.length"
          :hasNextLeafFolder="hasNextLeafFolder"
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

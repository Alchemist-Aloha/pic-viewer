<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, computed, CSSProperties, watch } from 'vue';
import { SelectFolder, ListImages, ReadImage, ListSubfolders } from '../../wailsjs/go/main/App';
import { main as models } from '../../wailsjs/go/models';
import { LogError } from '../../wailsjs/runtime/runtime';
import FolderTree from './FolderTree.vue';

const currentFolder = ref<string>("");
const images = ref<string[]>([]);
const currentIndex = ref<number>(-1);
const currentImageSrc = ref<string>("");
const isLoading = ref<boolean>(false);
const isTreeLoading = ref<boolean>(false);
const slideshowActive = ref<boolean>(false);
const slideshowInterval = ref<number | null>(null);
const slideshowDelay = ref<number>(3000);
const zoomLevel = ref<number>(1);
const minZoom = 0.2;
const maxZoom = 5;
const zoomStep = 0.1;
const folderTreeRoot = ref<models.Folder | null>(null); // Use namespaced type
const flatFolderList = ref<string[]>([]); // Add state for flattened folder list
const treeError = ref<string>(""); // State for tree loading errors
const preloadedImageSrc = ref<string>(""); // State for preloaded image data
const preloadedIndex = ref<number>(-1); // State for the index of the preloaded image
const preloadedFolder = ref<string>(""); // State for the folder of the preloaded image

// Function to get the parent directory path
// Basic implementation, might need refinement for edge cases (root drives)
function getParentDirectory(path: string): string {
    // Handle potential trailing slashes
    const cleanedPath = path.replace(/[\/]$/, '');
    const lastSeparatorIndex = Math.max(cleanedPath.lastIndexOf('/'), cleanedPath.lastIndexOf('\\'));
    if (lastSeparatorIndex === -1) {
        return ""; // Or handle root case appropriately
    }
    // Handle root case like C:\ or /
    if (lastSeparatorIndex === 0 || (cleanedPath.length > 1 && cleanedPath[lastSeparatorIndex - 1] === ':')) {
        return cleanedPath.substring(0, lastSeparatorIndex + 1);
    }
    return cleanedPath.substring(0, lastSeparatorIndex);
}

async function selectFolder() {
  treeError.value = "";
  folderTreeRoot.value = null;
  try {
    const selectedPath = await SelectFolder();
    if (selectedPath) {
      await handleFolderSelected(selectedPath);
    }
  } catch (err) {
    LogError("Error selecting folder: " + err);
    console.error("Error selecting folder:", err);
    treeError.value = "Failed to select folder.";
  }
}

// Renamed loadImages to reflect it loads images for a specific path
async function loadImagesForPath(folderPath: string) {
  if (!folderPath) return;
  isLoading.value = true;
  stopSlideshow();
  zoomLevel.value = 1;
  currentFolder.value = folderPath; // Update current folder state
  preloadedImageSrc.value = ""; // Clear preload
  preloadedIndex.value = -1;
  preloadedFolder.value = "";
  try {
    images.value = await ListImages(folderPath);
    currentIndex.value = images.value.length > 0 ? 0 : -1;
    await displayCurrentImage();
  } catch (err) {
    LogError("Error listing images: " + err);
    console.error("Error listing images:", err);
    images.value = [];
    currentIndex.value = -1;
    currentImageSrc.value = "";
  } finally {
    isLoading.value = false;
  }
}

// Helper function to flatten the tree (depth-first)
function flattenTree(node: models.Folder | null): string[] {
    if (!node) return [];
    let paths: string[] = [node.path];
    if (node.children) {
        for (const child of node.children) {
            paths = paths.concat(flattenTree(child));
        }
    }
    return paths;
}

// Function to load the folder tree structure
async function loadFolderTree(basePath: string) {
    if (!basePath) return;
    isTreeLoading.value = true;
    treeError.value = "";
    folderTreeRoot.value = null;
    flatFolderList.value = []; // Clear the flat list
    console.log("Loading tree for:", basePath);
    try {
        // Ensure the return type matches the namespaced type
        const tree = await ListSubfolders(basePath);
        folderTreeRoot.value = tree;
        flatFolderList.value = flattenTree(tree); // Update the flat list
        console.log("Tree loaded:", folderTreeRoot.value);
        console.log("Flat folder list:", flatFolderList.value);
    } catch (err: any) {
        LogError("Error loading folder tree: " + err);
        console.error("Error loading folder tree:", err);
        treeError.value = `Failed to load folder tree: ${err.message || err}`;
    } finally {
        isTreeLoading.value = false;
    }
}

// New function to handle folder selection from both button and tree
async function handleFolderSelected(selectedPath: string) {
    if (!selectedPath) return; // Don't proceed if path is empty

    // Only reload tree if the root is different or not set
    const parentPath = getParentDirectory(selectedPath);
    const needsTreeLoad = parentPath && (!folderTreeRoot.value || folderTreeRoot.value.path !== parentPath);

    // Avoid reloading images if the folder is already selected AND the tree doesn't need reloading
    // This prevents unnecessary reloads when clicking the already selected folder in the tree
    if (selectedPath === currentFolder.value && !needsTreeLoad) {
        console.log("Folder already selected and tree is current, skipping reload.");
        return;
    }

    // Load images for the selected folder *before* potentially loading the tree
    // This makes the UI feel more responsive if the tree load is slow
    await loadImagesForPath(selectedPath);

    // Load or update the folder tree if necessary
    if (needsTreeLoad) {
        await loadFolderTree(parentPath);
    } else if (!folderTreeRoot.value && parentPath) {
        // Handle case where a folder was selected but tree wasn't loaded initially
         await loadFolderTree(parentPath);
    } else {
         // Ensure currentFolder is updated even if tree doesn't reload
         // This might be redundant if loadImagesForPath already sets it, but safe to keep
         currentFolder.value = selectedPath;
    }
    preloadedImageSrc.value = ""; // Clear preload
    preloadedIndex.value = -1;
    preloadedFolder.value = "";
}

// Handler for the event emitted by FolderTree component
function handleFolderSelectedFromTree(path: string) {
    handleFolderSelected(path);
}

// Function to preload the next image
async function preloadNextImage() {
  preloadedImageSrc.value = ""; // Clear previous preload
  preloadedIndex.value = -1;
  preloadedFolder.value = "";

  if (images.value.length < 1) return; // No images to preload

  let nextIdx = currentIndex.value + 1;
  let nextFolder = currentFolder.value;
  let nextImagePath = "";

  if (nextIdx >= images.value.length) {
    // Need to potentially move to the next folder
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    if (currentFolderIndexInFlatList !== -1 && currentFolderIndexInFlatList < flatFolderList.value.length - 1) {
      // There is a next folder in the flattened list
      nextFolder = flatFolderList.value[currentFolderIndexInFlatList + 1];
      try {
        // Need to list images in the next folder to get the first one
        const nextFolderImages = await ListImages(nextFolder);
        if (nextFolderImages.length > 0) {
          nextImagePath = nextFolderImages[0];
          nextIdx = 0; // Index within the next folder
        } else {
          // Next folder is empty, can't preload
          return;
        }
      } catch (err) {
        LogError(`Error listing images in next folder for preload: ${err}`);
        return; // Can't preload if listing fails
      }
    } else {
      // Last image of the last folder, nothing more to preload
      return;
    }
  } else {
    // Next image is within the current folder
    nextImagePath = images.value[nextIdx];
  }

  if (!nextImagePath) return; // Safety check

  // console.log(`Preloading: Folder=${nextFolder}, Index=${nextIdx}, Path=${nextImagePath}`);
  try {
    const data = await ReadImage(nextImagePath);
    preloadedImageSrc.value = data;
    preloadedIndex.value = nextIdx;
    preloadedFolder.value = nextFolder;
    // console.log(`Preloaded successfully: Index=${nextIdx} in Folder=${nextFolder}`);
  } catch (err) {
    LogError(`Error preloading image ${nextImagePath}: ${err}`);
    // Clear preload state on error
    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";
  }
}

async function displayCurrentImage() {
  if (currentIndex.value >= 0 && currentIndex.value < images.value.length) {
    isLoading.value = true;
    zoomLevel.value = 1; // Reset zoom on image change
    try {
      const imagePath = images.value[currentIndex.value];
      currentImageSrc.value = await ReadImage(imagePath);
      // Trigger preload after current image is displayed
      preloadNextImage(); // Call preload here
    } catch (err) {
      LogError("Error reading image: " + err);
      console.error("Error reading image:", err);
      currentImageSrc.value = ""; // Clear image on error
      // Clear preload if current image fails to load
      preloadedImageSrc.value = "";
      preloadedIndex.value = -1;
      preloadedFolder.value = "";
    } finally {
      isLoading.value = false;
    }
  } else {
    currentImageSrc.value = ""; // No image to display
    zoomLevel.value = 1; // Reset zoom
  }
}

function nextImage() {
  if (images.value.length === 0) return;

  let targetIdx = currentIndex.value + 1;
  let targetFolder = currentFolder.value;
  let isMovingFolder = false;

  if (targetIdx >= images.value.length) {
    // Check if moving to next folder is possible
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    if (currentFolderIndexInFlatList !== -1 && currentFolderIndexInFlatList < flatFolderList.value.length - 1) {
      targetFolder = flatFolderList.value[currentFolderIndexInFlatList + 1];
      targetIdx = 0; // Index in the new folder
      isMovingFolder = true;
    } else {
      console.log("Last image of the last folder reached.");
      return; // Stop at the end
    }
  }

  // Check if the target image is the one we preloaded
  if (!isMovingFolder && targetIdx === preloadedIndex.value && targetFolder === preloadedFolder.value && preloadedImageSrc.value) {
    // console.log(`Using preloaded image: Index=${targetIdx}`);
    isLoading.value = true; // Briefly set loading to prevent flickering/double actions
    currentImageSrc.value = preloadedImageSrc.value;
    currentIndex.value = targetIdx;
    zoomLevel.value = 1; // Reset zoom

    // Clear the used preload data
    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";

    // Trigger the next preload
    preloadNextImage();
    isLoading.value = false; // Loading finished

  } else if (isMovingFolder) {
    // Moving to a new folder, handle selection which includes loading and preloading
    handleFolderSelected(targetFolder);
  } else {
    // Preloaded image not available or not matching, load normally
    // console.log(`Loading image normally: Index=${targetIdx}`);
    currentIndex.value = targetIdx;
    displayCurrentImage(); // This will load the image and trigger the next preload
  }
}

function prevImage() {
  if (images.value.length === 0) return;
  // Clear preload when going back
  preloadedImageSrc.value = "";
  preloadedIndex.value = -1;
  preloadedFolder.value = "";

  currentIndex.value = (currentIndex.value - 1 + images.value.length) % images.value.length;
  displayCurrentImage(); // This will trigger preload for the *next* image relative to the new current one
}

function startSlideshow() {
  if (slideshowActive.value || images.value.length < 2) return;
  slideshowActive.value = true;
  slideshowInterval.value = window.setInterval(() => {
    nextImage();
  }, slideshowDelay.value);
}

function stopSlideshow() {
  if (slideshowInterval.value !== null) {
    clearInterval(slideshowInterval.value);
    slideshowInterval.value = null;
  }
  slideshowActive.value = false;
}

function toggleSlideshow() {
  if (slideshowActive.value) {
    stopSlideshow();
  } else {
    startSlideshow();
  }
}

function zoomIn() {
  zoomLevel.value = Math.min(maxZoom, zoomLevel.value + zoomStep);
}

function zoomOut() {
  zoomLevel.value = Math.max(minZoom, zoomLevel.value - zoomStep);
}

function handleWheel(event: WheelEvent) {
  event.preventDefault(); // Prevent default page scroll

  if (event.ctrlKey) {
    // Ctrl + Scroll = Zoom
    if (!currentImageSrc.value) return; // Don't zoom if no image
    if (event.deltaY < 0) {
      zoomIn();
    } else if (event.deltaY > 0) {
      zoomOut();
    }
  } else {
    // Scroll = Next/Previous Image
    if (isLoading.value || images.value.length < 2) return; // Don't change image if loading or only one image
    if (event.deltaY < 0) {
      // Wheel up
      prevImage();
    } else if (event.deltaY > 0) {
      // Wheel down
      nextImage();
    }
  }
}

// Keyboard navigation & zoom
function handleKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowRight':
      nextImage();
      break;
    case 'ArrowLeft':
      prevImage();
      break;
    case ' ':
      toggleSlideshow();
      event.preventDefault();
      break;
    case '+':
    case '=': // Handle '+' on layouts where it shares key with '='
      zoomIn();
      event.preventDefault();
      break;
    case '-':
      zoomOut();
      event.preventDefault();
      break;
  }
}

// Computed style for the image
const imageStyle = computed<CSSProperties>(() => ({
  transform: `scale(${zoomLevel.value})`,
  maxWidth: '100%', // Keep these to ensure initial fit
  maxHeight: '100%',
  objectFit: 'contain',
  display: 'block',
  transition: 'transform 0.1s ease-out', // Smooth zoom transition
}));

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
  // Wheel listener added dynamically in template if needed, or here:
  // Note: Attaching wheel listener directly to image-container in template is often better
});

onUnmounted(() => {
  stopSlideshow();
  window.removeEventListener('keydown', handleKeydown);
});

</script>

<template>
  <div class="layout-container">
    <!-- Sidebar for Folder Tree -->
    <div class="sidebar">
      <div class="sidebar-header">
        <button @click="selectFolder" title="Select Root Folder for Tree">Browse...</button>
      </div>
      <div v-if="isTreeLoading" class="loading-tree">Loading Tree...</div>
      <div v-else-if="treeError" class="tree-error">{{ treeError }}</div>
      <div v-else-if="folderTreeRoot" class="folder-tree-container">
        <FolderTree 
          :folder="folderTreeRoot" 
          :selectedPath="currentFolder"
          @folder-selected="handleFolderSelectedFromTree" />
      </div>
      <div v-else class="no-tree">
        Click 'Browse' to select a root folder.
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="main-content">
      <div class="container">
        <div class="controls">
          <!-- Folder path display moved here or could be removed if tree is primary -->
          <span v-if="currentFolder" class="folder-path">Current: {{ currentFolder }}</span>
          <span v-else>No folder selected</span>
        </div>

        <div class="image-viewer" @wheel="handleWheel">
          <div v-if="isLoading" class="loading">Loading Image...</div>
          <img v-else-if="currentImageSrc"
               :src="currentImageSrc"
               alt="Current Image"
               :style="imageStyle" />
          <div v-else-if="currentFolder && images.length === 0 && !isLoading" class="no-images">
            No images found in this folder.
          </div>
          <div v-else class="no-folder">
            Select a folder from the tree.
          </div>
        </div>

        <div class="navigation" v-if="images.length > 0">
          <button @click="prevImage" :disabled="isLoading || images.length < 2">Previous</button>
          <span>{{ currentIndex + 1 }} / {{ images.length }}</span>
          <button @click="nextImage" :disabled="isLoading || images.length < 2">Next</button>
          <button @click="toggleSlideshow" :disabled="isLoading || images.length < 2">
            {{ slideshowActive ? 'Stop Slideshow' : 'Start Slideshow' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden; /* Prevent body scroll */
}

.sidebar {
  width: 250px; /* Adjust width as needed */
  flex-shrink: 0;
  background-color: #333740; /* Slightly different shade for sidebar */
  border-right: 1px solid #444;
  display: flex;
  flex-direction: column;
  overflow-y: auto; /* Allow scrolling if tree is long */
  color: #ccc;
}

.sidebar-header {
    padding: 10px;
    border-bottom: 1px solid #444;
}

.sidebar-header button {
    width: 100%;
}

.loading-tree,
.tree-error,
.no-tree {
    padding: 15px;
    text-align: center;
    color: #aaa;
}

.tree-error {
    color: #ff8a8a;
}

.folder-tree-container {
    flex-grow: 1;
    padding: 5px 0; /* Add some padding around the tree */
    overflow-y: auto; /* Ensure tree itself can scroll if needed */
}

.main-content {
  flex-grow: 1;
  display: flex; /* Use flex to make container fill space */
  overflow: hidden; /* Prevent overflow issues */
}

/* Existing container styles adapted */
.container {
  display: flex;
  flex-direction: column;
  /* height: 100vh; Remove fixed height, let it fill main-content */
  width: 100%; /* Fill the main-content area */
  background-color: #282c34;
  color: white;
  font-family: sans-serif;
}

.controls {
  padding: 10px;
  background-color: #3c4049;
  border-bottom: 1px solid #555;
  display: flex;
  align-items: center;
  gap: 15px;
  flex-shrink: 0; /* Prevent controls from shrinking */
}

.folder-path {
  font-size: 0.9em;
  color: #ccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-grow: 1; /* Allow path to take available space */
  min-width: 0; /* Important for text-overflow in flex */
}

.image-viewer {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: auto;
  position: relative;
  cursor: grab;
  min-height: 0; /* Important for flex-grow in column layout */
}
.image-viewer:active {
    cursor: grabbing;
}

.loading,
.no-images,
.no-folder {
  font-size: 1.2em;
  color: #aaa;
  padding: 20px;
}

.image-viewer img {
  /* Styles are primarily handled by the computed imageStyle */
}

.navigation {
  padding: 10px;
  background-color: #3c4049;
  border-top: 1px solid #555;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 15px;
  flex-shrink: 0; /* Prevent navigation from shrinking */
}

button {
  padding: 8px 15px;
  background-color: #61dafb;
  border: none;
  border-radius: 4px;
  color: #282c34;
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s ease;
}

button:hover:not(:disabled) {
  background-color: #4fa8c5;
}

button:disabled {
  background-color: #555;
  color: #999;
  cursor: not-allowed;
}

/* Ensure FolderTree component styles are applied (they are scoped) */

</style>

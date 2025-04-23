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
const leafFolderList = ref<string[]>([]); // Add state for leaf folders only
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
  // Clear previous tree and images immediately for better UX
  folderTreeRoot.value = null;
  images.value = [];
  currentIndex.value = -1;
  currentImageSrc.value = "";
  currentFolder.value = "";
  flatFolderList.value = [];
  isLoading.value = false; // Ensure loading state is reset
  isTreeLoading.value = false; // Ensure tree loading state is reset

  try {
    const selectedPath = await SelectFolder();
    if (selectedPath) {
      // When selecting initially, load both images and the tree
      await loadImagesForPath(selectedPath); // Load images first
      // Load the tree starting from the selected path itself
      await loadFolderTree(selectedPath);
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

// Helper function to get only leaf folders (depth-first)
function getLeafFolders(node: models.Folder | null): string[] {
    if (!node) return [];
    let leafPaths: string[] = [];
    // If a node has no children or an empty children array, it's a leaf
    if (!node.children || node.children.length === 0) {
        leafPaths.push(node.path);
    } else {
        // Otherwise, recurse into children
        for (const child of node.children) {
            leafPaths = leafPaths.concat(getLeafFolders(child));
        }
    }
    return leafPaths;
}

// Function to load the folder tree structure
async function loadFolderTree(basePath: string) {
    if (!basePath) return;
    isTreeLoading.value = true;
    treeError.value = "";
    folderTreeRoot.value = null;
    flatFolderList.value = []; // Clear the flat list
    leafFolderList.value = []; // Clear the leaf list
    console.log("Loading tree for:", basePath);
    try {
        // Ensure the return type matches the namespaced type
        const tree = await ListSubfolders(basePath);
        folderTreeRoot.value = tree;
        flatFolderList.value = flattenTree(tree); // Update the flat list
        leafFolderList.value = getLeafFolders(tree); // Update the leaf list
        console.log("Tree loaded:", folderTreeRoot.value);
        console.log("Flat folder list:", flatFolderList.value);
        console.log("Leaf folder list:", leafFolderList.value);
    } catch (err: any) {
        LogError("Error loading folder tree: " + err);
        console.error("Error loading folder tree:", err);
        treeError.value = `Failed to load folder tree: ${err.message || err}`;
    } finally {
        isTreeLoading.value = false;
    }
}

// Modified function to handle folder selection - ONLY loads images now
async function handleFolderSelected(selectedPath: string) {
    if (!selectedPath) return; // Don't proceed if path is empty

    // Avoid reloading images if the folder is already selected
    if (selectedPath === currentFolder.value) {
        console.log("Folder already selected, skipping image reload.");
        return;
    }

    // Only load images for the selected folder
    await loadImagesForPath(selectedPath);

    // Clear preload state when changing folders this way too
    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";
}

// Handler for the event emitted by FolderTree component
function handleFolderSelectedFromTree(path: string) {
    // This now calls the simplified handler that only loads images
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

// Function to navigate to a random folder from the leaf list
function goToRandomFolder() {
  // Use leafFolderList instead of flatFolderList
  if (leafFolderList.value.length === 0 || isTreeLoading.value) return;

  let randomIndex;
  let randomFolder;
  // Ensure we don't pick the current folder if possible and if there are other options
  if (leafFolderList.value.length > 1) {
    do {
      randomIndex = Math.floor(Math.random() * leafFolderList.value.length);
      randomFolder = leafFolderList.value[randomIndex];
    } while (randomFolder === currentFolder.value);
  } else {
    // Only one leaf folder, just go to it (or stay if it's the current one)
    randomIndex = 0;
    randomFolder = leafFolderList.value[randomIndex];
  }

  if (randomFolder) {
    handleFolderSelected(randomFolder);
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

        <div class="navigation" v-if="images.length > 0 || flatFolderList.length > 0">
          <button @click="prevImage" :disabled="isLoading || images.length < 2">Previous</button>
          <span v-if="images.length > 0">{{ currentIndex + 1 }} / {{ images.length }}</span>
          <button @click="nextImage" :disabled="isLoading || images.length < 2">Next</button>
          <button @click="toggleSlideshow" :disabled="isLoading || images.length < 2">
            {{ slideshowActive ? 'Stop Slideshow' : 'Start Slideshow' }}
          </button>
          <button @click="goToRandomFolder" :disabled="isTreeLoading || leafFolderList.length === 0" title="Go to a random leaf folder in the tree">
            Random Folder
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Dracula Color Palette */
:root {
  --background: #282a36;
  --current-line: #44475a;
  --foreground: #f8f8f2;
  --comment: #6272a4;
  --cyan: #8be9fd;
  --green: #50fa7b;
  --orange: #ffb86c;
  --pink: #ff79c6;
  --purple: #bd93f9;
  --red: #ff5555;
  --yellow: #f1fa8c;
}

/* Move variables outside :root to be accessible within scoped component */
.layout-container {
  --background: #282a36;
  --current-line: #44475a;
  --foreground: #f8f8f2;
  --comment: #6272a4;
  --cyan: #8be9fd;
  --green: #50fa7b;
  --orange: #ffb86c;
  --pink: #ff79c6;
  --purple: #bd93f9;
  --red: #ff5555;
  --yellow: #f1fa8c;
  
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden; /* Prevent body scroll */
  background-color: var(--background);
  color: var(--foreground);
}

.sidebar {
  width: 250px; /* Adjust width as needed */
  flex-shrink: 0;
  background-color: var(--background); /* Darker background for sidebar */
  border-right: 1px solid var(--current-line); /* Use variable */
  display: flex;
  flex-direction: column;
  overflow-y: auto; /* Allow scrolling if tree is long */
  color: var(--foreground); /* Use variable */
}

.sidebar-header {
    padding: 10px;
    border-bottom: 1px solid var(--current-line); /* Use variable */
}

.sidebar-header button {
    width: 100%;
}

.loading-tree,
.tree-error,
.no-tree {
    padding: 15px;
    text-align: center;
    color: var(--comment); /* Use variable */
}

.tree-error {
    color: var(--red); /* Use variable */
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
  background-color: var(--background); /* Use variable */
  color: var(--foreground); /* Use variable */
  font-family: sans-serif;
}

.controls {
  padding: 10px;
  background-color: var(--current-line); /* Use variable */
  border-bottom: 1px solid var(--background); /* Use variable */
  display: flex;
  align-items: center;
  gap: 15px;
  flex-shrink: 0; /* Prevent controls from shrinking */
}

.folder-path {
  font-size: 0.9em;
  color: var(--foreground); /* Use variable */
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
  background-color: var(--background); /* Ensure viewer bg matches */
}
.image-viewer:active {
    cursor: grabbing;
}

.loading,
.no-images,
.no-folder {
  font-size: 1.2em;
  color: var(--comment); /* Use variable */
  padding: 20px;
}

.image-viewer img {
  /* Styles are primarily handled by the computed imageStyle */
}

.navigation {
  padding: 10px;
  background-color: var(--current-line); /* Use variable */
  border-top: 1px solid var(--background); /* Use variable */
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px; /* Adjusted gap slightly for potentially more buttons */
  flex-shrink: 0; /* Prevent navigation from shrinking */
}

button {
  padding: 8px 15px;
  background-color: var(--purple); /* Use variable */
  border: none;
  border-radius: 4px;
  color: var(--foreground); /* Use variable */
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s ease;
}

button:hover:not(:disabled) {
  background-color: var(--pink); /* Use variable for hover */
}

button:disabled {
  background-color: var(--comment); /* Use comment color for background */
  color: var(--background); /* Use background color for text for better contrast */
  cursor: not-allowed;
}

/* Ensure FolderTree component styles are applied (they are scoped) */

</style>

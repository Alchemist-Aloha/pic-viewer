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
const slideshowDelay = ref<number>(3000); // Default delay
type SlideshowMode = 'sequence' | 'random-folder' | 'random-all';
const slideshowMode = ref<SlideshowMode>('sequence'); // Default mode
const zoomLevel = ref<number>(1);
const minZoom = 0.2;
const maxZoom = 5;
const zoomStep = 0.1;
const folderTreeRoot = ref<models.Folder | null>(null); // Use namespaced type
const flatFolderList = ref<string[]>([]); // Add state for flattened folder list
const leafFolderList = ref<string[]>([]); // Add state for leaf folders only
const treeError = ref<string>(""); // State for tree loading errors
const lastVisitedFolder = ref<string>(""); // State for the previously visited folder
const preloadedImageSrc = ref<string>(""); // State for preloaded image data
const preloadedIndex = ref<number>(-1); // State for the index of the preloaded image
const preloadedFolder = ref<string>(""); // State for the folder of the preloaded image

// Computed property for the current filename
const currentFilename = computed(() => {
  if (currentIndex.value >= 0 && images.value.length > 0 && currentIndex.value < images.value.length) {
    const fullPath = images.value[currentIndex.value];
    // Extract filename from path (works for both / and \)
    return fullPath.substring(Math.max(fullPath.lastIndexOf('/'), fullPath.lastIndexOf('\\')) + 1);
  }
  return ""; // Return empty string if no image is selected/loaded
});

// Computed property to check if a subsequent leaf folder exists
const hasNextLeafFolder = computed(() => {
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    // If current folder isn't found or flat list is empty, assume no next folder
    if (currentFolderIndexInFlatList === -1 || flatFolderList.value.length === 0) {
        return false;
    }

    // Search for the *next* leaf folder in the flat list
    for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
        const potentialNextFolder = flatFolderList.value[i];
        if (leafFolderList.value.includes(potentialNextFolder)) {
            return true; // Found a subsequent leaf folder
        }
    }

    // If the loop finishes without finding a next leaf folder
    return false; // No next leaf folder found
});

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

    // Store the current folder as the last visited *before* changing it
    // Make sure we don't store an empty string if it's the first selection
    if (currentFolder.value && currentFolder.value !== selectedPath) {
        lastVisitedFolder.value = currentFolder.value;
        console.log("Last visited folder set to:", lastVisitedFolder.value);
    }

    // Only load images for the selected folder
    await loadImagesForPath(selectedPath); // This updates currentFolder.value internally

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

// Function to navigate back to the last visited folder
function goToLastVisitedFolder() {
    if (!lastVisitedFolder.value || isTreeLoading.value) {
        console.log("No last visited folder stored or tree is loading.");
        return;
    }
    // Navigate back. Note: This will update lastVisitedFolder again with the *current* folder.
    handleFolderSelected(lastVisitedFolder.value);
}

// Function to preload the next image (conditionally disabled during random slideshow)
async function preloadNextImage() {
  // Disable preloading if a random slideshow is active
  if (slideshowActive.value && slideshowMode.value !== 'sequence') {
      preloadedImageSrc.value = "";
      preloadedIndex.value = -1;
      preloadedFolder.value = "";
      return;
  }

  preloadedImageSrc.value = ""; // Clear previous preload
  preloadedIndex.value = -1;
  preloadedFolder.value = "";

  if (images.value.length < 1 && flatFolderList.value.length < 1) return; // Nothing to preload if no images and no other folders

  let nextIdx = currentIndex.value + 1;
  let nextFolder = currentFolder.value;
  let nextImagePath = "";

  // Check if we need to move to the next folder
  if (images.value.length === 0 || nextIdx >= images.value.length) {
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    let nextLeafFolderFound = false;

    if (currentFolderIndexInFlatList !== -1) {
        // Search for the next leaf folder in the flat list
        for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
            const potentialNextFolder = flatFolderList.value[i];
            if (leafFolderList.value.includes(potentialNextFolder)) {
                nextFolder = potentialNextFolder;
                nextLeafFolderFound = true;
                break; // Found the next leaf folder
            }
        }
    }

    if (!nextLeafFolderFound) {
        // console.log("Preload: No subsequent leaf folder found in flat list.");
        return; // Last image of the last leaf folder (or no leaf folders after current)
    }

    // Found the next leaf folder, try to get its first image
    try {
      const nextFolderImages = await ListImages(nextFolder);
      if (nextFolderImages.length > 0) {
        nextImagePath = nextFolderImages[0];
        nextIdx = 0; // Index within the next folder
      } else {
        // Next leaf folder is empty, can't preload from it
        // console.log(`Preload: Next leaf folder ${nextFolder} is empty.`);
        return;
      }
    } catch (err) {
      LogError(`Error listing images in next leaf folder ${nextFolder} for preload: ${err}`);
      return; // Can't preload if listing fails
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
      // Trigger preload after current image is displayed (respecting slideshow mode)
      preloadNextImage(); // Call preload here (it checks internally)
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
    // Attempt to preload if we landed here from an empty folder during sequence
    if (slideshowMode.value === 'sequence') {
        preloadNextImage();
    }
  }
}

// Helper for random image within the current folder
function displayRandomImageInFolder() {
    if (images.value.length < 1) return; // Nothing to randomize

    let randomIndex;
    if (images.value.length === 1) {
        randomIndex = 0; // Only one image
    } else {
        // Try to pick a different image than the current one
        do {
            randomIndex = Math.floor(Math.random() * images.value.length);
        } while (randomIndex === currentIndex.value);
    }
    currentIndex.value = randomIndex;
    displayCurrentImage(); // Display the randomly selected image
}

// Helper for random image across all leaf folders
async function displayRandomImageInAllFolders() {
    if (leafFolderList.value.length === 0) return; // No folders to choose from

    try {
        isLoading.value = true; // Show loading indicator during potential folder switch
        const randomFolderPath = leafFolderList.value[Math.floor(Math.random() * leafFolderList.value.length)];
        const randomFolderImages = await ListImages(randomFolderPath);

        if (randomFolderImages.length > 0) {
            const randomImageIndex = Math.floor(Math.random() * randomFolderImages.length);
            const randomImagePath = randomFolderImages[randomImageIndex];

            // Update state regardless of whether the folder changed
            currentFolder.value = randomFolderPath;
            images.value = randomFolderImages;
            currentIndex.value = randomImageIndex;

            // Now display the chosen image
            currentImageSrc.value = await ReadImage(randomImagePath);
            zoomLevel.value = 1; // Reset zoom

        } else {
            // Selected random folder was empty, maybe try again or just skip?
            console.warn(`Randomly selected folder ${randomFolderPath} was empty.`);
            // Optionally, call the function again to try another folder:
            // displayRandomImageInAllFolders();
        }
    } catch (err) {
        LogError(`Error during random-all slideshow: ${err}`);
        console.error("Error during random-all slideshow:", err);
        // Clear image on error
        currentImageSrc.value = "";
        images.value = [];
        currentIndex.value = -1;
    } finally {
        isLoading.value = false;
    }
}


async function nextImage() { // Add async here
  // This function now primarily handles sequential 'next' for manual controls and sequence slideshow
  if (images.value.length === 0 && flatFolderList.value.length === 0) return; // Nothing to navigate

  let targetIdx = currentIndex.value + 1;
  let targetFolder = currentFolder.value;
  let isMovingFolder = false;

  // Handle moving past the end of the current folder's images
  if (images.value.length === 0 || targetIdx >= images.value.length) {
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    let nextLeafFolderFound = false;

    if (currentFolderIndexInFlatList !== -1) {
        // Search for the next leaf folder in the flat list
        for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
            const potentialNextFolder = flatFolderList.value[i];
            if (leafFolderList.value.includes(potentialNextFolder)) {
                targetFolder = potentialNextFolder;
                targetIdx = 0; // Index in the new folder
                isMovingFolder = true;
                nextLeafFolderFound = true;
                break; // Found the next leaf folder
            }
        }
    }

    if (!nextLeafFolderFound) {
        console.log("Last image of the last leaf folder reached (sequence).");
        stopSlideshow(); // Stop slideshow if it was running sequentially
        return; // Do not proceed further
    }
  }

  // Check if the target image is the one we preloaded (only relevant for sequence)
  if (isMovingFolder && targetFolder === preloadedFolder.value && targetIdx === preloadedIndex.value && preloadedImageSrc.value) {
     // Using preloaded image from the *start* of the next leaf folder
    // console.log(`Using preloaded image from next leaf folder: Folder=${targetFolder}, Index=${targetIdx}`);
    isLoading.value = true;
    currentImageSrc.value = preloadedImageSrc.value;
    currentFolder.value = targetFolder; // Update current folder
    // Need to load images for the new folder to update the list and count
    images.value = await ListImages(targetFolder); // Await this!
    currentIndex.value = targetIdx;
    zoomLevel.value = 1;

    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";

    preloadNextImage(); // Trigger preload for the image *after* this one
    isLoading.value = false;

  } else if (isMovingFolder) {
    // Moving to a new leaf folder, but preload didn't match (or wasn't ready)
    // console.log(`Moving to next leaf folder (no preload match): ${targetFolder}`);
    // Use handleFolderSelected which loads images, displays first, and triggers preload
    handleFolderSelected(targetFolder);

  } else if (!isMovingFolder && slideshowMode.value === 'sequence' && targetIdx === preloadedIndex.value && targetFolder === preloadedFolder.value && preloadedImageSrc.value) {
    // Using preloaded image within the *same* folder
    // console.log(`Using preloaded image within same folder: Index=${targetIdx}`);
    isLoading.value = true;
    currentImageSrc.value = preloadedImageSrc.value;
    currentIndex.value = targetIdx;
    zoomLevel.value = 1;

    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";

    preloadNextImage();
    isLoading.value = false;

  } else {
    // Preloaded image not available/matching, or not in sequence mode, or just moving within current folder. Load normally.
    // console.log(`Loading image normally: Index=${targetIdx}`);
    currentIndex.value = targetIdx;
    displayCurrentImage(); // This will load the image and trigger the next preload (if applicable)
  }
}

function prevImage() {
  // Previous always works sequentially within the current folder for simplicity
  if (images.value.length === 0) return;

  // Clear preload when going back manually
  preloadedImageSrc.value = "";
  preloadedIndex.value = -1;
  preloadedFolder.value = "";

  let targetIdx = currentIndex.value - 1;
  if (targetIdx < 0) {
      // TODO: Optionally implement moving to the *last* image of the *previous* folder
      // For now, wrap around within the current folder
      targetIdx = images.value.length - 1;
  }

  currentIndex.value = targetIdx;
  displayCurrentImage(); // This will trigger preload for the *next* image relative to the new current one
}

function startSlideshow() {
  // Validate delay
  if (isNaN(slideshowDelay.value) || slideshowDelay.value < 500) {
      slideshowDelay.value = 500; // Enforce minimum delay
  }

  // Conditions to prevent starting
  if (slideshowActive.value) return;
  if (slideshowMode.value === 'sequence' && images.value.length < 1 && flatFolderList.value.length < 2) return; // Need something to sequence through
  if (slideshowMode.value === 'random-folder' && images.value.length < 1) return;
  if (slideshowMode.value === 'random-all' && leafFolderList.value.length < 1) return;

  // Clear preload when starting random slideshows
  if (slideshowMode.value !== 'sequence') {
      preloadedImageSrc.value = "";
      preloadedIndex.value = -1;
      preloadedFolder.value = "";
  }

  slideshowActive.value = true;
  console.log(`Starting slideshow: Mode=${slideshowMode.value}, Delay=${slideshowDelay.value}`);

  // Initial action based on mode
  switch (slideshowMode.value) {
      case 'sequence':
          nextImage(); // Start with the next sequential image
          break;
      case 'random-folder':
          displayRandomImageInFolder();
          break;
      case 'random-all':
          displayRandomImageInAllFolders();
          break;
  }

  // Set interval for subsequent actions
  slideshowInterval.value = window.setInterval(() => {
    switch (slideshowMode.value) {
      case 'sequence':
        nextImage();
        break;
      case 'random-folder':
        displayRandomImageInFolder();
        break;
      case 'random-all':
        displayRandomImageInAllFolders();
        break;
    }
  }, slideshowDelay.value);
}


function stopSlideshow() {
  if (slideshowInterval.value !== null) {
    clearInterval(slideshowInterval.value);
    slideshowInterval.value = null;
  }
  slideshowActive.value = false;
  console.log("Slideshow stopped.");
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

    // Check for previous image (wheel up)
    if (event.deltaY < 0) {
      // Allow previous if not loading and there are images to go back to
      if (!isLoading.value && images.value.length > 0) {
         prevImage();
      }
    } 
    // Check for next image (wheel down)
    else if (event.deltaY > 0) {
      // Allow next if not loading AND (it's not the last image OR there is a next leaf folder)
      const isLastImageInCurrentFolder = currentIndex.value >= images.value.length - 1;
      if (!isLoading.value && (!isLastImageInCurrentFolder || hasNextLeafFolder.value)) {
         nextImage();
      }
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

// Function to navigate to the next folder in the leaf list
function goToNextLeafFolder() {
    if (leafFolderList.value.length < 2 || isTreeLoading.value) return; // Need at least two leaf folders

    const currentLeafIndex = leafFolderList.value.indexOf(currentFolder.value);

    if (currentLeafIndex === -1) {
        // Current folder is not a leaf folder, maybe go to the first leaf?
        // Or find the next leaf *after* the current non-leaf folder in the flat list?
        // For simplicity, let's just go to the first leaf if current isn't a leaf.
        handleFolderSelected(leafFolderList.value[0]);
        return;
    }

    const nextLeafIndex = currentLeafIndex + 1;

    if (nextLeafIndex < leafFolderList.value.length) {
        // Go to the next leaf folder
        handleFolderSelected(leafFolderList.value[nextLeafIndex]);
    } else {
        // Reached the end of the leaf folder list
        console.log("Last leaf folder reached.");
        // Optionally wrap around: handleFolderSelected(leafFolderList.value[0]);
    }
}

// Function to navigate to the previous folder in the leaf list
function goToPrevLeafFolder() {
    if (leafFolderList.value.length < 2 || isTreeLoading.value) return; // Need at least two leaf folders

    const currentLeafIndex = leafFolderList.value.indexOf(currentFolder.value);

    if (currentLeafIndex === -1) {
        // Current folder is not a leaf folder. Maybe go to the *last* leaf folder?
        // Or find the leaf *before* the current non-leaf folder in the flat list?
        // For simplicity, let's go to the last leaf folder if current isn't a leaf.
        handleFolderSelected(leafFolderList.value[leafFolderList.value.length - 1]);
        return;
    }

    const prevLeafIndex = currentLeafIndex - 1;

    if (prevLeafIndex >= 0) {
        // Go to the previous leaf folder
        handleFolderSelected(leafFolderList.value[prevLeafIndex]);
    } else {
        // Reached the beginning of the leaf folder list
        console.log("First leaf folder reached.");
        // Optionally wrap around: handleFolderSelected(leafFolderList.value[leafFolderList.value.length - 1]);
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
          <!-- Current filename display -->
          <span v-if="currentFilename" class="file-name" :title="currentFilename">File: {{ currentFilename }}</span>
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
          <!-- MODIFIED :disabled condition for Next button -->
          <button 
            @click="nextImage" 
            :disabled="isLoading || (currentIndex >= images.length - 1 && !hasNextLeafFolder)"
            title="Go to the next image or next folder">
            Next
          </button>
          <button @click="goToLastVisitedFolder" :disabled="isTreeLoading || !lastVisitedFolder" title="Go to the previously visited folder">
            Last Visited
          </button>
          <button @click="goToPrevLeafFolder" :disabled="isTreeLoading || leafFolderList.length < 2" title="Go to the previous leaf folder in the tree">
            Prev Folder
          </button>
          <button @click="goToNextLeafFolder" :disabled="isTreeLoading || leafFolderList.length < 2 || !hasNextLeafFolder" title="Go to the next leaf folder in the tree">
            Next Folder
          </button>
          <button @click="goToRandomFolder" :disabled="isTreeLoading || leafFolderList.length === 0" title="Go to a random leaf folder in the tree">
            Random Folder
          </button>
          <!-- Slideshow Controls -->
          <div class="slideshow-controls">
              <button @click="toggleSlideshow"
                      :disabled="isLoading || (slideshowMode === 'sequence' && images.length < 1 && flatFolderList.length < 2) || (slideshowMode === 'random-folder' && images.length < 1) || (slideshowMode === 'random-all' && leafFolderList.length < 1)">
                {{ slideshowActive ? 'Stop' : 'Start' }}
              </button>
              <select v-model="slideshowMode" :disabled="slideshowActive || isLoading">
                <option value="sequence">Sequence</option>
                <option value="random-folder" :disabled="images.length < 1">Random (Folder)</option>
                <option value="random-all" :disabled="leafFolderList.length < 1">Random (All)</option>
              </select>
              <label>
                Delay (ms):
                <input type="number" v-model.number="slideshowDelay" min="500" step="100" :disabled="slideshowActive || isLoading" class="delay-input">
              </label>
          </div>
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
  width: 350px; /* Increased width */
  min-width: 200px; /* Add a minimum width */
  max-width: 50%; /* Add a maximum width relative to viewport */
  resize: horizontal; /* Allow horizontal resizing by the user */
  overflow: auto; /* Needed for resize handle */
  flex-shrink: 0;
  background-color: var(--background); /* Darker background for sidebar */
  border-right: 1px solid var(--current-line); /* Use variable */
  display: flex;
  flex-direction: column;
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

.navigation {
  /* ... existing navigation styles ... */
  gap: 5px; /* Adjust gap if needed */
  flex-wrap: wrap; /* Allow controls to wrap on smaller screens */
}

.slideshow-controls {
    display: flex;
    align-items: center;
    gap: 8px; /* Gap between slideshow elements */
    padding: 0 5px; /* Add some padding */
    border-left: 1px solid var(--background);
    border-right: 1px solid var(--background);
}

.slideshow-controls label {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.9em;
}

.delay-input {
    width: 60px; /* Adjust width as needed */
    padding: 4px;
    background-color: var(--background);
    color: var(--foreground);
    border: 1px solid var(--comment);
    border-radius: 3px;
}
.delay-input:disabled {
    background-color: var(--current-line);
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

select:hover:not(:disabled) {
    background-color: var(--pink);
}

/* Adjust button padding if needed */
button {
  padding: 6px 12px; /* Slightly reduced padding */
}

</style>

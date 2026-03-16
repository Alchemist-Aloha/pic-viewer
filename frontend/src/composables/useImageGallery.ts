import { ref, computed, Ref } from 'vue';
import { ListImages, ReadImage, FindNextFolder, FindPrevFolder } from '../../wailsjs/go/main/App';
import { LogError } from '../../wailsjs/runtime/runtime';

export function useImageGallery(
  currentFolder: Ref<string>,
  rootPath: Ref<string>
) {
  const images = ref<string[]>([]);
  const currentIndex = ref<number>(-1);
  const currentImageSrc = ref<string>("");
  const isLoading = ref<boolean>(false);
  const preloadedImageSrc = ref<string>("");
  const preloadedIndex = ref<number>(-1);
  const preloadedFolder = ref<string>("");

  const currentFilename = computed(() => {
    if (currentIndex.value >= 0 && images.value.length > 0 && currentIndex.value < images.value.length) {
      const fullPath = images.value[currentIndex.value];
      return fullPath.substring(Math.max(fullPath.lastIndexOf('/'), fullPath.lastIndexOf('\\')) + 1);
    }
    return "";
  });

  const hasNextLeafFolder = ref(false);
  const hasPreviousLeafFolder = ref(false);

  async function updateNavigationAvailability() {
    if (!currentFolder.value || !rootPath.value) {
        hasNextLeafFolder.value = false;
        hasPreviousLeafFolder.value = false;
        return;
    }
    try {
        const next = await FindNextFolder(currentFolder.value, rootPath.value);
        hasNextLeafFolder.value = !!next;
        const prev = await FindPrevFolder(currentFolder.value, rootPath.value);
        hasPreviousLeafFolder.value = !!prev;
    } catch (err) {
        LogError("Error updating nav availability: " + err);
    }
  }

  async function preloadNextImage(slideshowActive: boolean, slideshowMode: string) {
    if (slideshowActive && slideshowMode !== 'sequence') {
      clearPreload();
      return;
    }
    clearPreload();

    if (images.value.length < 1 && !hasNextLeafFolder.value) return;

    let nextIdx = currentIndex.value + 1;
    let nextFolder = currentFolder.value;
    let nextImagePath = "";

    if (nextIdx >= images.value.length) {
        try {
            const nextFolderRes = await FindNextFolder(currentFolder.value, rootPath.value);
            if (!nextFolderRes) return;
            nextFolder = nextFolderRes;
            const nextFolderImages = await ListImages(nextFolder);
            if (nextFolderImages.length > 0) {
              nextImagePath = nextFolderImages[0];
              nextIdx = 0;
            } else {
              return;
            }
        } catch (err) {
            LogError("Error preloading next folder: " + err);
            return;
        }
    } else {
      nextImagePath = images.value[nextIdx];
    }

    if (!nextImagePath) return;

    try {
      const data = await ReadImage(nextImagePath);
      preloadedImageSrc.value = data;
      preloadedIndex.value = nextIdx;
      preloadedFolder.value = nextFolder;
    } catch (err) {
      LogError(`Error preloading image ${nextImagePath}: ${err}`);
      clearPreload();
    }
  }

  function clearPreload() {
    preloadedImageSrc.value = "";
    preloadedIndex.value = -1;
    preloadedFolder.value = "";
  }

  async function displayCurrentImage(onImageLoad?: () => void) {
    if (currentIndex.value >= 0 && currentIndex.value < images.value.length) {
      isLoading.value = true;
      if (onImageLoad) onImageLoad(); // Reset zoom, etc.
      try {
        const imagePath = images.value[currentIndex.value];
        currentImageSrc.value = await ReadImage(imagePath);
      } catch (err) {
        LogError("Error reading image: " + err);
        currentImageSrc.value = "";
        clearPreload();
      } finally {
        isLoading.value = false;
      }
    } else {
      currentImageSrc.value = "";
    }
  }

  async function loadImagesForPath(folderPath: string, onImageLoad?: () => void) {
    if (!folderPath) return;
    isLoading.value = true;
    currentFolder.value = folderPath;
    clearPreload();
    try {
      images.value = await ListImages(folderPath);
      currentIndex.value = images.value.length > 0 ? 0 : -1;
      await displayCurrentImage(onImageLoad);
      await updateNavigationAvailability();
      if (currentIndex.value !== -1) {
        preloadNextImage(false, 'sequence');
      }
    } catch (err) {
      LogError("Error listing images: " + err);
      images.value = [];
      currentIndex.value = -1;
      currentImageSrc.value = "";
    } finally {
      isLoading.value = false;
    }
  }

  return {
    images,
    currentIndex,
    currentImageSrc,
    isLoading,
    currentFilename,
    hasNextLeafFolder,
    hasPreviousLeafFolder,
    preloadedImageSrc,
    preloadedIndex,
    preloadedFolder,
    preloadNextImage,
    clearPreload,
    displayCurrentImage,
    loadImagesForPath,
    updateNavigationAvailability
  };
}

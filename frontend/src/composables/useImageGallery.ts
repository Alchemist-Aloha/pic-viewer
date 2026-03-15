import { ref, computed, Ref } from 'vue';
import { ListImages, ReadImage } from '../../wailsjs/go/main/App';
import { LogError } from '../../wailsjs/runtime/runtime';

export function useImageGallery(
  currentFolder: Ref<string>,
  flatFolderList: Ref<string[]>,
  leafFolderList: Ref<string[]>
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

  const hasNextLeafFolder = computed(() => {
    const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
    if (currentFolderIndexInFlatList === -1 || flatFolderList.value.length === 0) {
      return false;
    }
    for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
      const potentialNextFolder = flatFolderList.value[i];
      if (leafFolderList.value.includes(potentialNextFolder)) {
        return true;
      }
    }
    return false;
  });

  const hasPreviousLeafFolder = computed(() => {
    const currentFolderIndexInLeafList = leafFolderList.value.indexOf(currentFolder.value);
    return currentFolderIndexInLeafList > 0;
  });

  async function preloadNextImage(slideshowActive: boolean, slideshowMode: string) {
    if (slideshowActive && slideshowMode !== 'sequence') {
      clearPreload();
      return;
    }
    clearPreload();

    if (images.value.length < 1 && flatFolderList.value.length < 1) return;

    let nextIdx = currentIndex.value + 1;
    let nextFolder = currentFolder.value;
    let nextImagePath = "";

    if (images.value.length === 0 || nextIdx >= images.value.length) {
      const currentFolderIndexInFlatList = flatFolderList.value.indexOf(currentFolder.value);
      let nextLeafFolderFound = false;

      if (currentFolderIndexInFlatList !== -1) {
        for (let i = currentFolderIndexInFlatList + 1; i < flatFolderList.value.length; i++) {
          const potentialNextFolder = flatFolderList.value[i];
          if (leafFolderList.value.includes(potentialNextFolder)) {
            nextFolder = potentialNextFolder;
            nextLeafFolderFound = true;
            break;
          }
        }
      }

      if (!nextLeafFolderFound) return;

      try {
        const nextFolderImages = await ListImages(nextFolder);
        if (nextFolderImages.length > 0) {
          nextImagePath = nextFolderImages[0];
          nextIdx = 0;
        } else {
          return;
        }
      } catch (err) {
        LogError(`Error listing images in next leaf folder ${nextFolder} for preload: ${err}`);
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
  };
}

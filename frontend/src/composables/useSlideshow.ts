import { ref, Ref } from 'vue';

export type SlideshowMode = 'sequence' | 'random-folder' | 'random-all';

export function useSlideshow(
  nextImage: () => Promise<void>,
  displayRandomImageInFolder: () => void,
  displayRandomImageInAllFolders: () => Promise<void>,
  canStart: (mode: SlideshowMode) => boolean
) {
  const slideshowActive = ref<boolean>(false);
  const slideshowInterval = ref<number | null>(null);
  const slideshowDelay = ref<number>(3000);
  const slideshowMode = ref<SlideshowMode>('sequence');

  function startSlideshow() {
    if (isNaN(slideshowDelay.value) || slideshowDelay.value < 500) {
      slideshowDelay.value = 500;
    }

    if (slideshowActive.value) return;
    if (!canStart(slideshowMode.value)) return;

    slideshowActive.value = true;

    // Initial action
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

    slideshowInterval.value = window.setInterval(async () => {
      switch (slideshowMode.value) {
        case 'sequence':
          await nextImage();
          break;
        case 'random-folder':
          displayRandomImageInFolder();
          break;
        case 'random-all':
          await displayRandomImageInAllFolders();
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
  }

  function toggleSlideshow() {
    if (slideshowActive.value) {
      stopSlideshow();
    } else {
      startSlideshow();
    }
  }

  return {
    slideshowActive,
    slideshowDelay,
    slideshowMode,
    startSlideshow,
    stopSlideshow,
    toggleSlideshow,
  };
}

import { ref, computed, CSSProperties } from 'vue';

export function useImageZoom() {
  const zoomLevel = ref<number>(1);
  const minZoom = 0.2;
  const maxZoom = 5;
  const zoomStep = 0.1;

  const zoomIn = () => {
    zoomLevel.value = Math.min(maxZoom, zoomLevel.value + zoomStep);
  };

  const zoomOut = () => {
    zoomLevel.value = Math.max(minZoom, zoomLevel.value - zoomStep);
  };

  const resetZoom = () => {
    zoomLevel.value = 1;
  };

  const imageStyle = computed<CSSProperties>(() => ({
    transform: `scale(${zoomLevel.value})`,
    maxWidth: '100%',
    maxHeight: '100%',
    objectFit: 'contain',
    display: 'block',
    transition: 'transform 0.1s ease-out',
  }));

  return {
    zoomLevel,
    zoomIn,
    zoomOut,
    resetZoom,
    imageStyle,
  };
}

# README

## About
pic-viewer is a simple image viewer built with Vue.js and Go using the Wails framework. It aims to provide a fast and lightweight image viewing experience, especially for navigating large collections of images organized in nested folders.

## Features
- **Fast Image Loading:** Leverages Go for backend processing and efficient image reading.
- **Folder Tree Navigation:** Displays a hierarchical view of folders starting from a selected root.
    - Sidebar panel is resizable.
    - Clicking a folder loads its images.
- **Sequential Navigation:**
    - Move between images within the current folder (Next/Previous buttons, Arrow Keys, Mouse Wheel).
    - Automatically advances to the *first image* of the *next leaf folder* when reaching the end of the current folder.
- **Folder Navigation:**
    - Buttons to jump to the Next/Previous *leaf* folder in the tree order.
    - Button to jump to a Random *leaf* folder.
    - Button to return to the Last Visited folder.
- **Slideshow Modes:**
    - **Sequence:** Advances through images sequentially, moving between leaf folders as needed.
    - **Random (Folder):** Shows random images from the currently selected folder.
    - **Random (All):** Shows random images from any *leaf* folder within the loaded tree.
    - Configurable delay (in milliseconds).
- **Zoom:**
    - Zoom in/out using Ctrl + Mouse Wheel
- **Image Preloading:** Preloads the next sequential image for a smoother viewing experience (disabled during random slideshows).
- **Format Support:** Displays common image formats and extracts embedded JPEGs from Fuji RAF raw files.

## Supported Formats
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- WebP (.webp)
- Fuji RAW (.raf) - *Displays the embedded JPEG preview*

## Controls

### Mouse
- **Scroll Wheel:** Next/Previous image.
- **Ctrl + Scroll Wheel:** Zoom In/Out.
- **Click Folder in Tree:** Load images from that folder.
- **Click Buttons:** Perform corresponding actions (Browse, Navigation, Slideshow).

### Keyboard
- **Arrow Right:** Next image / Next folder (sequential leaf).
- **Arrow Left:** Previous image.
- **Spacebar:** Start/Stop slideshow.
- **`+` / `=`:** Zoom In.
- **`-`:** Zoom Out.

## Building from Source
1. Ensure you have Go, Node.js, npm, and the Wails CLI installed.
2. Clone the repository.
3. Navigate to the project directory in your terminal.
4. Run `wails build`.
5. The executable will be in the `build/bin` directory.

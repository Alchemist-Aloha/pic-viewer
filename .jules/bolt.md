## 2024-03-12 - ListImages WalkDir vs ReadDir
**Learning:** `filepath.WalkDir` was traversing all subdirectories even when `ListImages` only wanted the immediate files in the current directory. This is extremely slow for deep or large directory structures (like a photo album root folder with thousands of nested directories/files).
**Action:** Always verify if a recursive walk is necessary. If only immediate children are needed, `os.ReadDir` is significantly faster and uses less memory.

## 2024-05-14 - Redundant Image Decoding Anti-Pattern
**Learning:** Found a major performance anti-pattern where files were being manually decoded into uncompressed memory (`image.Decode`) and then re-encoded to PNG (`png.Encode`) before being sent to the frontend via base64. Because modern browsers natively support rendering image formats like JPG, PNG, and WebP, this decoding/encoding step is a complete waste of CPU cycles and memory, and it makes the final base64 string substantially larger.
**Action:** When sending common web-supported image files (like .jpg, .png, .webp) to a frontend via base64, directly read the file bytes and encode them, skipping any explicit image decoding/encoding steps in the backend, unless specific image manipulation (resizing, cropping) is strictly required.

## 2024-06-25 - Array.includes() Performance Bottleneck in Vue Computed/Methods
**Learning:** Found a major performance bottleneck where `.includes()` was used on large arrays (`leafFolderList`) within loops and frequently called methods/computed properties. In Vue reactivity, operations like this on large reactive arrays block the main thread and cause noticeable lag. A benchmark (`frontend/benchmark.mjs`) showed that `Set.has()` is ~31x faster for directory traversal lookups compared to `Array.includes()`.
**Action:** Always pre-calculate a `Set` (often using a `computed` property) when doing frequent membership checks (`.includes()`) against a static or slowly-changing collection, especially within loops or reactive getters.

## 2024-03-19 - Skip Redundant Image Decoding And Re-encoding
**Learning:** Found a major performance anti-pattern where files were being manually decoded into uncompressed memory (`image.Decode`) and then re-encoded to PNG (`png.Encode`) before being sent to the frontend via base64. Because modern browsers natively support rendering image formats like JPG, PNG, and WebP, this decoding/encoding step is a complete waste of CPU cycles and memory. The benchmark results improved performance by almost 45x and significantly cut string memory consumption during base64 payload assembly by using `unsafe.String(unsafe.SliceData(buf), len(buf))`.
**Action:** When sending common web-supported image files (like .jpg, .png, .webp) to a frontend via base64, directly read the file bytes and encode them. Furthermore, make sure to minimize base64 payload copies using `unsafe.String(unsafe.SliceData(buf), len(buf))` during the generation process.

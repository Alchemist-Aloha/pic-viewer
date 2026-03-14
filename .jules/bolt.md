## 2024-03-12 - ListImages WalkDir vs ReadDir
**Learning:** `filepath.WalkDir` was traversing all subdirectories even when `ListImages` only wanted the immediate files in the current directory. This is extremely slow for deep or large directory structures (like a photo album root folder with thousands of nested directories/files).
**Action:** Always verify if a recursive walk is necessary. If only immediate children are needed, `os.ReadDir` is significantly faster and uses less memory.

## 2024-05-14 - Redundant Image Decoding Anti-Pattern
**Learning:** Found a major performance anti-pattern where files were being manually decoded into uncompressed memory (`image.Decode`) and then re-encoded to PNG (`png.Encode`) before being sent to the frontend via base64. Because modern browsers natively support rendering image formats like JPG, PNG, and WebP, this decoding/encoding step is a complete waste of CPU cycles and memory, and it makes the final base64 string substantially larger.
**Action:** When sending common web-supported image files (like .jpg, .png, .webp) to a frontend via base64, directly read the file bytes and encode them, skipping any explicit image decoding/encoding steps in the backend, unless specific image manipulation (resizing, cropping) is strictly required.

## 2024-05-15 - Base64 Encoding Memory Allocation
**Learning:** Using `base64.StdEncoding.EncodeToString` combined with `fmt.Sprintf` to prepend a MIME type prefix for Data URIs causes extreme memory allocations. For a large image file, `fmt.Sprintf` allocates another completely new string that is the sum of the already large base64 encoded string and the prefix, causing excessive memory consumption and garbage collection load.
**Action:** When creating large Data URIs, manually calculate the total string length, allocate a single `[]byte` slice of that exact size, copy the prefix in first, and directly encode the data into the remaining slice space using `base64.StdEncoding.Encode`. Finally, cast the entire slice to a string to return it. This drastically reduces allocations and CPU time.

## 2024-03-12 - ListImages WalkDir vs ReadDir
**Learning:** `filepath.WalkDir` was traversing all subdirectories even when `ListImages` only wanted the immediate files in the current directory. This is extremely slow for deep or large directory structures (like a photo album root folder with thousands of nested directories/files).
**Action:** Always verify if a recursive walk is necessary. If only immediate children are needed, `os.ReadDir` is significantly faster and uses less memory.

## 2024-05-14 - Redundant Image Decoding Anti-Pattern
**Learning:** Found a major performance anti-pattern where files were being manually decoded into uncompressed memory (`image.Decode`) and then re-encoded to PNG (`png.Encode`) before being sent to the frontend via base64. Because modern browsers natively support rendering image formats like JPG, PNG, and WebP, this decoding/encoding step is a complete waste of CPU cycles and memory, and it makes the final base64 string substantially larger.
**Action:** When sending common web-supported image files (like .jpg, .png, .webp) to a frontend via base64, directly read the file bytes and encode them, skipping any explicit image decoding/encoding steps in the backend, unless specific image manipulation (resizing, cropping) is strictly required.

## 2025-03-03 - Avoid base64.EncodeToString and fmt.Sprintf for large payloads
**Learning:** `base64.StdEncoding.EncodeToString` combined with `fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)` results in double string allocation and unnecessary copying of large payloads (like megabytes of image data).
**Action:** When creating Data URI base64 strings in Go, especially for large files, pre-allocate a single `[]byte` buffer with the exact required length (`len(prefix) + base64.EncodedLen`). Copy the prefix directly into the buffer, and use `base64.StdEncoding.Encode(buf[len(prefix):], data)` to encode the data directly into the remaining space, avoiding intermediate string allocations. Convert back to string efficiently.

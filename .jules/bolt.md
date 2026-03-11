## 2024-03-12 - ListImages WalkDir vs ReadDir
**Learning:** `filepath.WalkDir` was traversing all subdirectories even when `ListImages` only wanted the immediate files in the current directory. This is extremely slow for deep or large directory structures (like a photo album root folder with thousands of nested directories/files).
**Action:** Always verify if a recursive walk is necessary. If only immediate children are needed, `os.ReadDir` is significantly faster and uses less memory.

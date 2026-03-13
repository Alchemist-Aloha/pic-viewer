package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// Import local raw package
	"pic-viewer/raw"

	"github.com/facette/natsort"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
// Keep Greet function for now, might remove later if unused
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// SelectFolder opens a directory selection dialog
func (a *App) SelectFolder() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Picture Folder",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

// ListImages returns a list of image file paths in a directory
func (a *App) ListImages(dirPath string) ([]string, error) {
	var imageFiles []string
	validExtensions := map[string]bool{
		// Standard formats
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		// Fuji RAW format handled by local package
		".raf": true,
	}

	// ⚡ Bolt: Use os.ReadDir instead of filepath.WalkDir to only read immediate children.
	// This avoids an O(N) traversal of all nested subdirectories, massively improving performance
	// for deep or large directory structures when we only want files in the current folder.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, d := range entries {
		if d.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if validExtensions[ext] {
			imageFiles = append(imageFiles, filepath.Join(dirPath, d.Name()))
		}
	}

	sort.Strings(imageFiles)
	return imageFiles, nil
}

// ReadImage reads an image file (including HDR and RAF) and returns its base64 encoded content
func (a *App) ReadImage(filePath string) (encodedImage string, err error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle RAF files using the local raw package
	if ext == ".raf" {
		// Use panic recovery in case raw.ReadRAF panics
		defer func() {
			if r := recover(); r != nil {
				// Log the panic and set the named return error
				errMsg := fmt.Sprintf("panic occurred while decoding RAF file %s: %v", filePath, r)
				runtime.LogError(a.ctx, errMsg)
				// Set the named error return value
				encodedImage = ""
				err = fmt.Errorf("%s", errMsg)
			}
		}()

		// Call the potentially panicking function
		rafData := raw.ReadRAF(filePath)

		// Check if an error was set by the recover() block above
		if err != nil {
			return // Return immediately if panic occurred and err was set
		}

		if rafData == nil || len(rafData.Jpeg) == 0 {
			err = fmt.Errorf("failed to extract JPEG from RAF file %s", filePath)
			return // Return the error
		}
		// Encode the extracted JPEG data directly
		encoded := base64.StdEncoding.EncodeToString(rafData.Jpeg)
		encodedImage = fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
		return // Return success (encodedImage, nil)
	}

	// Handle other formats
	// ⚡ Bolt: Read the file directly and base64 encode it instead of
	// decoding to image.Image and re-encoding to PNG. This massively
	// reduces memory usage and CPU time, and keeps the payload small.
	var data []byte
	data, err = os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to read file %s: %w", filePath, err)
		return
	}

	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".bmp":
		mimeType = "image/bmp"
	case ".webp":
		mimeType = "image/webp"
	default:
		// Log unsupported format
		runtime.LogWarningf(a.ctx, "Unsupported format '%s' encountered for file %s", ext, filePath)
		mimeType = "application/octet-stream"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	encodedImage = fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	err = nil
	return
}

// Folder represents a directory in the tree view
type Folder struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Children []*Folder `json:"children,omitempty"`
}

// ListSubfolders recursively lists subdirectories for the tree view
func (a *App) ListSubfolders(basePath string) (*Folder, error) {
	rootInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat base path '%s': %w", basePath, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", basePath)
	}

	root := &Folder{
		Name: filepath.Base(basePath),
		Path: basePath,
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		// Don't fail completely, just log and return what we have (maybe permissions issue)
		runtime.LogError(a.ctx, fmt.Sprintf("Error reading directory %s: %v", basePath, err))
		return root, nil // Return the root even if reading contents failed
	}

	var children []*Folder
	for _, entry := range entries {
		if entry.IsDir() {
			// Ignore hidden directories (like .git, .vscode etc.)
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childPath := filepath.Join(basePath, entry.Name())
			childNode, err := a.ListSubfolders(childPath) // Recursive call
			if err != nil {
				// Log error for the specific subdirectory but continue with others
				runtime.LogError(a.ctx, fmt.Sprintf("Error processing subdirectory %s: %v", childPath, err))
				// Optionally add a placeholder node indicating the error
				// children = append(children, &Folder{Name: entry.Name() + " (Error)", Path: childPath})
				continue
			}
			if childNode != nil { // Ensure we got a valid node back
				children = append(children, childNode)
			}
		}
	}

	// Sort children alphabetically by name using natural sort.
	// ⚡ Bolt: Pre-calculate lowercased names to avoid O(N log N) string
	// allocations and conversions during the sort comparison phase.
	type sortNode struct {
		folder *Folder
		lower  string
	}
	nodes := make([]sortNode, len(children))
	for i, c := range children {
		nodes[i] = sortNode{folder: c, lower: strings.ToLower(c.Name)}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return natsort.Compare(nodes[i].lower, nodes[j].lower)
	})

	for i, n := range nodes {
		children[i] = n.folder
	}

	root.Children = children
	return root, nil
}

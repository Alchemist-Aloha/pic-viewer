package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		// Add other extensions if needed
	}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Only process files directly within the selected directory (not subdirs)
		if !d.IsDir() && filepath.Dir(path) == dirPath {
			ext := strings.ToLower(filepath.Ext(path))
			if validExtensions[ext] {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return imageFiles, nil
}

// ReadImage reads an image file and returns its base64 encoded content
func (a *App) ReadImage(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Determine mime type (optional but good for data URI)
	// For simplicity, we'll rely on browser detection or common types
	// A more robust solution would use net/http.DetectContentType
	var mimeType string
	ext := strings.ToLower(filepath.Ext(filePath))
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
		mimeType = "application/octet-stream" // Fallback
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
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

	// Sort children alphabetically by name
	sort.Slice(children, func(i, j int) bool {
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})

	root.Children = children
	return root, nil
}

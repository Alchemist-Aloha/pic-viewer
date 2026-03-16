package main

import (
	"context"

	"pic-viewer/internal/fs"
	"pic-viewer/internal/image"

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
	return fs.ListImages(dirPath)
}

// ReadImage reads an image file and returns its base64 encoded content
func (a *App) ReadImage(filePath string) (string, error) {
	return image.ReadImage(filePath)
}

// ListSubfolders lists immediate subdirectories for the tree view (non-recursive)
func (a *App) ListSubfolders(basePath string) ([]*fs.Folder, error) {
	return fs.ListSubfolders(basePath)
}

// GetFolderInfo returns a single Folder struct for the given path
func (a *App) GetFolderInfo(path string) (*fs.Folder, error) {
	return fs.GetFolderInfo(path)
}

// FindNextFolder finds the next folder with images in a DFS traversal
func (a *App) FindNextFolder(currentPath string, rootPath string) (string, error) {
	return fs.FindNextFolder(currentPath, rootPath)
}

// FindPrevFolder finds the previous folder with images in a DFS traversal
func (a *App) FindPrevFolder(currentPath string, rootPath string) (string, error) {
	return fs.FindPrevFolder(currentPath, rootPath)
}

package fs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/facette/natsort"
)

// Folder represents a directory in the tree view
type Folder struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Children    []*Folder `json:"children,omitempty"`
	HasChildren bool      `json:"hasChildren"`
	HasImages   bool      `json:"hasImages"`
}

var validExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".raf":  true,
}

// HasImages checks if a directory contains any supported image files
func HasImages(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if validExtensions[ext] {
				return true
			}
		}
	}
	return false
}

// ListImages returns a list of image file paths in a directory
func ListImages(dirPath string) ([]string, error) {
	var imageFiles []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, d := range entries {
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if validExtensions[ext] {
				imageFiles = append(imageFiles, filepath.Join(dirPath, d.Name()))
			}
		}
	}

	sort.Slice(imageFiles, func(i, j int) bool {
		return natsort.Compare(imageFiles[i], imageFiles[j])
	})

	return imageFiles, nil
}

// ListSubfolders lists immediate subdirectories for the tree view (non-recursive)
func ListSubfolders(basePath string) ([]*Folder, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var folders []*Folder
	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childPath := filepath.Join(basePath, entry.Name())

			// Check if this child has its own subdirectories
			hasChildren := false
			subEntries, _ := os.ReadDir(childPath)
			hasImages := false
			for _, subEntry := range subEntries {
				if subEntry.IsDir() && !strings.HasPrefix(subEntry.Name(), ".") {
					hasChildren = true
				} else if !subEntry.IsDir() {
					ext := strings.ToLower(filepath.Ext(subEntry.Name()))
					if validExtensions[ext] {
						hasImages = true
					}
				}
				if hasChildren && hasImages {
					break
				}
			}

			folders = append(folders, &Folder{
				Name:        entry.Name(),
				Path:        childPath,
				HasChildren: hasChildren,
				HasImages:   hasImages,
			})
		}
	}

	sort.Slice(folders, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(folders[i].Name), strings.ToLower(folders[j].Name))
	})

	return folders, nil
}

// FindNextFolder finds the next folder with images in a DFS traversal
func FindNextFolder(currentPath string, rootPath string) (string, error) {
	// 1. Check subfolders of currentPath
	next, found := findFirstFolderWithImages(currentPath)
	if found {
		return next, nil
	}

	// 2. Move up and find next sibling
	curr := currentPath
	for curr != "" && curr != filepath.Dir(rootPath) && curr != rootPath {
		parent := filepath.Dir(curr)
		siblings, err := os.ReadDir(parent)
		if err != nil {
			break
		}

		foundCurr := false
		for _, sibling := range siblings {
			if !sibling.IsDir() || strings.HasPrefix(sibling.Name(), ".") {
				continue
			}
			siblingPath := filepath.Join(parent, sibling.Name())
			if foundCurr {
				// Search this sibling's subtree
				if HasImages(siblingPath) {
					return siblingPath, nil
				}
				next, found := findFirstFolderWithImages(siblingPath)
				if found {
					return next, nil
				}
			}
			if siblingPath == curr {
				foundCurr = true
			}
		}
		curr = parent
		if curr == rootPath {
			break
		}
	}

	return "", nil
}

// FindPrevFolder finds the previous folder with images in a DFS traversal
func FindPrevFolder(currentPath string, rootPath string) (string, error) {
	curr := currentPath
	for curr != "" && curr != filepath.Dir(rootPath) && curr != rootPath {
		parent := filepath.Dir(curr)
		siblings, err := os.ReadDir(parent)
		if err != nil {
			break
		}

		// Find current index
		currIdx := -1
		var siblingDirs []string
		for _, sibling := range siblings {
			if sibling.IsDir() && !strings.HasPrefix(sibling.Name(), ".") {
				siblingPath := filepath.Join(parent, sibling.Name())
				siblingDirs = append(siblingDirs, siblingPath)
				if siblingPath == curr {
					currIdx = len(siblingDirs) - 1
				}
			}
		}

		if currIdx > 0 {
			// Check previous siblings' subtrees from right to left
			for i := currIdx - 1; i >= 0; i-- {
				siblingPath := siblingDirs[i]
				next, found := findLastFolderWithImages(siblingPath)
				if found {
					return next, nil
				}
				if HasImages(siblingPath) {
					return siblingPath, nil
				}
			}
		}

		// If we reach the parent, check if the parent has images
		if HasImages(parent) && parent != filepath.Dir(rootPath) {
			return parent, nil
		}
		curr = parent
		if curr == rootPath {
			break
		}
	}
	return "", nil
}

func findFirstFolderWithImages(path string) (string, bool) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", false
	}

	var subdirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subdirs = append(subdirs, filepath.Join(path, entry.Name()))
		}
	}
	sort.Slice(subdirs, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(subdirs[i]), strings.ToLower(subdirs[j]))
	})

	for _, subdir := range subdirs {
		if HasImages(subdir) {
			return subdir, true
		}
		if res, found := findFirstFolderWithImages(subdir); found {
			return res, true
		}
	}
	return "", false
}

func findLastFolderWithImages(path string) (string, bool) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", false
	}

	var subdirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subdirs = append(subdirs, filepath.Join(path, entry.Name()))
		}
	}
	sort.Slice(subdirs, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(subdirs[i]), strings.ToLower(subdirs[j]))
	})

	for i := len(subdirs) - 1; i >= 0; i-- {
		subdir := subdirs[i]
		if res, found := findLastFolderWithImages(subdir); found {
			return res, true
		}
		if HasImages(subdir) {
			return subdir, true
		}
	}
	return "", false
}

// GetFolderInfo returns a single Folder struct for the given path
func GetFolderInfo(path string) (*Folder, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}

	hasChildren := false
	hasImages := false
	entries, _ := os.ReadDir(path)
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			hasChildren = true
		} else if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if validExtensions[ext] {
				hasImages = true
			}
		}
		if hasChildren && hasImages {
			break
		}
	}

	return &Folder{
		Name:        filepath.Base(path),
		Path:        path,
		HasChildren: hasChildren,
		HasImages:   hasImages,
	}, nil
}

// FindRandomFolderWithImages finds a random folder containing images under rootPath
func FindRandomFolderWithImages(rootPath string) (string, error) {
	var foldersWithImages []string

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories but don't skip the root itself if it's "hidden" (though root usually isn't)
			if strings.HasPrefix(d.Name(), ".") && path != rootPath {
				return filepath.SkipDir
			}
			if HasImages(path) {
				foldersWithImages = append(foldersWithImages, path)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(foldersWithImages) == 0 {
		return "", nil
	}

	// Seed random for each call to ensure variety
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return foldersWithImages[r.Intn(len(foldersWithImages))], nil
}

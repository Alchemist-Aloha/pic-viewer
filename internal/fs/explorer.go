package fs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/facette/natsort"
)

// Folder represents a directory in the tree view
type Folder struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Children []*Folder `json:"children,omitempty"`
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

// ListImages returns a list of image file paths in a directory
func ListImages(dirPath string) ([]string, error) {
	var imageFiles []string

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
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
	sort.Strings(imageFiles)
	return imageFiles, nil
}

// ListSubfolders recursively lists subdirectories for the tree view
func ListSubfolders(basePath string) (*Folder, error) {
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
		return root, err // Return the root even if reading contents failed
	}

	var children []*Folder
	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childPath := filepath.Join(basePath, entry.Name())
			childNode, err := ListSubfolders(childPath)
			if err != nil {
				continue
			}
			if childNode != nil {
				children = append(children, childNode)
			}
		}
	}

	sort.Slice(children, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(children[i].Name), strings.ToLower(children[j].Name))
	})

	root.Children = children
	return root, nil
}

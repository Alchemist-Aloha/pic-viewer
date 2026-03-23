package image

import (
	"strings"
	"testing"
	"os"
	"path/filepath"
)

func TestReadImageWebSupported(t *testing.T) {
	// Create a temp dummy image
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.jpg")
	dummyData := []byte("dummy jpeg data")
	err := os.WriteFile(filePath, dummyData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ReadImage(filePath)
	if err != nil {
		t.Fatalf("ReadImage failed: %v", err)
	}

	if !strings.HasPrefix(result, "data:image/jpeg;base64,") {
		t.Errorf("Expected prefix 'data:image/jpeg;base64,', got: %s", result)
	}
}

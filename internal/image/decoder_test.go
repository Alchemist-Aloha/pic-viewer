package image

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestReadImage(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "image_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy image file (just text content for test purposes)
	dummyContent := []byte("dummy image content")
	dummyEncoded := base64.StdEncoding.EncodeToString(dummyContent)

	testCases := []struct {
		ext      string
		mimeType string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".bmp", "image/bmp"},
		{".webp", "image/webp"},
		{".unknown", "application/octet-stream"},
	}

	for _, tc := range testCases {
		t.Run(tc.ext, func(t *testing.T) {
			filePath := filepath.Join(tempDir, "test_image"+tc.ext)
			err := os.WriteFile(filePath, dummyContent, 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Expected data URI format
			expected := "data:" + tc.mimeType + ";base64," + dummyEncoded

			// Act
			result, err := ReadImage(filePath)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	}
}

func TestEncodeDataURI(t *testing.T) {
	data := []byte("hello world")
	mimeType := "text/plain"
	encoded := base64.StdEncoding.EncodeToString(data)
	expected := "data:" + mimeType + ";base64," + encoded

	result := encodeDataURI(mimeType, data)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

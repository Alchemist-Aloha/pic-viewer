package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

// ReadImage reads an image file (including HDR and RAF) and returns its base64 encoded content
func ReadImage(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle RAF files
	if ext == ".raf" {
		rafData, err := ReadRAF(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read RAF file: %w", err)
		}
		if rafData == nil || len(rafData.Jpeg) == 0 {
			return "", fmt.Errorf("failed to extract JPEG from RAF file")
		}
		encoded := base64.StdEncoding.EncodeToString(rafData.Jpeg)
		return fmt.Sprintf("data:image/jpeg;base64,%s", encoded), nil
	}

	// Read file bytes directly, skipping any explicit image decoding/encoding steps
	// for common formats to prevent high memory allocations and GC pressure.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
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
		// Attempt to fallback to decoding for unsupported formats
		file, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		img, _, decodeErr := image.Decode(file)
		if decodeErr != nil {
			mimeType = "application/octet-stream"
		} else {
			// For successfully decoded fallback images, encode as PNG
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				return "", fmt.Errorf("failed to encode image to PNG: %w", err)
			}

			// Re-use the optimization for the fallback
			data = buf.Bytes()
			mimeType = "image/png"
		}
	}

	// ⚡ Bolt: Fast base64 encoding using pre-calculation and zero-copy string conversion
	prefix := "data:" + mimeType + ";base64,"
	prefixLen := len(prefix)
	encodedLen := base64.StdEncoding.EncodedLen(len(data))

	buf := make([]byte, prefixLen+encodedLen)
	copy(buf, prefix)
	base64.StdEncoding.Encode(buf[prefixLen:], data)

	return unsafe.String(unsafe.SliceData(buf), len(buf)), nil
}

// Note: Logging should be handled by the caller or a dedicated logging service.
// The format name %s (from image.Decode) could be logged if needed.

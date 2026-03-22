package image

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

// encodeDataURI directly encodes byte slice into a base64 data URI
// using a single allocation and unsafe.String for zero-copy conversion.
func encodeDataURI(mimeType string, data []byte) string {
	prefix := "data:" + mimeType + ";base64,"
	encodedLen := base64.StdEncoding.EncodedLen(len(data))
	buf := make([]byte, len(prefix)+encodedLen)

	copy(buf, prefix)
	base64.StdEncoding.Encode(buf[len(prefix):], data)

	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

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
		return encodeDataURI("image/jpeg", rafData.Jpeg), nil
	}

	// Handle standard web formats directly without decoding
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
		mimeType = "application/octet-stream"
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return encodeDataURI(mimeType, data), nil
}

// Note: Logging should be handled by the caller or a dedicated logging service.

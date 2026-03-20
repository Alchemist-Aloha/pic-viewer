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

// encodeBase64DataURI efficiently encodes bytes to a base64 Data URI string.
// It pre-allocates a single buffer and uses unsafe.String for zero-copy conversion,
// preventing high memory allocations and GC pressure.
func encodeBase64DataURI(mimeType string, data []byte) string {
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
		return encodeBase64DataURI("image/jpeg", rafData.Jpeg), nil
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
	}

	// Fast path for web-supported formats: bypass image.Decode entirely
	// to save CPU cycles and memory.
	if mimeType != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return encodeBase64DataURI(mimeType, data), nil
	}

	// Fallback for other formats using image.Decode
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	img, _, decodeErr := image.Decode(file)
	if decodeErr != nil {
		// Fallback for formats not handled by image.Decode
		_, seekErr := file.Seek(0, 0)
		if seekErr != nil {
			return "", fmt.Errorf("failed to seek file after decode error: %w", seekErr)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to decode or read file: decodeErr=%v, readErr=%w", decodeErr, err)
		}

		return encodeBase64DataURI("application/octet-stream", data), nil
	}

	// For successfully decoded images, encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	return encodeBase64DataURI("image/png", buf.Bytes()), nil
}

// Note: Logging should be handled by the caller or a dedicated logging service.
// The format name %s (from image.Decode) could be logged if needed.

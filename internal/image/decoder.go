package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	_ "image/gif"
	_ "image/jpeg"
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
		return encodeBytesToBase64DataURI(rafData.Jpeg, "image/jpeg"), nil
	}

	// For web-supported formats, skip decoding to save CPU/memory and encode directly
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

	if mimeType != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return encodeBytesToBase64DataURI(data, mimeType), nil
	}

	// Handle other formats using image.Decode (fallback)
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

		return encodeBytesToBase64DataURI(data, "application/octet-stream"), nil
	}

	// For successfully decoded images, encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	return encodeBytesToBase64DataURI(buf.Bytes(), "image/png"), nil
}

// encodeBytesToBase64DataURI efficiently creates a base64 Data URI with zero-copy string conversion
func encodeBytesToBase64DataURI(data []byte, mimeType string) string {
	prefix := "data:" + mimeType + ";base64,"
	prefixLen := len(prefix)
	encodedLen := base64.StdEncoding.EncodedLen(len(data))

	// Pre-allocate exact buffer size
	buf := make([]byte, prefixLen+encodedLen)

	// Copy prefix
	copy(buf, prefix)

	// Encode directly into the buffer
	base64.StdEncoding.Encode(buf[prefixLen:], data)

	// Zero-copy conversion to string
	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

// Note: Logging should be handled by the caller or a dedicated logging service.
// The format name %s (from image.Decode) could be logged if needed.

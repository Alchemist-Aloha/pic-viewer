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

// encodeToBase64DataURI encodes raw bytes to a base64 Data URI string efficiently.
func encodeToBase64DataURI(data []byte, mimeType string) string {
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

		return encodeToBase64DataURI(rafData.Jpeg, "image/jpeg"), nil
	}

	// Optimize common web formats - read directly and base64 encode, skipping image.Decode/Encode
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

		return encodeToBase64DataURI(data, mimeType), nil
	}

	// Handle other formats using image.Decode
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

		return encodeToBase64DataURI(data, "application/octet-stream"), nil
	}

	// For successfully decoded images, encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	return encodeToBase64DataURI(buf.Bytes(), "image/png"), nil
}

// Note: Logging should be handled by the caller or a dedicated logging service.
// The format name %s (from image.Decode) could be logged if needed.

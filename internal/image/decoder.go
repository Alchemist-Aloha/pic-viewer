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
		encoded := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
	}

	// For successfully decoded images, encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

// Note: Logging should be handled by the caller or a dedicated logging service.
// The format name %s (from image.Decode) could be logged if needed.

package image

import (
	"encoding/base64"
	"fmt"
	"os"

	"testing"
	"bytes"
	"image"
	"image/png"
	_ "image/gif"
	_ "image/jpeg"
	"image/color"
)

// Create a dummy image for testing
func createDummyImage(tb testing.TB, ext string) string {
	tb.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	tmpfile, err := os.CreateTemp("", "testimage*"+ext)
	if err != nil {
		tb.Fatal(err)
	}
	defer tmpfile.Close()

	if err := png.Encode(tmpfile, img); err != nil {
		tb.Fatal(err)
	}

	return tmpfile.Name()
}

// Old approach
func ReadImageOld(filePath string) (string, error) {


	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, decodeErr := image.Decode(file)
	if decodeErr != nil {
		return "", decodeErr
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

func ReadImageNew(filePath string) (string, error) {
	return ReadImage(filePath)
}

func BenchmarkReadImageOld(b *testing.B) {
	tmpfile := createDummyImage(b, ".png")
	defer os.Remove(tmpfile)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ReadImageOld(tmpfile)
	}
}

func BenchmarkReadImageNew(b *testing.B) {
	tmpfile := createDummyImage(b, ".png")
	defer os.Remove(tmpfile)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ReadImageNew(tmpfile)
	}
}

func TestReadImageNew(t *testing.T) {
	tmpfile := createDummyImage(t, ".png")
	defer os.Remove(tmpfile)

	result, err := ReadImage(tmpfile)
	if err != nil {
		t.Fatalf("ReadImage failed: %v", err)
	}

	if result == "" {
		t.Errorf("Expected non-empty result")
	}

	if result[:22] != "data:image/png;base64," {
		t.Errorf("Expected data URI prefix 'data:image/png;base64,', got %s", result[:22])
	}
}

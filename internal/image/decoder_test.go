package image

import (
	"os"
	"testing"
)

func BenchmarkReadImage(b *testing.B) {
	os.WriteFile("dummy_test.jpg", []byte("fake image data"), 0644)
	defer os.Remove("dummy_test.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadImage("dummy_test.jpg")
	}
}

package image

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

// RAF is the Fuji raw file format
type RAF struct {
	Header struct {
		Magic         [16]byte
		FormatVersion [4]byte
		CameraID      [8]byte
		Camera        [32]byte
		Dir           struct {
			Version [4]byte
			_       [20]byte
			Jpeg    struct {
				IDX int32
				Len int32
			}
		}
	}
	Jpeg []byte
	Exif *exif.Exif
}

// WriteJpeg writes the raw preview jpeg
func (r *RAF) WriteJpeg(w io.Writer) (int, error) {
	return w.Write(r.Jpeg)
}

// ReadRAF makes a new RAF from a file
func ReadRAF(fname string) (*RAF, error) {
	raf := new(RAF)

	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	err = binary.Read(f, binary.BigEndian, &raf.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary header: %w", err)
	}

	jbuf := make([]byte, raf.Header.Dir.Jpeg.Len)
	_, err = f.ReadAt(jbuf, int64(raf.Header.Dir.Jpeg.IDX))
	if err != nil {
		return nil, fmt.Errorf("failed to read JPEG data from RAF: %w", err)
	}
	raf.Jpeg = jbuf

	buf := bytes.NewBuffer(jbuf)
	raf.Exif, err = exif.Decode(buf)
	if err != nil {
		// Log error but don't fail, EXIF is secondary
		// return raf, nil // Or handle it differently
	}

	return raf, nil
}

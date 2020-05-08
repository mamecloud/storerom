package processzip

import (
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"io"
)

type Fingerprint struct {
	size int64
	crc  string
	sha1 string
}

type countingWriter struct {
	size int64
}

func (w *countingWriter) Write(p []byte) (n int, err error) {
	written := len(p)
	w.size += int64(written)
	return written, nil
}

// Fingerprint a file in the Mame way
func fingerprint(source io.Reader, destination io.Writer) (Fingerprint, error) {

	result := Fingerprint{}

	// Prepare destination Writers
	size := new(countingWriter)
	crc := crc32.NewIEEE()
	sha1 := sha1.New()
	destinations := io.MultiWriter(size, crc, sha1, destination)

	// Copy content
	if _, err := io.Copy(destinations, source); err != nil {
		return result, err
	}

	// Collect results
	result.size = size.size
	result.crc = hex.EncodeToString(crc.Sum(nil))
	result.sha1 = hex.EncodeToString(sha1.Sum(nil))

	return result, nil
}


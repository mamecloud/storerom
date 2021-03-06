package processzip

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"hash/crc32"
	"io"
)

// Fingerprint calculates identifiers of a rom in the Mame way
type Fingerprint struct {
	Size int64
	Crc  string
	Sha1 string

	writer io.Writer
	sha1   hash.Hash
	crc    hash.Hash32
}

// Digest calculates the crc and sha1 values and stores them in the struct
func (f *Fingerprint) Digest() {
	f.Crc = hex.EncodeToString(f.crc.Sum(nil))
	f.Sha1 = hex.EncodeToString(f.sha1.Sum(nil))
}

func (f *Fingerprint) Write(p []byte) (int, error) {
	f.Size += int64(len(p))
	return f.writer.Write(p)
}

// FingerprintWriter wraps an io.Writer so that size, crc and sha1 can be computed after writing (using the Digest() function).
func FingerprintWriter() (f *Fingerprint) {

	f = new(Fingerprint)

	// Prepare destination Writers
	f.crc = crc32.NewIEEE()
	f.sha1 = sha1.New()
	f.writer = io.MultiWriter(f.crc, f.sha1)

	return f
}

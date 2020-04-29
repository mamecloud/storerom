package processzip

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

type Fingerprint struct {
	size int64
	crc  string
	sha1 string
}

// Fingerprint a file in the Mame way
func fingerprint(file string) Fingerprint {

	result := Fingerprint{}

	// Open source file
	source, err := os.Open(file)
	if err != nil {
		panic(fmt.Sprintf("Error opening %s for fingerprint: %v\n", file, err))
	}
	defer source.Close()

	// Get the file size
	info, err := source.Stat()
	if err != nil {
		panic(fmt.Sprintf("Error statting file for fingerprint: %v\n", err))
	}
	result.size = info.Size()

	// Calculate hashes
	crc := crc32.NewIEEE()
	sha1 := sha1.New()
	destination := io.MultiWriter(crc, sha1)
	if _, err := io.Copy(destination, source); err != nil {
		panic(fmt.Sprintf("Error hashing file for CRC32/SHA1: %v\n", err))
	}
	result.crc = hex.EncodeToString(crc.Sum(nil))
	result.sha1 = hex.EncodeToString(sha1.Sum(nil))

	return result
}

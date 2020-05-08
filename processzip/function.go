package processzip

import (
	"context"
	"archive/zip"
	"cloud.google.com/go/storage"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"path/filepath"
)

var projectID string = os.Getenv("PROJECT_ID")
var sourceBucket string = fmt.Sprintf("%s-roms-upload", projectID)
var targetBucket string = fmt.Sprintf("%s-roms", projectID)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// ProcessZip processes a zipfile into the rom store.
func ProcessZip(ctx context.Context, m PubSubMessage) error {

	object := string(m.Data)
	client := createClient(ctx)

	zipfile := download(ctx, sourceBucket, object, client)
	processZip(ctx, zipfile, client)

	return nil
}

// Processes all entries in a zip file
func processZip(ctx context.Context, zipfilename string, client *storage.Client) {
	defer os.Remove(zipfilename)

	// Open the zip file
	zipfile, err := zip.OpenReader(zipfilename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input zipfile: %v\n", err))
	}
	defer zipfile.Close()

	// Process the entries
	for _, entry := range zipfile.File {
		fmt.Printf("Entry %s\n", entry.Name)
		if !entry.FileInfo().IsDir() {
			fmt.Printf("Extracting: %s from %s\n", entry.Name, zipfilename)
			filename, fingerprint := processEntry(ctx, entry, client)

			// Upload
			name := filepath.Base(filename)
			objectpath := objectpath(name, fingerprint)
			if !exists(ctx, targetBucket, objectpath, client) {
				upload(ctx, filename, targetBucket, objectpath, client)
			}
		}
	}
}

// Process a zip entry into a new zip
func processEntry(ctx context.Context, sourceEntry *zip.File, client *storage.Client) (string, *Fingerprint) {
	fmt.Printf("Processing zip entry %s\n", sourceEntry.Name)

	// Input
	input, err := sourceEntry.Open()
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", sourceEntry.Name, err))
	}
	defer input.Close()

	// Output: file/zip/entry
	name := filepath.Base(sourceEntry.Name)
	tempFile, err := ioutil.TempFile("", name + "_*")
	if err != nil {
		panic(fmt.Sprintf("Error creating temp file: %v\n", err))
	}
	defer tempFile.Close()
	zipfile := zip.NewWriter(tempFile)
	entry, err := zipfile.Create(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to create zip entry %s: %v\n", name, err))
	}
	fingerprint := FingerprintWriter(entry)

	// Copy
	if _, err := io.Copy(fingerprint, input); err != nil {
		panic(fmt.Sprintf("Failed to write and fingerprint zip entry %s: %v\n", name, err))
	}
	fingerprint.Digest()

	// Finish writing the zip file (central directory)
	err = zipfile.Close()
	if err != nil {
		panic(fmt.Sprintf("Failed to finalise zip file: %v\n", err))
	}

	return tempFile.Name(), fingerprint
}
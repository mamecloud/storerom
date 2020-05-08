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
	process(ctx, zipfile, targetBucket, client)

	return nil
}

// Processes all entries in a zip file
func process(ctx context.Context, zipfilename string, bucket string, client *storage.Client) {
	defer os.Remove(zipfilename)

	// Open the zip file
	zipfile, err := zip.OpenReader(zipfilename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input zipfile: %v\n", err))
	}
	defer zipfile.Close()

	// Process the entries
	for _, entry := range zipfile.File {
		if !entry.FileInfo().IsDir() {
			fmt.Printf("Extracting: %s from %s\n", entry.Name, zipfilename)
			processEntry(ctx, entry, bucket, client)
		}
	}
}

// Process a zip entry into a new zip
func processEntry(ctx context.Context, sourceEntry *zip.File, bucket string, client *storage.Client) {
	fmt.Printf("Processing zip entry %s\n", sourceEntry.Name)

	// Input
	input, err := sourceEntry.Open()
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", sourceEntry.Name, err))
	}
	defer input.Close()

	// Output: file/zip/entry
	tempFile, err := ioutil.TempFile(sourceEntry.Name, ".zip")
	if err != nil {
		panic(fmt.Sprintf("Error creating temp file: %v\n", err))
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	zipfile := zip.NewWriter(tempFile)
	name := filepath.Base(sourceEntry.Name)
	entry, err := zipfile.Create(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to create zip entry %s: %v\n", name, err))
	}
	fingerprint := fingerprint(entry)

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

	// Upload
	objectpath := objectpath(name, fingerprint)
	upload(ctx, tempFile.Name(), bucket, objectpath, client)
}
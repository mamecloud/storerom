package processzip

import (
	"archive/zip"
	"bufio"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
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
	defer duration(track(fmt.Sprintf("*processZip %s", filepath.Base(zipfilename))))
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
			objectpath, exists := existsEntry(ctx, entry, client)
			if !exists {
				uploadEntry(ctx, entry, objectpath, client)
			}
		}
	}
}

// Test if an entry exists
// NB in the majority of cases an entry will already exist, so it makes sense not to extract the entry
// and accept duplicate work in those cases where an entry doesn't already exist.
func existsEntry(ctx context.Context, sourceEntry *zip.File, client *storage.Client) (string, bool) {
	name := filepath.Base(sourceEntry.Name)
	defer duration(track(fmt.Sprintf("*existsEntry %s", name)))
	fmt.Printf("Checking if zip entry %s exists\n", sourceEntry.Name)

	// Input
	inputEntry, err := sourceEntry.Open()
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", sourceEntry.Name, err))
	}
	defer inputEntry.Close()
	inputBuffer := bufio.NewReader(inputEntry)

	// Output
	outputFingerprint := FingerprintWriter()

	// Copy
	msg, time := track(fmt.Sprintf("*existsEntry-copy %s", name))
	if _, err := inputBuffer.WriteTo(outputFingerprint); err != nil {
		panic(fmt.Sprintf("Failed to write and fingerprint zip entry %s: %v\n", name, err))
	}
	fmt.Printf("%s: %v bytes\n", name, outputFingerprint.Size)
	duration(msg, time)

	// Result
	outputFingerprint.Digest()
	objectpath := objectpath(name, outputFingerprint)
	exists := exists(ctx, targetBucket, objectpath, client)
	return objectpath, exists
}

// Process a zip entry into a new zip
func uploadEntry(ctx context.Context, sourceEntry *zip.File, objectpath string, client *storage.Client) {
	name := filepath.Base(sourceEntry.Name)
	defer duration(track(fmt.Sprintf("*uploadEntry %s", name)))
	fmt.Printf("Uploading zip entry %s\n", sourceEntry.Name)

	// Input
	inputEntry, err := sourceEntry.Open()
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", sourceEntry.Name, err))
	}
	defer inputEntry.Close()
	inputBuffer := bufio.NewReader(inputEntry)

	// Output: file/zip/entry
	tempFile, err := ioutil.TempFile("", name+"_*")
	if err != nil {
		panic(fmt.Sprintf("Error creating temp file: %v\n", err))
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	outputZipfile := zip.NewWriter(tempFile)
	outputEntry, err := outputZipfile.Create(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to create zip entry %s: %v\n", name, err))
	}
	outputBuffer := bufio.NewWriter(outputEntry)

	// Copy
	msg, time := track(fmt.Sprintf("*uploadEntry-copy %s", name))
	if _, err := inputBuffer.WriteTo(outputBuffer); err != nil {
		panic(fmt.Sprintf("Failed to write and fingerprint zip entry %s: %v\n", name, err))
	}
	if err := outputBuffer.Flush(); err != nil {
		panic(fmt.Sprintf("Failed to flush output buffer for zip entry %s: %v\n", name, err))
	}
	duration(msg, time)

	// Finish writing the zip file (central directory)
	err = outputZipfile.Close()
	if err != nil {
		panic(fmt.Sprintf("Failed to finalise zip file: %v\n", err))
	}

	// Upload
	upload(ctx, tempFile.Name(), targetBucket, objectpath, client)
}

// Thanks to: https://yourbasic.org/golang/measure-execution-time/

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	fmt.Printf("%v: %v\n", msg, time.Since(start))
}

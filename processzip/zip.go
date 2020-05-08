package processzip

import (
	"context"
	"archive/zip"
	"cloud.google.com/go/storage"
	"fmt"
	"os"
	"io"
	"path/filepath"
)

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
	tempFile := tempFile()
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

// // Extracts all entries in a zip file and returns the paths of the extracted files
// func extractAll(zipfilename string) string {
// 	defer os.Remove(zipfilename)

// 	tempDir := tempDir()

// 	fmt.Printf("Unzipping %s\n", zipfilename)
// 	zipfile, err := zip.OpenReader(zipfilename)
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to open input zipfile: %v\n", err))
// 	}
// 	defer zipfile.Close()

// 	for _, entry := range zipfile.File {
// 		fmt.Printf("Extracting: %s\n", entry.Name)
// 		extract(tempDir, entry)
// 	}

// 	return tempDir
// }

// // Extracts a single zip entry
// func extract(directory string, entry *zip.File) {

// 	path := filepath.Join(directory, entry.Name)

// 	// Zip Slip check
// 	if !strings.HasPrefix(path, filepath.Clean(directory)+string(os.PathSeparator)) {
// 		panic(fmt.Sprintf("%s: Illegal file path", path))
// 	}

// 	if entry.FileInfo().IsDir() {
// 		os.MkdirAll(path, entry.Mode())
// 	} else {

// 		// Source data
// 		source, err := entry.Open()
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", entry.Name, err))
// 		}
// 		defer source.Close()

// 		// destination file
// 		os.MkdirAll(filepath.Dir(path), entry.Mode())
// 		destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to open file to extract to: %v\n", err))
// 		}
// 		defer destination.Close()

// 		// Extract
// 		_, err = io.Copy(destination, source)
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to extract entry %s: %v\n", entry.Name, err))
// 		}
// 	}
// }

// func zipFile(filename string) string {
// 	defer os.Remove(filename)

// 	// Source data
// 	source, err := os.Open(filename)
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to open source file: %v\n", err))
// 	}
// 	defer source.Close()

// 	// Destination zip/entry
// 	tempFile := tempFile()
// 	defer tempFile.Close()
// 	zipfile := zip.NewWriter(tempFile)
// 	entryName := filepath.Base(filename)
// 	entry, err := zipfile.Create(entryName)
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to create zip entry %s: %v\n", entryName, err))
// 	}

// 	// Write the data
// 	_, err = io.Copy(entry, source)
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to write zip entry %s: %v\n", entryName, err))
// 	}

// 	// Make sure to check the error on closing the zip.
// 	err = zipfile.Close()
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to close zip file: %v\n", err))
// 	}

// 	return tempFile.Name()
// }

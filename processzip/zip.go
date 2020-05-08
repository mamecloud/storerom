package processzip

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// Processes all entries in a zip file
func process(zipfilename string, bucket string) {
	defer os.Remove(zipfilename)

	// Open the zip file
	fmt.Printf("Unzipping %s\n", zipfilename)
	zipfile, err := zip.OpenReader(zipfilename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input zipfile: %v\n", err))
	}
	defer zipfile.Close()

	// Process the entries
	for _, entry := range zipfile.File {
		if !entry.FileInfo().IsDir() {
			fmt.Printf("Extracting: %s from %s\n", entry.Name, zipfilename)
			processEntry(entry, bucket)
		}
	}
}

// Process a zip entry into a new zip
func processEntry(sourceEntry *zip.File, bucket string) {

	// Source data
	source, err := sourceEntry.Open()
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip entry %s: %v\n", sourceEntry.Name, err))
	}
	defer source.Close()

	// Destination zip/entry
	tempFile := tempFile()
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	zipfile := zip.NewWriter(tempFile)
	entryName := filepath.Base(sourceEntry.Name)
	destination, err := zipfile.Create(entryName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create zip entry %s: %v\n", entryName, err))
	}

	// Write the data
	fingerprint, err := fingerprint(source, destination)
	if err != nil {
		panic(fmt.Sprintf("Failed to write and fingerprint zip entry %s: %v\n", entryName, err))
	}

	// Make sure to check the error on closing the zip.
	err = zipfile.Close()
	if err != nil {
		panic(fmt.Sprintf("Failed to close zip file: %v\n", err))
	}

	objectpath := objectpath(entryName, fingerprint)
	upload(tempFile.Name(), bucket, objectpath)
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

package storerom

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Creates a temporary file
func tempFile() *os.File {

	file, err := ioutil.TempFile("", "")

	if err != nil {
		panic(fmt.Sprintf("Error creating temp file: %v\n", err))
	}

	return file
}

// Creates a temporary directory and returns the name (full path)
func tempDir() string {

	name, err := ioutil.TempDir("", "")

	if err != nil {
		panic(fmt.Sprintf("Error creating temp dir: %v\n", err))
	}

	return name
}

// Lists files in a directory
func listFiles(folder string) []string {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		panic(fmt.Sprintf("Error listing unzipped files: %v\n", err))
	}
	result := make([]string, 5)
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result
}

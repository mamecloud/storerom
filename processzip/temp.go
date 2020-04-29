package processzip

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

package processzip

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"strconv"
)

var projectID := os.Getenv("PROJECT_ID")
var sourceBucket string = fmt.Sprintf("%s-roms-upload", projectID)
var targetBucket string = fmt.Sprintf("%s-roms", projectID)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}


// ProcessZip processes a zipfile into the rom store.
func ProcessZip(ctx context.Context, m PubSubMessage) error {

	object := string(m.Data)
	fmt.Printf("Received file: %s\n", object)

	client := client(ctx)

	fmt.Printf("Downloading object %s from bucket %s\n", object, bucket)
	zipfile := download(bucket, object, client)
	fmt.Printf("Download completed: %s\n", object)

	fmt.Printf("Unzipping zipfile from %s\n", zipfile)
	folder := extractAll(zipfile)
	defer os.RemoveAll(folder)

	fmt.Printf("Processing files from %s\n", folder)
	processRoms(ctx context.Context, folder)

	if err := client.Close(); err != nil {
		panic(fmt.Sprintf("Failed to close client: %v\n", err))
	}
}

func size(size string) int64 {
	result, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse object size %s: %v\n", size, err))
	}
	return result
}

func processRoms(ctx context.Context, folder string) {

	files := listFiles(folder)
	for _, file := range files {
		if len(file) == 0 {
			continue
		}
		filename := filepath.Join(folder, file)

		fi, err := os.Stat(filename)
		if err != nil {
			panic(fmt.Sprintf("Failed to stat %s: %v\n", filename, err))
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			fmt.Printf("%s is a directory, recursing\n", filename)
			processRoms(ctx, filename)
		case mode.IsRegular():
			fmt.Printf("Fingerprinting: %s\n", filename)
			fingerprint := fingerprint(filename)
			fmt.Printf("Fingerprint of %s is %v\n", filename, fingerprint)
			objectpath := objectpath(file, fingerprint)
			fmt.Printf("Object path is %s\n", objectpath)
			if !exists(bucket, objectpath, client) {
				fmt.Printf("Zipping up %s\n", filename)
				zip := zipFile(filename)
				defer os.Remove(zip)
				fmt.Printf("Uploading %s to %s\n", zip, objectpath)
				upload(zip, targetBucket, objectpath, client)
				fmt.Printf("Uploaded %s (%v)\n", filepath.Base(objectpath), fingerprint)
			} else {
				fmt.Printf("Object exists, moving on (%s)\n", objectpath)
			}
		}
	}
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

func processRom(ctx context.Context, filename string) {

	fmt.Printf("Fingerprinting: %s\n", filename)
	fingerprint := fingerprint(filename)
	fmt.Printf("Fingerprint of %s is %v\n", filename, fingerprint)
	objectpath := objectpath(file, fingerprint)
	fmt.Printf("Object path is %s\n", objectpath)
	if !exists(bucket, objectpath, client) {
		fmt.Printf("Zipping up %s\n", filename)
		zip := zipFile(filename)
		defer os.Remove(zip)
		fmt.Printf("Uploading %s to %s\n", zip, objectpath)
		upload(zip, targetBucket, objectpath, client)
		fmt.Printf("Uploaded %s (%v)\n", filepath.Base(objectpath), fingerprint)
	} else {
		fmt.Printf("Object exists, moving on (%s)\n", objectpath)
	}
}

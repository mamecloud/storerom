package processzip

import (
	"context"
	"fmt"
	"os"
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

	fmt.Printf("Downloading object %s from bucket %s\n", object, sourceBucket)
	zipfile := download(sourceBucket, object)

	fmt.Printf("Processing zipfile %s from %s\n", zipfile, object)
	process(zipfile, targetBucket)

	return nil
}

// func size(size string) int64 {
// 	result, err := strconv.ParseInt(size, 10, 64)
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to parse object size %s: %v\n", size, err))
// 	}
// 	return result
// }

// func processRoms(ctx context.Context, folder string) {

// 	var wg sync.WaitGroup
// 	files := listFiles(folder)
// 	for _, file := range files {

// 		// Skip blank filename (current directory?)
// 		if len(file) == 0 {
// 			continue
// 		}
// 		filename := filepath.Join(folder, file)

// 		// Is this a directory?
// 		fi, err := os.Stat(filename)
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to stat %s: %v\n", filename, err))
// 		}
// 		isDir := fi.Mode().IsDir()

// 		if isDir {
// 			// Some zips have subdirectories
// 			fmt.Printf("%s is a directory, recursing\n", filename)
// 			processRoms(ctx, filename)
// 		} else {
// 			wg.Add(1)
// 			processRom(ctx, filename, &wg)
// 		}
// 	}
// 	wg.Wait()
// }

// func processRom(ctx context.Context, filename string, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	client := client(ctx)

// 	fmt.Printf("Fingerprinting: %s\n", filename)
// 	fingerprint := fingerprint(filename)
// 	fmt.Printf("Fingerprint of %s is %v\n", filename, fingerprint)
// 	objectpath := objectpath(filename, fingerprint)
// 	fmt.Printf("Object path for %s is %s\n", filename, objectpath)
// 	duplicate := exists(targetBucket, objectpath, client)
// 	if duplicate {
// 		fmt.Printf("Object exists, moving on (%s)\n", objectpath)
// 		os.Remove(filename)
// 	} else {
// 		fmt.Printf("Zipping up %s\n", filename)
// 		zip := zipFile(filename)
// 		fmt.Printf("Uploading %s to %s\n", zip, objectpath)
// 		upload(zip, targetBucket, objectpath, client)
// 	}

// 	if err := client.Close(); err != nil {
// 		panic(fmt.Sprintf("Failed to close client: %v\n", err))
// 	}

// 	fmt.Printf("Done\n")
// }

// func listFiles(folder string) []string {
// 	files, err := ioutil.ReadDir(folder)
// 	if err != nil {
// 		panic(fmt.Sprintf("Error listing unzipped files: %v\n", err))
// 	}
// 	result := make([]string, 5)
// 	for _, file := range files {
// 		result = append(result, file.Name())
// 	}
// 	return result
// }

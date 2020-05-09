package processzip

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
	"bufio"
)

const timeout = 100

// Create a storage client
func createClient(ctx context.Context) *storage.Client {
	defer duration(track("*client"))

	bctx := context.Background()
	client, err := storage.NewClient(bctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create storage client: %v\n", err))
	}
	return client
}

// Close a storage client
func closeClient(client *storage.Client) {
	if err := client.Close(); err != nil {
		panic(fmt.Sprintf("Failed to close storage client: %v\n", err))
	}
}

// Computes an object path, based on a fingerprint.
func objectpath(file string, fingerprint *Fingerprint) string {
	name := filepath.Base(file)
	return path.Join(name, strconv.FormatInt(fingerprint.Size, 10), fingerprint.Crc, fingerprint.Sha1, name+".zip")
}

// Checks whether the given bucket contains the given object path
func exists(ctx context.Context, bucket string, objectpath string, client *storage.Client) bool {
	defer duration(track(fmt.Sprintf("*exists %s", filepath.Base(objectpath))))
	fmt.Printf("Testing whether %s exists in %s\n", objectpath, bucket)

	tctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	query := &storage.Query{Prefix: objectpath}
	it := client.Bucket(bucket).Objects(tctx, query)
	for {
		_, err := it.Next()
		if err == iterator.Done {
			fmt.Printf("Nothing found for prefix %s\n", objectpath)
			return false
		}
		if err != nil {
			panic(fmt.Sprintf("Error checking if bukcket object exists: %v\n", err))
		}
		fmt.Printf("Found something for prefix %s\n", objectpath)
		return true
	}
}

// Uploads file content to the given bucket and object path
func upload(ctx context.Context, filename string, bucket string, object string, client *storage.Client) {
	defer duration(track(fmt.Sprintf("*upload %s", filepath.Base(object))))
	fmt.Printf("Uploading to %s in bucket %s from %s\n", object, bucket, filename)

	tctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	// Input
	input, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input file: %v\n", err))
	}
	defer input.Close()
	inputBuffer := bufio.NewReader(input)

	// Output
	output := client.Bucket(bucket).Object(object).NewWriter(tctx)
	defer output.Close()
	outputBuffer := bufio.NewWriter(output)

	// Copy
	if _, err := inputBuffer.WriteTo(outputBuffer); err != nil {
		panic(fmt.Sprintf("Failed to download object content: %v\n", err))
	}
	if err := outputBuffer.Flush(); err != nil {
		panic(fmt.Sprintf("Failed to flush output buffer for upload of %s: %v\n", object, err))
	}
}

// Downloads content from the given bucket and object path to a temp file and returns the temp file name
func download(ctx context.Context, bucket string, object string, client *storage.Client) string {
	defer duration(track(fmt.Sprintf("*download %s", filepath.Base(object))))
	fmt.Printf("Domnloading %s from %s\n", object, bucket)

	tctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	// Input
	input, err := client.Bucket(bucket).Object(object).NewReader(tctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input object: %v\n", err))
	}
	defer input.Close()
	inputBuffer := bufio.NewReader(input)

	// Output
	output, err := ioutil.TempFile("", filepath.Base(object)+"_*")
	if err != nil {
		panic(fmt.Sprintf("Error creating temp file: %v\n", err))
	}
	defer output.Close()
	outputBuffer := bufio.NewWriter(output)

	// Copy
	if _, err := inputBuffer.WriteTo(outputBuffer); err != nil {
		panic(fmt.Sprintf("Failed to upload file content: %v\n", err))
	}
	if err := outputBuffer.Flush(); err != nil {
		panic(fmt.Sprintf("Failed to flush output buffer for download of %s: %v\n", object, err))
	}

	return output.Name()
}

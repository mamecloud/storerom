package storerom

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"runtime"
)

// Create a storage client
func client(ctx context.Context) *storage.Client {
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v\n", err))
	}
	return client
}

// Computes a path for a fingerprint.
func objectpath(file string, fingerprint Fingerprint) string {
	name := filepath.Base(file)
	return path.Join(name, strconv.FormatInt(fingerprint.size, 10), fingerprint.crc, fingerprint.sha1, name+".zip")
}

func exists(bucket string, objectpath string, client *storage.Client) bool {

	fmt.Printf("Testing whether %s exists in %s\n", objectpath, bucket)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	query := &storage.Query{Prefix: objectpath}
	it := client.Bucket(bucket).Objects(ctx, query)
	for {
		_, err := it.Next()
		if err == iterator.Done {
			return false
		}
		if err != nil {
			panic(fmt.Sprintf("Error checking if bukcket object exists: %v\n", err))
		}
		return true
	}
}

func upload(filename string, bucket string, object string, client *storage.Client) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	input, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input file: %v\n", err))
	}
	defer input.Close()

	output := client.Bucket(bucket).Object(object).NewWriter(ctx)
	defer output.Close()

	if _, err = io.Copy(output, input); err != nil {
		panic(fmt.Sprintf("Failed to copy content: %v\n", err))
	}
}

// Maximum size of data to attempt to download in one go:
const chunkSize = 10 * 1024 * 1024

func download(bucket string, object string, client *storage.Client) string {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Get the size of the object
	objectHandle := client.Bucket(bucket).Object(object)
	objectAttrs, err := objectHandle.Attrs(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to get object attributes: %v\n", err))
	}
	objectSize := objectAttrs.Size

	// Download in one go, or as chunks, depending on the size
	var result string
	if  objectSize <= chunkSize {
		result = downloadSmall(ctx, objectHandle, client)
	} else {
		result = downloadLarge(ctx, *objectHandle, objectSize, client)
	}
	return result
}

// Single download
func downloadSmall(ctx context.Context, objectHandle *storage.ObjectHandle, client *storage.Client) string {

	input, err := objectHandle.NewReader(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input object: %v\n", err))
	}
	defer input.Close()

	output := tempFile()
	defer output.Close()

	if _, err = io.Copy(output, input); err != nil {
		panic(fmt.Sprintf("Failed to copy content: %v\n", err))
	}

	return output.Name()
}

// Chunked download
func downloadLarge(ctx context.Context, objectHandle storage.ObjectHandle, objectSize int64, client *storage.Client) string {

	offset := int64(0)

	// Calculate the number of chunks needed to download an object of this size
	chunkCount := objectSize / chunkSize
	if objectSize % chunkSize > 0 {
		chunkCount++
	}
	fmt.Printf("Chunk count: object Size:%d / chunkSize:%d = %d:chunks\n", objectSize, chunkSize, chunkCount)
	chunks := make([]string, chunkCount)

	index := 0
	var wg sync.WaitGroup
	for {

		if offset < objectSize {

			// Start a chunk download
			wg.Add(1)
			if offset + chunkSize < objectSize {
				go downloadChunk(ctx, objectHandle, offset, chunkSize, index, chunks, &wg)
			} else {
				go downloadChunk(ctx, objectHandle, offset, -1, index, chunks, &wg)
			}

			// Increment to the next chunk
			index++
			offset += chunkSize

		} else {
			break
		}
	}
	wg.Wait()
	runtime.GC()

	return assembleChunks(chunks)
}


func downloadChunk(ctx context.Context, objectHandle storage.ObjectHandle, offset, length int64, index int, chunks []string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Open a range on the object
	input, err := objectHandle.NewRangeReader(ctx, offset, length)
	if err != nil {
		panic(fmt.Sprintf("Failed to open input object: %v\n", err))
	}
	defer input.Close()

	// Create a temp file to receive the range
	output := tempFile()
	defer output.Close()

	// Download the data
	//fmt.Printf("Downloading chunk %d to %s\n", index, output.Name())
	if _, err = io.Copy(output, input); err != nil {
		panic(fmt.Sprintf("Failed to copy content: %v\n", err))
	}

	chunks[index] = output.Name()
}

func assembleChunks(chunks []string) string {

	// Create a destination file
	output := tempFile()
	defer output.Close()
	assembled := output.Name()

	// Open the chunks as a MultiReader
	readers := make([]io.Reader, 0)
	for index, tempFile := range chunks {

		// Open the chunk
		input, err := os.Open(tempFile)
		if err != nil {
			panic(fmt.Sprintf("Failed to open chunk %d: %v\n", index, err))
		}
		defer input.Close()
		readers = append(readers, input)
	}
	input := io.MultiReader(readers...)

	// Copy chunks to the output file
	count, err := io.Copy(output, input)
	if err != nil {
		panic(fmt.Sprintf("Failed to collate chunks: %v\n", err))
	}
	runtime.GC()

	fmt.Printf("Assembled %d bytes to %s", count, assembled)
	return assembled
}

// func assembleChunks(chunks []string) string {

// 	// Create a destination file
// 	output := tempFile()
// 	defer output.Close()
// 	assembled := output.Name()
// 	var outputSize int64

// 	// Process each chunk into the destination file
// 	for index, tempFile := range chunks {

// 		// Open the chunk
// 		input, err := os.Open(tempFile)
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to open chunk %d: %v\n", index, err))
// 		}

// 		// Copy it to the output file
// 		fmt.Printf("Collecting chunk %d to %s\n", index, output.Name())
// 		count, err := io.Copy(output, input)
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to copy content of chunk %d: %v\n", index, err))
// 		} else {
// 			outputSize += count
// 		}

// 		// Close the chunk
// 		if err := input.Close(); err != nil {
// 			panic(fmt.Sprintf("Failed to close chunk %d: %v\n", index, err))
// 		}
// 	}

// 	fmt.Printf("Assembled %d bytes to %s", outputSize, assembled)
// 	return assembled
// }

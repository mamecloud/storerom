package storerom

import (
	"context"
	"fmt"
	"os"
	"io"
	"time"
	"strconv"
	"path"
	"path/filepath"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// Computes a path for a fingerprint.
func objectpath(file string, fingerprint Fingerprint) string {
        name := filepath.Base(file)
        return path.Join(name, strconv.FormatInt(fingerprint.size, 10), fingerprint.crc, fingerprint.sha1, name + ".zip")
}

// Create a storage client
func client(ctx context.Context) *storage.Client {
        client, err := storage.NewClient(ctx)
        if err != nil {
                panic(fmt.Sprintf("Failed to create client: %v\n", err))
        }
        return client
}

func exists(bucket string, object string, client *storage.Client) bool {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	query := &storage.Query{Prefix: object}
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

// func download(bucket string, object string, client *storage.Client) string {

// 	ctx := context.Background()
// 	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
// 	defer cancel()

// 	input, err := client.Bucket(bucket).Object(object).NewReader(ctx)
// 	if err != nil {
// 			panic(fmt.Sprintf("Failed to open input object: %v\n", err))
// 	}
// 	defer input.Close()

// 	output := tempFile()
// 	defer output.Close()

// 	if _, err = io.Copy(output, input); err != nil {
// 			panic(fmt.Sprintf("Failed to copy content: %v\n", err))
// 	}

// 	return output.Name()
// }

func download(bucket string, object string, client *storage.Client) string {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	objectHandle := client.Bucket(bucket).Object(object)
	objectAttrs, err := objectHandle.Attrs(ctx)
	if err != nil {
			panic(fmt.Sprintf("Failed to get object attributes: %v\n", err))
	}
	if objectAttrs.Size < 10 * 1024 * 1024 {
		return downloadSmall(ctx, objectHandle)
	} else {
		return downloadLarge(ctx, objectHandle)
	}
}

func downloadSmall(ctx context.Context, objectHandle storage.ObjectHandle, client *storage.Client) string {

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

func downloadLarge(ctx context.Context, objectHandle storage.ObjectHandle, size int64, client *storage.Client) string {

	offset := int64(0)
	length := int64(10 * 1024 * 1024)

	output := tempFile()
	for { repeats...
		input, err := objectHandle.NewRangeReader(ctx, offset, length)
		if err != nil {
				panic(fmt.Sprintf("Failed to open input object: %v\n", err))
		}
		defer input.Close()

		defer output.Close()

		if _, err = io.Copy(output, input); err != nil {
				panic(fmt.Sprintf("Failed to copy content: %v\n", err))
		}
	}

	return output.Name()
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

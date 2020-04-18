package storerom

import (
	"context"
	"fmt"
	"os"
	"io"
        "path/filepath"
        "strings"
	"time"
)

var bucket string = "mamecloud-roms"

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
        Kind                    string                 `json:"kind"`
        ID                      string                 `json:"id"`
        SelfLink                string                 `json:"selfLink"`
        Name                    string                 `json:"name"`
        Bucket                  string                 `json:"bucket"`
        Generation              string                 `json:"generation"`
        Metageneration          string                 `json:"metageneration"`
        ContentType             string                 `json:"contentType"`
        TimeCreated             time.Time              `json:"timeCreated"`
        Updated                 time.Time              `json:"updated"`
        TemporaryHold           bool                   `json:"temporaryHold"`
        EventBasedHold          bool                   `json:"eventBasedHold"`
        RetentionExpirationTime time.Time              `json:"retentionExpirationTime"`
        StorageClass            string                 `json:"storageClass"`
        TimeStorageClassUpdated time.Time              `json:"timeStorageClassUpdated"`
        Size                    string                 `json:"size"`
        MD5Hash                 string                 `json:"md5Hash"`
        MediaLink               string                 `json:"mediaLink"`
        ContentEncoding         string                 `json:"contentEncoding"`
        ContentDisposition      string                 `json:"contentDisposition"`
        CacheControl            string                 `json:"cacheControl"`
        Metadata                map[string]interface{} `json:"metadata"`
        CRC32C                  string                 `json:"crc32c"`
        ComponentCount          int                    `json:"componentCount"`
        Etag                    string                 `json:"etag"`
        CustomerEncryption      struct {
                EncryptionAlgorithm string `json:"encryptionAlgorithm"`
                KeySha256           string `json:"keySha256"`
        }
        KMSKeyName    string `json:"kmsKeyName"`
        ResourceState string `json:"resourceState"`
}

// StoreRom processes a zipfile upload into the rom store.
func StoreRom(ctx context.Context, e GCSEvent) error {

        bucket := e.Bucket
        object := e.Name

        fmt.Printf("Received file: %s\n", object)
        if strings.ToLower(filepath.Ext(object)) != ".zip" {
                fmt.Printf("Not a zip file, moving on: %s\n", object)
        } else {
                fmt.Printf("File size is %s\n", e.Size)
                storeRom(ctx, bucket, object)
        }

        return nil
}

// Does the work of storing rom files.
func storeRom(ctx context.Context, bucket string, object string) {

        client := client(ctx)

        fmt.Printf("Downloading object %s from bucket %s\n", object, bucket)
        zipfile := download(bucket, object, client)

        fmt.Printf("Unzipping zipfile from %s\n", zipfile)
        folder := extractAll(zipfile)
        defer os.RemoveAll(folder)
        
        fmt.Printf("Processing files from %s\n", folder)
        files := listFiles(folder)
        for _, file := range files {
                if len(file) == 0 {
                        continue
                }
                filename := filepath.Join(folder, file)
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
                        upload(zip, "mamecloud-roms", objectpath, client)
                        fmt.Printf("Uploaded %s (%v)\n", filepath.Base(objectpath), fingerprint)
                } else {
                        fmt.Printf("Object exists, moving on (%s)\n", objectpath)
                }
        }

        if err := client.Close(); err != nil {
                panic(fmt.Sprintf("Failed to close client: %v\n", err))
        }
}

// Local test function.
func Testing() {

        fmt.Printf("Downloading object %s from bucket %s\n", "rtypeleo.zip", bucket)
        temp := tempFile()
        zipname := temp.Name()
        romfile, err := os.Open("../rtypeleo.zip")
	if err != nil {
                panic(fmt.Sprintf("Failed to open input file: %v\n", err))
	}
        defer romfile.Close()

	if _, err = io.Copy(temp, romfile); err != nil {
			panic(fmt.Sprintf("Failed to copy content: %v\n", err))
        }
	if err != nil {
                panic(fmt.Sprintf("Failed to copy input file: %v\n", err))
	}
        if err := temp.Close(); err != nil {
                panic(fmt.Sprintf("Failed to close output file: %v\n", err))
        }
        fmt.Printf("Copied rtypeleo.zip to %s\n", zipname)

        fmt.Printf("Unzipping zipfile from %s\n", zipname)
        folder := extractAll(zipname)
        defer os.RemoveAll(folder)
        fmt.Printf("Extracted to %s\n", folder)
        
        files := listFiles(folder)
        fmt.Printf("Processing files from %s: %v\n", folder, files)
        for _, file := range files {
                if len(file) == 0 {
                        continue
                }
                filename := filepath.Join(folder, file)
                fmt.Printf("Fingerprinting [%s]\n", filename)
                fingerprint := fingerprint(filename)
                fmt.Printf("Fingerprint of %s is %v\n", filename, fingerprint)
                objectpath := objectpath(file, fingerprint)
                fmt.Printf("Object path will be %s\n", objectpath)
        //     if !exists(bucket, objectpath, client) {
                zip := zipFile(filename)
                defer os.Remove(zip)
                //upload(zip, "mamecloud-roms", objectpath, client)
                // fmt.Printf("Uploaded %s\n", filepath.Base(objectpath))
                fmt.Printf("Processed %s\n", zip)
        //     }
        }
}
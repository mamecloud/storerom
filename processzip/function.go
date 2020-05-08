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
	client := createClient(ctx)

	fmt.Printf("Downloading object %s from bucket %s\n", object, sourceBucket)
	zipfile := download(ctx, sourceBucket, object, client)

	fmt.Printf("Processing zipfile %s from %s\n", zipfile, object)
	process(ctx, zipfile, targetBucket, client)

	return nil
}
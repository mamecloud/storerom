package sortbysize

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"strconv"
	"cloud.google.com/go/pubsub"
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

const topicSmall = "rom-upload-small"
const topicMedium = "rom-upload-medium"
const topicLarge = "rom-upload-large"
const topicXLarge = "rom-upload-xlarge"

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// PublishRom publishes a message about an uploaded rom.
// The topic the rom is published to depends on the file size.
// This enables functions with different memory allocationt to handle different file sizes. 
func PublishRom(ctx context.Context, e GCSEvent) error {

	bucket := e.Bucket
	object := e.Name
	isRom := strings.ToLower(filepath.Ext(object)) == ".zip"

	if isRom {

		sizeM := size(e.Size)
		fmt.Printf("Received zip file %s (%dM)\n", object, sizeM)

		// Choose the topic to publish to, based on the file size:
		var topicID string
		if sizeM < 0 {
			// Disabled for now because we always get one or two errors.
			topicID = topicSmall
		} else if sizeM < 20 {
			topicID = topicMedium
		} else if sizeM < 80 {
			topicID = topicLarge
		} else {
			topicID = topicXLarge
		}

		publish(bucket, object, topicID)

	} else {

		fmt.Printf("Not a zip file, Skipping: %s\n", object)

	}

	return nil
}

// Converts the string size in bytes to int64 size in megabytes
func size(size string) int64 {
	result, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse object size %s: %v\n", size, err))
	}
	return result / (1024 * 1024)
}

// Published an object name to the given Pubsub topic
func publish(bucket, object string, topicID string) {

	// Pubsub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "mamecloud")
	if err != nil {
		panic(fmt.Sprintf("Error creating pubsub client: %v\n", err))
	}
	t := client.Topic(topicID)

	// Publish
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(object),
	})
	
	// The Get method blocks until a server-generated ID or
	// an error is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		panic(fmt.Sprintf("Error publishing %s: %v\n", object, err))
	}
	fmt.Printf("Published message for %s to %s; msg ID: %v\n", object, topicID, id)
}

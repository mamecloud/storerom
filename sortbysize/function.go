package sortbysize

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var projectID string = os.Getenv("PROJECT_ID")

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

	object := e.Name
	isRom := strings.ToLower(filepath.Ext(object)) == ".zip"

	if isRom {

		sizeM, display := size(e.Size)
		fmt.Printf("Received zip file %s (%s)\n", object, display)

		// Choose the topic to publish to, based on the file size:
		var topicID string
		if sizeM < 5 {
			topicID = topicSmall
		} else if sizeM < 50 {
			topicID = topicMedium
		} else if sizeM < 100 {
			topicID = topicLarge
		} else {
			topicID = topicXLarge
		}

		publish(object, topicID)

	} else {
		fmt.Printf("%s is not a zip file, skipping.\n", object)
	}

	return nil
}

// Converts the string size in bytes to int64 size in megabytes, plus a display value
func size(size string) (megabytes int64, display string) {
	bytes, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse object size %s: %v\n", size, err))
	}

	if bytes < 1024 {
		display = fmt.Sprintf("%vB", bytes)
	} else if bytes < (1024 * 1024) {
		display = fmt.Sprintf("%vK", bytes/1024)
	} else {
		display = fmt.Sprintf("%vM", bytes/(1024*1024))
	}

	megabytes = bytes / (1024 * 1024)
	return megabytes, display
}

// Published an object name to the given Pubsub topic
func publish(object string, topicID string) {

	// Pubsub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
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
	fmt.Printf("Published message to %s for %s. Msg ID: %v\n", topicID, object, id)
}

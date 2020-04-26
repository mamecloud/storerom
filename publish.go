package storerom

import (
	"context"
	"fmt"
	"cloud.google.com/go/pubsub"
)

const topicSmall = "rom-upload-small"
const topicMedium = "rom-upload-medium"
const topicLarge = "rom-upload-large"
const topicXLarge = "rom-upload-xlarge"

func publish(bucket, object string, sizeM int64) {

	// Pubsub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "mamecloud")
	if err != nil {
		panic(fmt.Sprintf("Error creating pubsub client: %v\n", err))
	}

	// Choose the topic to publish to, based on the file size:
	var topicID string
	if sizeM < 10 {
		topicID = topicSmall
	} else if sizeM < 50 {
		topicID = topicMedium
	} else if sizeM < 90 {
		topicID = topicLarge
	} else {
		topicID = topicXLarge
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
	fmt.Printf("Published message for %s; msg ID: %v\n", object, id)
}
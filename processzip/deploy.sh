#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat ../account.txt)
project_id=$(cat ../project.txt)
gcloud config set account $account
gcloud config set project $project_id

# Process small rom files from pubsub
gcloud functions deploy ProcessZipSmall \
    --entry-point ProcessZip \
    --runtime go111 \
    --memory 128MB \
    --trigger-topic rom-upload-small \
    --set-env-vars PROJECT_ID=$project_id \
    --allow-unauthenticated \
    --region europe-west2 &

# Process medium rom files from pubsub
gcloud functions deploy ProcessZipMedium \
    --entry-point ProcessZip \
    --runtime go111 \
    --memory 256MB \
    --trigger-topic rom-upload-medium \
    --set-env-vars PROJECT_ID=$project_id \
    --allow-unauthenticated \
    --region europe-west2 &

# Process large rom files from pubsub
gcloud functions deploy ProcessZipLarge \
    --entry-point ProcessZip \
    --runtime go111 \
    --memory 512MB \
    --trigger-topic rom-upload-large \
    --set-env-vars PROJECT_ID=$project_id \
    --allow-unauthenticated \
    --region europe-west2 &

# Process x-large rom files from pubsub
gcloud functions deploy ProcessZipXLarge \
    --entry-point ProcessZip \
    --runtime go111 \
    --memory 1024MB \
    --trigger-topic rom-upload-xlarge \
    --set-env-vars PROJECT_ID=$project_id \
    --allow-unauthenticated \
    --region europe-west2 &

wait



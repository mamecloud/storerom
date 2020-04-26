#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat account.txt)
project_id=$(cat project.txt)
gcloud config set account $account
gcloud config set project $project_id

# Upload bucket
gsutil ls -b gs://mamecloud-roms-upload || \
gsutil mb -l europe-west2 -b on gs://mamecloud-roms-upload && \
gsutil lifecycle set lifecycle.json gs://mamecloud-roms-upload

# Destination bucket
gsutil ls -b gs://mamecloud-roms || \
gsutil mb -l europe-west2 -b on gs://mamecloud-roms

# Pubsub topics
gcloud pubsub topics describe rom-upload-small || \
gcloud pubsub topics create rom-upload-small
gcloud pubsub topics describe rom-upload-medium || \
gcloud pubsub topics create rom-upload-medium
gcloud pubsub topics describe rom-upload-large || \
gcloud pubsub topics create rom-upload-large
gcloud pubsub topics describe rom-upload-xlarge || \
gcloud pubsub topics create rom-upload-xlarge

# Process small rom files from pubsub
gcloud functions deploy StoreRomSmall \
    --entry-point StoreRom \
    --runtime go111 \
    --memory 256MB \
    --trigger-topic rom-upload-small \
    --allow-unauthenticated \
    --region europe-west2 &

# Process medium rom files from pubsub
gcloud functions deploy StoreRomMedium \
    --entry-point StoreRom \
    --runtime go111 \
    --memory 512MB \
    --trigger-topic rom-upload-medium \
    --allow-unauthenticated \
    --region europe-west2 &

# Process large rom files from pubsub
gcloud functions deploy StoreRomLarge \
    --entry-point StoreRom \
    --runtime go111 \
    --memory 1024MB \
    --trigger-topic rom-upload-large \
    --allow-unauthenticated \
    --region europe-west2 &

# Process x-large rom files from pubsub
gcloud functions deploy StoreRomXLarge \
    --entry-point StoreRom \
    --runtime go111 \
    --memory 2048MB \
    --trigger-topic rom-upload-xlarge \
    --allow-unauthenticated \
    --region europe-west2 &

wait

# Publish rom files to pubsub
gcloud functions deploy PublishRom \
    --runtime go111 \
    --memory 128MB \
    --trigger-resource mamecloud-roms-upload \
    --trigger-event google.storage.object.finalize \
    --allow-unauthenticated \
    --region europe-west2



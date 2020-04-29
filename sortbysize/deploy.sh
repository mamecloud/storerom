#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat ../account.txt)
project_id=$(cat ../project.txt)
gcloud config set account $account
gcloud config set project $project_id

# Function: publish rom files to pubsub
gcloud functions deploy PublishRom \
    --runtime go111 \
    --memory 128MB \
    --trigger-resource mamecloud-roms-upload \
    --trigger-event google.storage.object.finalize \
    --allow-unauthenticated \
    --region europe-west2
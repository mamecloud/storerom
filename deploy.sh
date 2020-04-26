#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat account.txt)
project_id=$(cat project.txt)
gcloud config set account $account
gcloud config set project $project_id

# Upload bucket
gsutil ls -b gs://${project_id}-roms-upload || \
gsutil mb -l europe-west2 -b on gs://${project_id}-roms-upload && \
gsutil lifecycle set lifecycle.json gs://${project_id}-roms-upload

# Rom store bucket
gsutil ls -b gs://${project_id}-roms || \
gsutil mb -l europe-west2 -b on gs://${project_id}-roms

# Pubsub topics
gcloud pubsub topics describe rom-upload-small || \
gcloud pubsub topics create rom-upload-small
gcloud pubsub topics describe rom-upload-medium || \
gcloud pubsub topics create rom-upload-medium
gcloud pubsub topics describe rom-upload-large || \
gcloud pubsub topics create rom-upload-large
gcloud pubsub topics describe rom-upload-xlarge || \
gcloud pubsub topics create rom-upload-xlarge

base=$PWD

cd $base/sortbysize
./deploy.sh

cd $base/processzip
./deploy.sh

cd $base

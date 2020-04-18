#!/usr/bin/env bash

account=david@carboni.io
project_id=mamecloud
gcloud config set account $account
gcloud config set project $project_id

gsutil ls -b gs://mamecloud-roms-upload || \
gsutil mb -b on -l europe-west2 gs://mamecloud-roms-upload

gsutil ls -b gs://mamecloud-roms || \
gsutil mb -b on -l europe-west2 gs://mamecloud-roms

gsutil ls -b gs://mamecloud-extras-upload || \
gsutil mb -b on -l europe-west2 gs://mamecloud-extras-upload

gsutil ls -b gs://mamecloud-extras || \
gsutil mb -b on -l europe-west2 gs://mamecloud-extras

gsutil ls -b gs://mamecloud || \
gsutil mb -b on -l europe-west2 gs://mamecloud

gsutil ls -b gs://mamecloud.com || \
gsutil mb -b on -l europe-west2 gs://mamecloud.com && \
gsutil web set -m index.html gs://mamecloud.com

gcloud functions deploy StoreRom \
    --runtime go111 \
    --memory 512MB \
    --trigger-resource mamecloud-roms-upload \
    --trigger-event google.storage.object.finalize \
    --allow-unauthenticated \
    --region europe-west2 \
    && \
echo Pausing... && \
sleep 5 && \
gsutil cp ~/Downloads/rtypeleo.zip gs://mamecloud-roms-upload

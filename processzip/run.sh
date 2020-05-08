#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat ../account.txt)
project_id=$(cat ../project.txt)
gcloud config set account $account
gcloud config set project $project_id

gsutil -m cp -r gs://mamecloud-temp/* gs://mamecloud-roms-upload
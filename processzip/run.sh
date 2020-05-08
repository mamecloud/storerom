#!/usr/bin/env bash
set -euxo pipefail 

# Configuration:
account=$(cat ../account.txt)
project_id=$(cat ../project.txt)
gcloud config set account $account
gcloud config set project $project_id

gsutil -m rm -r gs://${project_id}-roms
gsutil mb -l europe-west2 -b on gs://${project_id}-roms
gsutil -m cp -r gs://${project_id}-temp/* gs://${project_id}-roms-upload

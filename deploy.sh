#!/usr/bin/env bash

account=david@carboni.io
project_id=mamecloud
gcloud config set account $account
gcloud config set project $project_id

gsutil mb gs://mamecloud-roms-upload

gcloud functions deploy collect_zip --runtime python37 --memory=512MB --trigger-resource mamecloud-roms-upload --trigger-event google.storage.object.finalize --allow-unauthenticated --region europe-west2 #--service-account romcollector@mamecloud.iam.gserviceaccount.com

#gsutil rm gs://mamecloud-roms-upload/deploy.sh
#gsutil cp ./deploy.sh gs://mamecloud-roms-upload

#gsutil rm gs://mamecloud-roms-upload/MAME\ \(v0.185\)\ -\ mtchxlgld.zip 
#gsutil cp ../torrent/processed/MAME\ 0.219\ Rollback\ ROMs/MAME\ \(v0.185\)\ -\ mtchxlgld.zip gs://mamecloud-roms-upload

sleep 5
gsutil cp ~/Downloads/rtypeleo.zip gs://mamecloud-roms-upload

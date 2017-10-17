#!/bin/bash


here=$(pwd)
(
  cd ..
  GOOS=linux GOARCH=amd64 go build 
  cp hpt $here/hpt
)

PACKER_LOG=1 packer build \
    -var "do_api_token=$DIGITALOCEAN_API_TOKEN" \
    -var "access_key=$DO_ACCESS_KEY" \
    -var "secret_key=$DO_SECRET_ACCESS_KEY" \
    centos7.json

#delete images
#imageid=$(doctl comput image list | awk '{if($2==}')
#for image_id in $(doctl compute image list | awk '{if($2 ~ /hpt-test/) print $1}')
#do
#    echo "deleting image id: $image_id"
#    doctl compute image delete $image_id --force
#done

rm hpt

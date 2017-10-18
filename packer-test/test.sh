#!/bin/bash


here=$(pwd)
(
  cd ..
  GOOS=linux GOARCH=amd64 go build 
  cp hpt $here/hpt
)

PACKER_LOG=1 packer build centos7.json

#delete images
#imageid=$(doctl comput image list | awk '{if($2==}')
#for image_id in $(doctl compute image list | awk '{if($2 ~ /hpt-test/) print $1}')
#do
#    echo "deleting image id: $image_id"
#    doctl compute image delete $image_id --force
#done

rm hpt

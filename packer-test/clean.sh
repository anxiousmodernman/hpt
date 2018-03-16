#!/bin/bash


# Get rid of our hpt-* test images

images=$(doctl compute image list | grep hpt-test- | awk '{print $1}')

for i in $images 
do
   echo "deleting image $i"
   doctl compute image delete $i --force
done


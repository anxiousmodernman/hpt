#!/bin/bash

pushd ..
go build
popd

cp ../hpt .

packer build base.json

rm -rf hpt

#!/bin/bash

set -e

local_repo=/home/tester/go/src/github.com/anxiousmodernman/hpt
user_dir=/home/tester/go/src/github.com/anxiousmodernman

useradd "test"

mkdir -p /home/tester/go/bin
mkdir -p $user_dir
yum install -y git wget unzip
cd $user_dir
git clone https://github.com/anxiousmodernman/hpt.git
cd /tmp
wget --progress=dot:mega https://dl.google.com/go/go1.10.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.10.linux-amd64.tar.gz
ln -s /usr/local/go/bin/go /usr/bin/go
export GOPATH=/home/tester/go
export PATH=$PATH:/home/tester/hpt/go/bin
cd /tmp
mkdir protobuf
wget --progress=dot:mega https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
chmod +x dep-linux-amd64
mv dep-linux-amd64 /usr/bin/dep #/home/tester/go/bin/dep
wget https://coleman.nyc3.digitaloceanspaces.com/software/protoc-gen-go/protoc-gen-go
chmod +x protoc-gen-go
mv protoc-gen-go /bin
wget --progress=dot:mega -P /tmp/protobuf https://github.com/google/protobuf/releases/download/v3.5.0/protoc-3.5.0-linux-x86_64.zip
cd protobuf
unzip protoc-3.5.0-linux-x86_64.zip
ls /tmp/protobuf/bin
sudo mv /tmp/protobuf/bin/protoc /bin
cd $local_repo
ls
#echo HACK: remove lockfile
#rm Gopkg.lock
echo "TRY MAKING VENDOR MANUALLY"
mkdir vendor
dep ensure -update
echo "AGAIN WHAT IS HERE: $(ls)"
echo "AGAIN WHAT IS VENDOR: $(ls vendor)"
./gen.sh
go build

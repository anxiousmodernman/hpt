#!/bin/bash

local_repo=/home/tester/go/src/github.com/anxiousmodernman/hpt
user_dir=/home/tester/go/src/github.com/anxiousmodernman

mkdir -p /home/tester/go/bin
mkdir -p $user_dir
yum install -y git wget unzip
cd $user_dir
echo GIT IS $(which git)
git clone https://github.com/anxiousmodernman/hpt.git
echo we are here $(pwd)
ls
echo inside!!!!
ls hpt
cd /tmp
wget --progress=dot:mega https://dl.google.com/go/go1.10.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.10.linux-amd64.tar.gz
ln -s /usr/local/go/bin/go /usr/bin/go
export GOPATH=/home/tester/go
export PATH=$PATH:/home/tester/hpt/go/bin
cd /tmp
echo PWD: $(pwd)
mkdir protobuf
wget --progress=dot:mega https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
chmod +x dep-linux-amd64
ls
echo PWD: $(pwd)
mv dep-linux-amd64 /home/tester/go/bin/dep
echo wtffffff
echo PWD: $(pwd)
ls /home/tester/go/bin
wget --progress=dot:mega -P /tmp/protobuf https://github.com/google/protobuf/releases/download/v3.5.0/protoc-3.5.0-linux-x86_64.zip
cd protobuf
echo PWD: $(pwd)
unzip protoc-3.5.0-linux-x86_64.zip
ls /tmp/protobuf/bin
sudo mv /tmp/protobuf/bin/protoc /bin
cd $local_repo
echo PWD: $(pwd)
ls
echo HACK: remove lockfile?!?!?!?!
rm Gopkg.lock
echo HACK: install protoc plugin with go get?? CMON!
go install github.com/golang/protobuf/protoc-gen-go
echo HACK: did it work? $(which protoc-gen-go)
/home/tester/go/bin/dep ensure
./gen.sh
/usr/local/go/bin/go build

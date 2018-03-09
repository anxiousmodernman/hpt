#!/bin/bash

go install ./vendor/github.com/golang/protobuf/protoc-gen-go 
protoc ./proto/hpt.proto --go_out=plugins=grpc:${GOPATH}/src


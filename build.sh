#!/bin/bash

protoc ./proto/hpt.proto --go_out=plugins=grpc:${GOPATH}/src

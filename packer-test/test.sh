#!/bin/bash

export GIT_BRANCH=`git rev-parse --abbrev-ref HEAD` 
packer build ci-fast.json


#!/bin/bash

# Unfortunately, some of our tests need to run as root. We hide them behind 
# a t.Skip call in our Go code, and we must call them by name here.

# Root does not have a GOPATH, so we provide ours. Here ya go, root.
export GOPATH="/home/coleman/go"

go test -run CreateFileWithPermissions



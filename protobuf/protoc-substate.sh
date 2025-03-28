#!/bin/bash

#go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.8

protoc --go_out=. substate.proto
protoc --go_out=. misc.proto

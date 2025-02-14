#!/bin/bash

set -e


if [[ -z $PROTOBUF_REPO ]]; then
  echo "PROTOBUF_REPO not set"
  exit -1
fi


protos=(
"any.proto"
"api.proto"
"descriptor.proto"
"duration.proto"
"empty.proto"
"field_mask.proto"
"source_context.proto"
"struct.proto"
"timestamp.proto"
"type.proto"
"wrappers.proto"
)

for proto in "${protos[@]}";
do
  cp -v $PROTOBUF_REPO/src/google/protobuf/$proto ./
done

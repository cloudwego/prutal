#!/bin/bash
set -x
set -e

echo "generating parser code ... "
# https://www.antlr.org/download.html
# or brew install antlr on mac
antlr -Dlanguage=Go -o parser ./Protobuf.g4

# we need to support old Go versions
# see: https://github.com/antlr/antlr4/pull/4754
echo "replacing antlr to internal ... "
GITHUB_PKG="github.com/antlr4-go/antlr/v4"
INTERNAL_PKG="github.com/cloudwego/prutal/prutalgen/internal/antlr"
sed -i.bak "s:$GITHUB_PKG:$INTERNAL_PKG:g" ./parser/*.go && rm ./parser/*.bak

echo "all done"

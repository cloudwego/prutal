#!/bin/bash

set -e

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'

PATH_BIN=$PWD/bin
export PATH=${PATH_BIN}:$PATH
mkdir -p $PATH_BIN

echo "building prutalgen ..."
cd ../prutalgen
go build -v
mv prutalgen $PATH_BIN
cd - >/dev/null
echo -e "building prutalgen ... ${GREEN}done${RESET}"

echo "installing protoc ... "
PROTOC_VERSION=v29.3
if [[ ! -f "${PATH_BIN}/protoc" ]]; then
  mkdir -p tmp
  cd tmp
  os=`uname -s | sed 's/Darwin/osx/'`
  arch=`uname -m | sed 's/amd64/x86_64/' | sed 's/arm64/aarch_64/'`
  suffix="${os}-${arch}"
  filename=protoc-${PROTOC_VERSION#v}-${suffix}.zip
  url=https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${filename}
  rm -f $filename
  wget -q $url
  unzip -o -q $filename -d ./
  mv ./bin/protoc $PATH_BIN
  cd - >/dev/null
  rm -rf ./tmp/
  echo -e "installing protoc ... ${GREEN}done${RESET}"
fi

echo "installing protoc-gen-go ..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
echo -e "installing protoc-gen-go ... ${GREEN}done${RESET}"

echo "installing protoc-gen-go-grpc ..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
echo -e "installing protoc-gen-go-grpc ... ${GREEN}done${RESET}"

echo -ne "installed: ${GREEN}"
which protoc 
echo -ne "${RESET}"
protoc --version

echo -ne "installed: ${GREEN}"
which protoc-gen-go-grpc
echo -ne "${RESET}"
protoc-gen-go-grpc --version

echo -ne "installed: ${GREEN}"
which prutalgen
echo -ne "${RESET}"
echo ""

# touch all proto files to ensure `make` will generate new ones
find . -type f -name "*.proto" -exec touch {} +

for dir in ./cases/*/; do
  if [ ! -d "$dir" ]; then
    continue
  fi
  cd $dir
  echo -e "${YELLOW}running test under $dir ...${RESET}"
  if [ -f "Makefile" ] || [ -f "makefile" ]; then
     make test
  elif [ -f "run.sh" ]; then
    ./run.sh
  else
    echo -e "${RED}no makefile or run.sh found${RESET}"
  fi
  echo ""
  cd - >/dev/null
done

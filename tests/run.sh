#!/bin/bash

set -e

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'

PATH_BIN=$PWD/bin
# protoc looks for the well-known type imports in ../include next to its
# binary, so the include dir of the release is kept beside bin
PATH_INCLUDE=$PWD/include
export PATH=${PATH_BIN}:$PATH
mkdir -p "$PATH_BIN"

echo "building prutalgen ..."
cd ../prutalgen
go build -v
mv prutalgen $PATH_BIN
cd - >/dev/null
echo -e "building prutalgen ... ${GREEN}done${RESET}"

echo "installing protoc ... "
PROTOC_VERSION=v29.3
PROTOC_VERSION_OUTPUT="libprotoc ${PROTOC_VERSION#v}"
PROTOC_VERSION_FILE="${PATH_INCLUDE}/.protoc-version"
if [[ ! -x "${PATH_BIN}/protoc" ||
      ! -d "${PATH_INCLUDE}/google/protobuf" ||
      ! -f "${PROTOC_VERSION_FILE}" ||
      "$("${PATH_BIN}/protoc" --version 2>/dev/null)" != "${PROTOC_VERSION_OUTPUT}" ||
      "$(<"${PROTOC_VERSION_FILE}")" != "${PROTOC_VERSION}" ]]; then
  mkdir -p tmp
  cd tmp
  os=`uname -s | sed 's/Darwin/osx/'`
  arch=`uname -m | sed 's/amd64/x86_64/' | sed 's/arm64/aarch_64/'`
  suffix="${os}-${arch}"
  filename=protoc-${PROTOC_VERSION#v}-${suffix}.zip
  url=https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${filename}
  rm -f "$filename"
  rm -rf ./bin ./include
  wget -q "$url"
  unzip -o -q "$filename" -d ./
  if [[ "$(./bin/protoc --version)" != "${PROTOC_VERSION_OUTPUT}" ||
        ! -d ./include/google/protobuf ]]; then
    echo -e "${RED}downloaded protoc release is incomplete or has the wrong version${RESET}"
    exit 1
  fi
  rm -rf "$PATH_INCLUDE"
  mv ./include "$PATH_INCLUDE"
  mv -f ./bin/protoc "${PATH_BIN}/protoc"
  printf '%s\n' "$PROTOC_VERSION" > "$PROTOC_VERSION_FILE"
  cd - >/dev/null
  rm -rf ./tmp/
  echo -e "installing protoc ... ${GREEN}done${RESET}"
fi

# the generators are pinned: their output is what the cases test against,
# and protoc-gen-go must match the protobuf runtime in go.mod
echo "installing protoc-gen-go ..."
PROTOC_GEN_GO_VERSION=$(go list -m -f '{{.Version}}' google.golang.org/protobuf)
GOBIN="$PATH_BIN" go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
echo -e "installing protoc-gen-go ... ${GREEN}done${RESET}"

echo "installing protoc-gen-go-grpc ..."
GOBIN="$PATH_BIN" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
echo -e "installing protoc-gen-go-grpc ... ${GREEN}done${RESET}"

# Forget any executable locations cached by the shell before using the pins.
hash -r

echo -ne "installed: ${GREEN}"
echo "${PATH_BIN}/protoc"
echo -ne "${RESET}"
"${PATH_BIN}/protoc" --version

echo -ne "installed: ${GREEN}"
echo "${PATH_BIN}/protoc-gen-go"
echo -ne "${RESET}"
"${PATH_BIN}/protoc-gen-go" --version

echo -ne "installed: ${GREEN}"
echo "${PATH_BIN}/protoc-gen-go-grpc"
echo -ne "${RESET}"
"${PATH_BIN}/protoc-gen-go-grpc" --version

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

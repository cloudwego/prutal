#!/bin/bash

set -e

# build prutalgen
echo "building prutalgen ..."
cd ../prutalgen
go build -v
cd - >/dev/null

# copy to bin and add to PATH
mkdir -p bin
cp ../prutalgen/prutalgen bin/
export PATH=$PWD/bin:$PATH
which prutalgen
echo ""

# touch all proto files to ensure `make` will generate new ones
find . -type f -name "*.proto" -exec touch {} +

for dir in ./cases/*/; do
  if [ ! -d "$dir" ]; then
    continue
  fi
  cd $dir
  echo "running test under $dir ..."
  if [ -f "Makefile" ] || [ -f "makefile" ]; then
     make test
  elif [ -f "run.sh" ]; then
    ./run.sh
  else
    echo "no makefile or run.sh found"
  fi
  echo ""
  cd - >/dev/null
done

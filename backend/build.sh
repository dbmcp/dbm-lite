#!/bin/bash
set -e

echo "Building dbm-lite backend..."
cd "$(dirname "$0")"

if [ ! -d "./data" ]; then
  mkdir -p ./data
fi

go mod tidy
go build -o ./bin/dbm-lite ./cmd/server
echo "Build completed: bin/dbm-lite"

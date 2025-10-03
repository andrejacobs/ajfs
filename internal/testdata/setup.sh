#!/bin/bash
# This sets up all the required test data

go run generate.go -- ./

./generate-expected-hashes.sh

echo "Generate expected ./scan path list..."
find ./scan ! -name '.DS_Store' > ./expected/scan.txt

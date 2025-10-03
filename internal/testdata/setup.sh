#!/bin/bash
# This sets up all the required test data

go run generate.go -- ./

./generate-expected-hashes.sh

echo "Generate expected ./scan path list..."
cd ./scan && find . ! -name '.DS_Store' > ../expected/scan.txt
if [[ "$(uname)" == "Darwin" ]]; then
    sed -i '' 's|^\./||' ../expected/scan.txt
else
    sed -i 's|^\./||' ../expected/scan.txt
fi

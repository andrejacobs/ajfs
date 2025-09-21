#!/bin/bash
# This sets up all the required test data

go run generate.go -- ./

./generate-expected-hashes.sh

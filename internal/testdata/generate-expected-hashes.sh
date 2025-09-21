#!/bin/bash
mkdir -p expected
cd scan
# hashdeep -c md5 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.md5
hashdeep -c sha1 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.sha1
hashdeep -c sha256 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.sha256

cd ../diff/a
hashdeep -c sha256 -l -r ./ | sed -e '/\.DS_Store/d' -e '/5.txt/d' -e '/7.txt/d' > ../../expected/update-test.sha256

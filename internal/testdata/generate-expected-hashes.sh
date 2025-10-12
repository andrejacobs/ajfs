#!/bin/bash
# Copyright (c) 2025 Andre Jacobs
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

echo "Generating expected hashes..."
mkdir -p expected
cd scan
# hashdeep -c md5 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.md5
hashdeep -c sha1 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.sha1
hashdeep -c sha256 -l -r ./ | sed '/\.DS_Store/d' > ../expected/scan.sha256

cd ../diff/a
hashdeep -c sha256 -l -r ./ | sed -e '/\.DS_Store/d' -e '/5.txt/d' -e '/7.txt/d' > ../../expected/update-test.sha256

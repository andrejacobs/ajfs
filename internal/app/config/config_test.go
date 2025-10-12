// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config_test

import (
	"bytes"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestPrintln(t *testing.T) {
	var buffer bytes.Buffer

	cfg := config.CommonConfig{
		Stdout: &buffer,
	}

	expected := "The quick brown fox jumped over the lazy dog!"
	cfg.Println(expected)
	assert.Equal(t, expected+"\n", buffer.String())
}

func TestVerbosePrintln(t *testing.T) {
	var buffer bytes.Buffer

	cfg := config.CommonConfig{
		Stdout: &buffer,
	}

	ignored := "Verbose is not enabled"
	cfg.VerbosePrintln(ignored)

	cfg.Verbose = true
	expected := "The quick brown fox jumped over the lazy dog!"
	cfg.VerbosePrintln(expected)
	assert.Equal(t, expected+"\n", buffer.String())
}

func TestErrorln(t *testing.T) {
	var buffer bytes.Buffer

	cfg := config.CommonConfig{
		Stderr: &buffer,
	}

	expected := "The quick brown fox jumped over the lazy dog!"
	cfg.Errorln(expected)
	assert.Equal(t, expected+"\n", buffer.String())
}

func TestProgressPrintln(t *testing.T) {
	var buffer bytes.Buffer

	cfg := config.CommonConfig{
		Stdout: &buffer,
	}

	ignored := "Verbose and Progress is not enabled"
	cfg.ProgressPrintln(ignored)
	assert.Equal(t, "", buffer.String())
	buffer.Reset()

	cfg.Progress = true
	expected := "The quick brown fox jumped over the lazy dog!"
	cfg.ProgressPrintln(expected)
	assert.Equal(t, expected+"\n", buffer.String())
}

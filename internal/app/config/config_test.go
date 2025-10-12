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

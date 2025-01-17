package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cu-library/ezproxy-config-lint/internal/linter"
)

func NewLinter() *linter.Linter {
	l := &linter.Linter{Output: io.Discard, FollowIncludeFile: true, AdditionalPHEChecks: true}
	if testing.Verbose() {
		l.Output = os.Stdout
	}
	return l
}

func TestInvalid(t *testing.T) {
	dirContent, err := filepath.Glob("testdata/invalid/*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "testdata/invalid/")
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		ret, _ := l.ProcessFile(f)
		if ret == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestValid(t *testing.T) {
	dirContent, err := filepath.Glob("testdata/valid/*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "testdata/valid/")
		t.Logf("> valid: %s\n", filename)

		l := NewLinter()
		ret, _ := l.ProcessFile(f)
		if ret != 0 {
			t.Errorf("Unexpected error on valid file: %s\n", filename)
		}
	}
}

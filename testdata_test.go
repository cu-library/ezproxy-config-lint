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
	root := "testdata/invalid/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestInvalidPedantic(t *testing.T) {
	root := "testdata/invalid/pedantic/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		l.Pedantic = true
		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
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
		warningCount, err := l.ProcessFile(f)
		if err != nil || warningCount != 0 {
			t.Errorf("Unexpected error on valid file: %s\n", filename)
		}
	}
}

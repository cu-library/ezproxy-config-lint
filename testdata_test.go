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
	l := &linter.Linter{Output: io.Discard, FollowIncludeFile: true}
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

func TestInvalidCase(t *testing.T) {
	root := "testdata/invalid_case/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		l.DirectiveCase = true

		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestInvalidHTTPS(t *testing.T) {
	root := "testdata/invalid_https/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		l.HTTPS = true

		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestInvalidOrigins(t *testing.T) {
	root := "testdata/invalid_origins/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		l.Origins = true

		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestInvalidPHE(t *testing.T) {
	root := "testdata/invalid_phe/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> invalid: %s\n", filename)

		l := NewLinter()
		l.AdditionalPHEChecks = true

		warningCount, err := l.ProcessFile(f)
		if err == nil && warningCount == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestValid(t *testing.T) {
	root := "testdata/valid/"
	dirContent, err := filepath.Glob(root + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, root)
		t.Logf("> valid: %s\n", filename)

		l := NewLinter()
		warningCount, err := l.ProcessFile(f)
		if err != nil || warningCount != 0 {
			t.Errorf("Unexpected error on valid file: %s\n", filename)
		}
	}
}

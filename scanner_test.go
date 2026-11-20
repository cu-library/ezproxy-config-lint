package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cu-library/ezproxy-config-lint/v5/internal/linter"
)

func TestProcessFileScannerLimits(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		tooLong       bool
	}{
		{"empty", "", false},
		{"growth", "#" + strings.Repeat("x", linter.DefaultBufferSize) + "\n", false},
		{"oversized", "#" + strings.Repeat("x", linter.MaxBufferSize), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.txt")
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := linter.NewLinter(linter.Config{Output: io.Discard}).ProcessFile(path)
			if tc.tooLong {
				if !errors.Is(err, bufio.ErrTooLong) {
					t.Fatalf("got %v, want token too long", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Test that nested scanners correctly handle their lifetime and resume processing
// after includes.
func TestProcessFileNestedScannerLifetime(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"parent.txt":     "IncludeFile child.txt\nIncludeFile tail.txt\n",
		"child.txt":      "IncludeFile grandchild.txt\n\n",
		"grandchild.txt": strings.Repeat("\n", 100),
		"tail.txt":       "IncludeFile missing.txt\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for range 10 {
		_, err := linter.NewLinter(linter.Config{Output: io.Discard, FollowIncludeFile: true}).ProcessFile(filepath.Join(dir, "parent.txt"))
		if !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "missing.txt") {
			t.Fatalf("parent must resume and process tail include: %v", err)
		}
	}
}

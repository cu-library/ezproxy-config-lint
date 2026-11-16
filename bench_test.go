package main

import (
	"path/filepath"
	"testing"

	"github.com/cu-library/ezproxy-config-lint/v4/internal/linter"
)

// BenchmarkProcessBenchFile processes testdata/bench.txt and reports memory allocations.
// Include paths in that file are relative to testdata/valid.
func BenchmarkProcessBenchFile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config := linter.Config{IncludeFileDirectory: filepath.Join("testdata", "valid")}
		l := linter.NewLinter(config)
		if _, err := l.ProcessFile("testdata/bench.txt"); err != nil {
			b.Fatal(err)
		}
	}
}

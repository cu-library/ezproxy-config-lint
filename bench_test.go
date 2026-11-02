package main

import (
	"path/filepath"
	"testing"
)

// BenchmarkProcessBenchFile processes testdata/bench.txt and reports memory allocations.
// Include paths in that file are relative to testdata/valid.
func BenchmarkProcessBenchFile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l := NewLinter()
		l.IncludeFileDirectory = filepath.Join("testdata", "valid")
		if _, err := l.ProcessFile("testdata/bench.txt"); err != nil {
			b.Fatal(err)
		}
	}
}

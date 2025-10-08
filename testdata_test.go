package main

import (
	"io"
	"os"
	"path/filepath"
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

type testOpts struct {
	Name    string
	Case    bool
	Fail    bool
	HTTPS   bool
	Origins bool
	PHE     bool
}

func TestDataFiles(t *testing.T) {
	opts := []testOpts{
		{Name: "valid"},
		{Name: "invalid", Fail: true},
		{Name: "invalid_case", Fail: true, Case: true},
		{Name: "invalid_https", Fail: true, HTTPS: true},
		{Name: "invalid_origins", Fail: true, Origins: true},
		{Name: "invalid_phe", Fail: true, PHE: true},
	}

	for _, o := range opts {
		t.Run(o.Name, func(t *testing.T) {
			runDataFileTest(t, o)
		})
	}
}

func runDataFileTest(t *testing.T, o testOpts) {
	root := filepath.Join("testdata", o.Name)
	dirContent, err := filepath.Glob(filepath.Join(root, "*.txt"))
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		t.Logf("> %s", f)
		l := NewLinter()
		l.DirectiveCase = o.Case
		l.HTTPS = o.HTTPS
		l.Origins = o.Origins
		l.AdditionalPHEChecks = o.PHE

		warningCount, err := l.ProcessFile(f)
		if o.Fail && (err == nil && warningCount == 0) {
			t.Errorf("Unexpected success on invalid file: %s\n", f)
		}
	}
}

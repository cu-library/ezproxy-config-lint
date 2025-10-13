package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cu-library/ezproxy-config-lint/internal/linter"
	"github.com/fatih/color"
)

var update = flag.Bool("update", false, "Update golden testdata fixtures")

func NewLinter() *linter.Linter {
	l := &linter.Linter{Output: io.Discard, FollowIncludeFile: true}
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

	// Disable colors for these tests.
	origColor := color.NoColor
	color.NoColor = true

	for _, o := range opts {
		t.Run(o.Name, func(t *testing.T) {
			runDataFileTest(t, o)
		})
	}

	color.NoColor = origColor
}

func runDataFileTest(t *testing.T, o testOpts) {
	root := filepath.Join("testdata", o.Name)
	dirContent, err := filepath.Glob(filepath.Join(root, "*.txt"))
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		l := NewLinter()
		l.DirectiveCase = o.Case
		l.HTTPS = o.HTTPS
		l.Origins = o.Origins
		l.AdditionalPHEChecks = o.PHE

		buf := bytes.NewBuffer(nil)
		l.Output = buf

		warningCount, err := l.ProcessFile(f)

		if o.Fail {
			golden := f + ".golden"
			if *update {
				// If we're in update mode, strip out the terminal color
				// sequences before writing the golden file.
				err := os.WriteFile(golden, buf.Bytes(), 0644)
				if err != nil {
					panic(err)
				}
			}

			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Errorf("Failed to read golden fixture: %s\nUse `go test -update ./...` to update fixtures.", golden)
				continue
			}

			if err == nil && warningCount == 0 {
				t.Errorf("Unexpected success on invalid file: %s\nwant:\n%s", f, expected)
				continue
			}

			// Verify that the output matches the golden fixture.
			if !bytes.Equal(buf.Bytes(), expected) {
				t.Errorf("Results did not match golden fixture:\nwant:\n%s\ngot:\n%s", string(expected), string(buf.Bytes()))
			}
		} else if err != nil || warningCount != 0 {
			t.Errorf("Unexpected error on valid file: %s\n%s", f, string(buf.Bytes()))
		}
	}
}

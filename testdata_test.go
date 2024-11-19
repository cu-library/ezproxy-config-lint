package main

import (
	"log"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalid(t *testing.T) {
	linter := &Linter{}

	dirContent, err := filepath.Glob("testdata/invalid/*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "testdata/invalid/")
		log.Printf("> invalid: %s\n", filename)

		ret, _ := linter.ProcessFile(f)
		if ret == 0 {
			t.Errorf("Unexpected success on invalid file: %s\n", filename)
		}
	}
}

func TestValid(t *testing.T) {
	linter := &Linter{}

	dirContent, err := filepath.Glob("testdata/valid/*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "testdata/valid/")
		log.Printf("> valid: %s\n", filename)

		ret, _ := linter.ProcessFile(f)
		if ret != 0 {
			t.Errorf("Unexpected error on valid file: %s\n", filename)
		}
	}
}

// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLineEndingInSpace(t *testing.T) {
	linter := Linter{Whitespace: true}
	expected := []string{"Line ends in a space or tab character"}
	messages := linter.ProcessLine("Title hello     ")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMissingURL(t *testing.T) {
	linter := Linter{State: State{
		Title: "A Title",
	}}
	expected := []string{"Stanza \"A Title\" has Title but no URL"}
	messages := linter.ProcessLine("")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestTrailingSpaceOrTabCheck(t *testing.T) {
	var tests = []struct {
		line     string
		expected bool
	}{
		{"", false},
		{" ", true},
		{"\t", true},
		{"   a", false},
		{"a ", true},
		{"a    ", true},
		{"a\t", true},
	}

	for _, tt := range tests {
		problem := TrailingSpaceOrTabCheck(tt.line)
		if problem != tt.expected {
			t.Fatalf("TrailingSpaceOrTabCheck() fails on \"%v\", wanted %v, got %v.\n", tt.line, tt.expected, problem)
		}
	}
}

func TestMultilineDirective(t *testing.T) {
	linter := Linter{}
	multiline := `ShibbolethMetadata \
                      -EntityID=EZproxyEntityID \
                      -File=MetadataFile \
                      -SignResponse=false -SignAssertion=true -EncryptAssertion=false \
                      -Cert=EZproxyCertNumber`
	for _, line := range strings.Split(multiline, "\n") {
		messages := linter.ProcessLine(line)
		if len(messages) != 0 {
			t.Fatalf("Multiline directive was not properly processed: %v", messages)
		}
	}
	if linter.State.Previous != ShibbolethMetadata {
		t.Fatalf("Processing multiline directive did not find the correct Directive")
	}
}

func TestFindReplacePair(t *testing.T) {
	linter := Linter{State: State{
		Previous: Find,
	}}
	expected := []string{"Find directive must be immediately proceeded with a Replace directive."}
	messages := linter.ProcessLine("NeverProxy google.com")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

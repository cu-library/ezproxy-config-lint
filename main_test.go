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
	messages := linter.ProcessLineAt("Title hello     ", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMissingURL(t *testing.T) {
	linter := Linter{State: State{
		Title: "A Title",
	}}
	expected := []string{"Stanza \"A Title\" has Title but no URL"}
	messages := linter.ProcessLineAt("", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMalformedURL(t *testing.T) {
	linter := Linter{State: State{
		Title:    "A Title",
		Previous: Title,
	}}
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[boo\": missing ']' in host"}
	messages := linter.ProcessLineAt("URL http://[boo", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestURLWithoutScheme(t *testing.T) {
	linter := Linter{State: State{
		Title:    "A Title",
		Previous: Title,
	}}
	expected := []string{"URL does not start with http or https"}
	messages := linter.ProcessLineAt("URL google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMalformedHost(t *testing.T) {
	linter := Linter{}
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[]w]w[ef\": invalid port \"w[ef\" after host"}
	messages := linter.ProcessLineAt("HJ []w]w[ef", "test:1")
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
		messages := linter.ProcessLineAt(line, "test:1")
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
	messages := linter.ProcessLineAt("NeverProxy google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMisstyledDirective(t *testing.T) {
	linter := Linter{State: State{}}
	expected := []string{"Title directive improperly styled as TITLE"}
	messages := linter.ProcessLineAt("TITLE Foo", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestUnknownDirective(t *testing.T) {
	linter := Linter{State: State{}}
	expected := []string{"Unknown directive FooBar"}
	messages := linter.ProcessLineAt("FooBar Baz", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

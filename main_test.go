// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"
)

func TestLineEndingInSpace(t *testing.T) {
	state := State{}
	expected := []string{"Line ends in a space or tab character"}
	messages := state.ProcessLine("Title hello     ")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMissingURL(t *testing.T) {
	state := State{
		Title: "A Title",
	}
	expected := []string{"Stanza \"A Title\" has Title but no URL"}
	messages := state.ProcessLine("")
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

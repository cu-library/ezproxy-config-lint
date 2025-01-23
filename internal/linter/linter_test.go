// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package linter

import (
	"reflect"
	"strings"
	"testing"
)

func TestLineEndingInSpace(t *testing.T) {
	linter := Linter{Whitespace: true}
	expected := []string{"Line ends in a space or tab character (L5002)"}
	messages := linter.ProcessLineAt("Title hello     ", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMissingURL(t *testing.T) {
	linter := Linter{State: State{
		Title: "A Title",
	}}
	expected := []string{"Stanza \"A Title\" has Title but no URL (L4003)"}
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
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[boo\": missing ']' in host (L3005)"}
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
	expected := []string{"URL does not start with http or https (L3006)"}
	messages := linter.ProcessLineAt("URL google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMalformedHost(t *testing.T) {
	linter := Linter{}
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[]w]w[ef\": invalid port \"w[ef\" after host (L3005)"}
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
	expected := []string{"\"Find\" directive must be immediately proceeded with a \"Replace\" directive (L4004)"}
	messages := linter.ProcessLineAt("NeverProxy google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestMisstyledDirective(t *testing.T) {
	linter := Linter{DirectiveCase: true, State: State{}}
	expected := []string{"\"TITLE\" directive does not have the right letter casing. It should be replaced by \"Title\" (L5001)"}
	messages := linter.ProcessLineAt("TITLE Foo", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestUnknownDirective(t *testing.T) {
	linter := Linter{State: State{}}
	expected := []string{"Unknown directive \"FooBar\" (L9001)"}
	messages := linter.ProcessLineAt("FooBar Baz", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %v instead of %v", messages, expected)
	}
}

func TestFindURLFromLine(t *testing.T) {
	var tests = []struct {
		line     string
		expected string
	}{
		{"Blag", ""},
		{"UR http://www.somedb.com", ""},
		{"URL http://www.somedb.com", "http://www.somedb.com"},
		{"U http://www.somedb.com", "http://www.somedb.com"},
		{"URL -Refresh a b", "b"},
		{"U -Refresh a b", "b"},
		{"URL -Redirect -Append -Encoded otherdb http://www.otherdb.com/search?q=", "http://www.otherdb.com/search?q="},
		{"U -Redirect -Append -Encoded otherdb http://www.otherdb.com/search?q=", "http://www.otherdb.com/search?q="},
		{"URL -Redirect -Append otherdb http://www.otherdb.com/search?q=", ""},
		{"URL -Redirect -Encoded otherdb http://www.otherdb.com/search?q=", ""},
		{"URL -Form=post somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
		{"U -Form=post somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
		{"URL -Form=post -RewriteHost somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
		{"U -Form=post -RewriteHost somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
	}

	for _, tt := range tests {
		urlQualifier := FindURLFromLine(tt.line)
		if urlQualifier != tt.expected {
			t.Fatalf("FindURLFromLine() fails on \"%v\", wanted \"%v\", got \"%v\".\n", tt.line, tt.expected, urlQualifier)
		}
	}
}

func TestUnclosedOptionDirectives(t *testing.T) {
	var tests = []struct {
		linter   Linter
		expected []string
	}{
		{
			Linter{
				State: State{
					Title:       "DomainCookieOnlyMissing",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionDomainCookieOnly},
				},
			},
			[]string{"Stanza \"DomainCookieOnlyMissing\" has \"Option DomainCookieOnly\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionNoCookie",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionNoCookie},
				},
			},
			[]string{"Stanza \"OptionNoCookie\" has \"Option NoCookie\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionCookiePassThrough",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionCookiePassThrough},
				},
			},
			[]string{"Stanza \"OptionCookiePassThrough\" has \"Option CookiePassThrough\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionHideEZproxy",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionHideEZproxy},
				},
			},
			[]string{"Stanza \"OptionHideEZproxy\" has \"Option HideEZproxy\" but doesn't have a " +
				"corresponding \"Option NoHideEZproxy\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionNoHttpsHyphens",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionNoHttpsHyphens},
				},
			},
			[]string{"Stanza \"OptionNoHttpsHyphens\" has \"Option NoHttpsHyphens\" but doesn't have a " +
				"corresponding \"Option HttpsHyphens\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionMetaEZproxyRewriting",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionMetaEZproxyRewriting},
				},
			},
			[]string{"Stanza \"OptionMetaEZproxyRewriting\" has \"Option MetaEZproxyRewriting\" but doesn't have a " +
				"corresponding \"Option NoMetaEZproxyRewriting\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionProxyFTP",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionProxyFTP},
				},
			},
			[]string{"Stanza \"OptionProxyFTP\" has \"Option ProxyFTP\" but doesn't have a " +
				"corresponding \"Option NoProxyFTP\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionUTF16",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionUTF16},
				},
			},
			[]string{"Stanza \"OptionUTF16\" has \"Option UTF16\" but doesn't have a " +
				"corresponding \"Option NoUTF16\" line at the end of the stanza (L4002)"},
		},
		{
			Linter{
				State: State{
					Title:       "OptionXForwardedFor",
					URL:         "https://test.com",
					OpenOptions: []Directive{OptionXForwardedFor},
				},
			},
			[]string{"Stanza \"OptionXForwardedFor\" has \"Option X-Forwarded-For\" but doesn't have a " +
				"corresponding \"Option NoX-Forwarded-For\" line at the end of the stanza (L4002)"},
		},
	}

	for _, tt := range tests {
		messages := tt.linter.ProcessLineAt("", "test:1")
		if !reflect.DeepEqual(messages, tt.expected) {
			t.Fatalf("incorrect messages %v instead of %v", messages, tt.expected)
		}
	}

}

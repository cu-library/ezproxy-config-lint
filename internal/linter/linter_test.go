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
	config := Config{Whitespace: true}
	linter := NewLinter(config)
	expected := []string{"Line ends in a space or tab character (L5002)"}
	messages := linter.ProcessLineAt("Title hello     ", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestMissingURL(t *testing.T) {
	linter := NewLinter(Config{})
	linter.State.Title = "A Title" //nolint:goconst
	expected := []string{"Stanza \"A Title\" has Title but no URL (L4003)"}
	messages := linter.ProcessLineAt("", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestMalformedURL(t *testing.T) {
	linter := NewLinter(Config{})
	linter.State.Title = "A Title"
	linter.State.Previous = Title
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[boo\": missing ']' in host (L3005)"}
	messages := linter.ProcessLineAt("URL http://[boo", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestURLWithoutScheme(t *testing.T) {
	linter := NewLinter(Config{})
	linter.State.Title = "A Title"
	linter.State.Previous = Title
	expected := []string{"URL does not start with http or https (L3006)"}
	messages := linter.ProcessLineAt("URL google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestMalformedHost(t *testing.T) {
	linter := NewLinter(Config{})
	expected := []string{"Unable to parse URL, might be malformed: parse \"http://[]w]w[ef\": invalid IP-literal (L3005)"}
	messages := linter.ProcessLineAt("HJ []w]w[ef", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
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
		result := TrailingSpaceOrTabCheck(tt.line)
		if result != tt.expected {
			t.Fatalf("TrailingSpaceOrTabCheck() fails on %q, wanted %v, got %v.\n", tt.line, tt.expected, result)
		}
	}
}

func TestMultilineDirective(t *testing.T) {
	linter := NewLinter(Config{})
	multiline := `ShibbolethMetadata \
                      -EntityID=EZproxyEntityID \
                      -File=MetadataFile \
                      -SignResponse=false -SignAssertion=true -EncryptAssertion=false \
                      -Cert=EZproxyCertNumber`
	for _, line := range strings.Split(multiline, "\n") {
		messages := linter.ProcessLineAt(line, "test:1")
		if len(messages) != 0 {
			t.Fatalf("Multiline directive was not properly processed: %q", messages)
		}
	}
	if linter.State.Previous != ShibbolethMetadata {
		t.Fatalf("Processing multiline directive did not find the correct Directive")
	}
}

func TestFindReplacePair(t *testing.T) {
	linter := NewLinter(Config{})
	linter.State.Previous = Find
	expected := []string{"\"Find\" directive must be immediately proceeded by a \"Replace\" directive (L4004)"}
	messages := linter.ProcessLineAt("NeverProxy google.com", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestMisstyledDirective(t *testing.T) {
	config := Config{DirectiveCase: true}
	linter := NewLinter(config)
	expected := []string{"\"TITLE\" directive does not have the right letter casing. It should be replaced by \"Title\" (L5001)"}
	messages := linter.ProcessLineAt("TITLE Foo", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestUnknownDirective(t *testing.T) {
	linter := NewLinter(Config{})
	expected := []string{"Unknown directive \"FooBar\" (L9001)"}
	messages := linter.ProcessLineAt("FooBar Baz", "test:1")
	if !reflect.DeepEqual(messages, expected) {
		t.Fatalf("incorrect messages %q instead of %q", messages, expected)
	}
}

func TestLabelToDirectiveKeysHaveNoTrailingSpace(t *testing.T) {
	for label := range LabelToDirective {
		if strings.HasSuffix(label, " ") {
			t.Errorf("LabelToDirective key %q has a trailing space", label)
		}
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
		{"URL -Form=post somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"}, //nolint:goconst
		{"U -Form=post somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
		{"URL -Form=post -RewriteHost somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
		{"U -Form=post -RewriteHost somedb http://www.somedb.com/login.asp", "http://www.somedb.com/login.asp"},
	}

	for _, tt := range tests {
		urlQualifier := FindURLFromLine(tt.line)
		if urlQualifier != tt.expected {
			t.Fatalf("FindURLFromLine() fails on %q, wanted %q, got %q.\n", tt.line, tt.expected, urlQualifier)
		}
	}
}

func TestUnclosedOptionDirectives(t *testing.T) {
	linter := NewLinter(Config{})
	var tests = []struct {
		title       string
		url         string
		openOptions []Directive
		expected    []string
	}{
		{
			"DomainCookieOnlyMissing",
			"https://test.com", //nolint:goconst
			[]Directive{OptionDomainCookieOnly},
			[]string{"Stanza \"DomainCookieOnlyMissing\" has \"Option DomainCookieOnly\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionNoCookie",
			"https://test.com",
			[]Directive{OptionNoCookie},
			[]string{"Stanza \"OptionNoCookie\" has \"Option NoCookie\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionCookiePassThrough",
			"https://test.com",
			[]Directive{OptionCookiePassThrough},
			[]string{"Stanza \"OptionCookiePassThrough\" has \"Option CookiePassThrough\" but doesn't have a " +
				"corresponding \"Option Cookie\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionHideEZproxy",
			"https://test.com",
			[]Directive{OptionHideEZproxy},
			[]string{"Stanza \"OptionHideEZproxy\" has \"Option HideEZproxy\" but doesn't have a " +
				"corresponding \"Option NoHideEZproxy\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionNoHttpsHyphens",
			"https://test.com",
			[]Directive{OptionNoHttpsHyphens},
			[]string{"Stanza \"OptionNoHttpsHyphens\" has \"Option NoHttpsHyphens\" but doesn't have a " +
				"corresponding \"Option HttpsHyphens\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionMetaEZproxyRewriting",
			"https://test.com",
			[]Directive{OptionMetaEZproxyRewriting},
			[]string{"Stanza \"OptionMetaEZproxyRewriting\" has \"Option MetaEZproxyRewriting\" but doesn't have a " +
				"corresponding \"Option NoMetaEZproxyRewriting\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionProxyFTP",
			"https://test.com",
			[]Directive{OptionProxyFTP},
			[]string{"Stanza \"OptionProxyFTP\" has \"Option ProxyFTP\" but doesn't have a " +
				"corresponding \"Option NoProxyFTP\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionUTF16",
			"https://test.com",
			[]Directive{OptionUTF16},
			[]string{"Stanza \"OptionUTF16\" has \"Option UTF16\" but doesn't have a " +
				"corresponding \"Option NoUTF16\" line at the end of the stanza (L4002)"},
		},
		{
			"OptionXForwardedFor",
			"https://test.com",
			[]Directive{OptionXForwardedFor},
			[]string{"Stanza \"OptionXForwardedFor\" has \"Option X-Forwarded-For\" but doesn't have a " +
				"corresponding \"Option NoX-Forwarded-For\" line at the end of the stanza (L4002)"},
		},
	}

	for _, tt := range tests {
		linter.State = NewState()
		linter.State.Title = tt.title
		linter.State.URL = tt.url
		linter.State.OpenOptions = tt.openOptions
		messages := linter.ProcessLineAt("", "test:1")
		if !reflect.DeepEqual(messages, tt.expected) {
			t.Fatalf("incorrect messages %q instead of %q", messages, tt.expected)
		}
	}
}

func TestTrimLabel(t *testing.T) {
	var tests = []struct {
		line     string
		label    string
		expected string
	}{
		{"AddUserHeader", "AddUserHeader", ""},
		{"DJ aviationweek.com", "DJ", "aviationweek.com"},
		{"HTTPHeader -request -process token", "HTTPHeader", "-request -process token"},
		{"AnonymousURL -*", "AnonymousURL", "-*"},
		{"Find name=\"landingPage\" value=\"http://", "Find", "name=\"landingPage\" value=\"http://"},
	}

	for _, tt := range tests {
		result := TrimLabel(tt.line, tt.label)
		if result != tt.expected {
			t.Fatalf("TrimLabel() fails on %q, wanted %q, got %q.\n", tt.line, tt.expected, result)
		}
	}
}

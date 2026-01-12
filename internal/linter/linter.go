// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package linter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/html"
)

const (
	DefaultBufferSize = 1 * 1024 * 1024        // 1 MiB, the default size when creating a buffer for a scanner.
	MaxBufferSize     = 5 * 1024 * 1024        // 5 MiB, the maximum size the scanner buffers can grow to.
	OCLCHTTPTimeout   = 10 * time.Second       // The timeout to set on contexts when querying the OCLC website.
	OCLCRequestDelay  = 300 * time.Millisecond // The time to wait after querying the OCLC website.
)

type State struct {
	AddUserHeaderNeedsClosing bool
	AnonymousURLNeedsClosing  bool
	OpenOptions               []Directive
	InMultiline               bool
	LastLineEmpty             bool
	OCLCTitle                 string
	Label                     string    `json:"PreviousLabel"`
	Current                   Directive `json:"-"`
	IsSeparator               bool
	Previous                  Directive `json:"PreviousDirective"`
	PreviousMultilineSegments string
	Source                    string
	ProxyHostnameEditPatterns map[string]*regexp.Regexp
	Title                     string
	URL                       string
	URLOrigin                 string
	URLAt                     string
	StanzaOrigins             map[string]string
}

type Linter struct {
	Annotate             bool
	Verbose              bool
	AdditionalPHEChecks  bool
	DirectiveCase        bool
	HTTPS                bool
	Origins              bool
	Source               bool
	Whitespace           bool
	FollowIncludeFile    bool
	IncludeFileDirectory string
	State                State
	Output               io.Writer
	PreviousTitles       map[string]string
	PreviousOrigins      map[string]string
}

func OptionPairs() map[Directive]Directive {
	return map[Directive]Directive{
		OptionDomainCookieOnly:     OptionCookie,
		OptionNoCookie:             OptionCookie,
		OptionCookiePassThrough:    OptionCookie,
		OptionHideEZproxy:          OptionNoHideEZproxy,
		OptionNoHttpsHyphens:       OptionHttpsHyphens,
		OptionMetaEZproxyRewriting: OptionNoMetaEZproxyRewriting,
		OptionProxyFTP:             OptionNoProxyFTP,
		OptionUTF16:                OptionNoUTF16,
		OptionXForwardedFor:        OptionNoXForwardedFor,
	}
}

func OpenerOptions() []Directive {
	return slices.Collect(maps.Keys(OptionPairs()))
}

func CloserOptions() []Directive {
	return slices.Collect(maps.Values(OptionPairs()))
}

var (
	URLV1Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(\S+)$`)
	URLV2Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(-Refresh )?\s*(-Redirect )?\s*(-Append -Encoded )?\s*(\S+)\s+(\S+)$`)
	URLV3Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(-Form)=([A-Za-z]+ )\s*(-RewriteHost )?\s*(\S+)\s+(\S+)$`)
)

func (l *Linter) ProcessFile(filePath string) (warningCount int, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return warningCount, err
	}
	defer f.Close()

	// If the IncludeFileDirectory was not set by the caller,
	// use the parent directory of first file the linter processes.
	if l.IncludeFileDirectory == "" {
		l.IncludeFileDirectory = filepath.Dir(filePath)
	}

	// Preallocate a buffer for the scanner.
	buf := make([]byte, DefaultBufferSize)
	// Make a scanner to go through the file line by line.
	scanner := bufio.NewScanner(f)
	// Use the buffer to store each line. The buffer can grow if needed.
	scanner.Buffer(buf, MaxBufferSize)

	// Store the line number for output.
	lineNum := 0

	// Store information about each stanza.
	l.State = State{}

	// Loop through each line in the file.
	for {
		// This hacky section is here to handle
		// the case where the config file ends without
		// an empty line.
		// If the scanner was able to advance,
		// get the line and increment the line number
		// counter.
		// If the scanner was unable to advance,
		// and the last processed line wasn't empty,
		// run the checks one last time with an
		// empty line.
		line := ""
		more := scanner.Scan()
		if more {
			// Get the string value of the current line.
			line = scanner.Text()
			// Increment the line number.
			lineNum++
		} else if l.State.LastLineEmpty {
			break
		}

		// If verbose mode is enabled, print the internal state
		// of the linter before each line.
		if l.Verbose {
			s, err := json.Marshal(l.State)
			if err != nil {
				return warningCount, err
			}
			fmt.Fprintf(l.Output, "%v\n", color.CyanString(string(s)))
		}

		at := fmt.Sprintf("%v:%v", filePath, lineNum)

		warnings := l.ProcessLineAt(line, at)
		if len(warnings) > 0 {
			warningCount += len(warnings)
			if l.State.LastLineEmpty {
				// This will print any warnings that can only be checked after a stanza is closed, and apply to the whole stanza.
				fmt.Fprintf(l.Output, "%v: %v\n", at, color.YellowString(fmt.Sprintf("↑ %v", strings.Join(warnings, ", "))))
				// If we're printing the whole file, print the empty line we just processed without any warnings.
				// This helps break up the annotated output with lines between stanzas.
				if l.Annotate && more {
					fmt.Fprintf(l.Output, "%v:\n", at)
				}
			} else {
				fmt.Fprintf(l.Output, "%v: %v %v\n", at, line, color.YellowString(fmt.Sprintf("← %v", strings.Join(warnings, ", "))))
			}
		} else if l.Annotate && more {
			fmt.Fprintf(l.Output, "%v: %v\n", at, line)
		}

		// Follow IncludeFile paths recursively.
		if l.FollowIncludeFile && l.State.Previous == IncludeFile {
			splitLine := strings.Split(line, " ")
			if len(splitLine) < 2 {
				return warningCount, fmt.Errorf("unable to find IncludeFile path on line %q", line)
			}
			includeFilePath := splitLine[1]
			// If the file path for the included file is not absolute, we should
			// join it with the IncludeFileDirectory, which has been set by the caller
			// or to the parent directory of the first file the linter processed.
			if !filepath.IsAbs(includeFilePath) {
				includeFilePath = filepath.Join(l.IncludeFileDirectory, includeFilePath)
				if l.Verbose {
					fmt.Fprintf(l.Output, "       Line: %v\n", line)
					fmt.Fprintf(l.Output, "    in file: %v\n", filePath)
					fmt.Fprintf(l.Output, "resolves to: %v\n", includeFilePath)
				}
			}

			includeFileWarningCount, err := l.ProcessFile(includeFilePath)
			if err != nil {
				fmt.Fprintf(l.Output, "Error encountered when processing line %q.\n", line)
				return warningCount, err
			}
			warningCount += includeFileWarningCount
		}
	}

	// If the scanner encountered any errors, report them to the caller.
	if err := scanner.Err(); err != nil {
		return warningCount, err
	}
	return warningCount, nil
}

func (l *Linter) ProcessLineAt(line, at string) (m []string) {
	// Get the OptionPairs which need to be closed.
	optionPairs := OptionPairs()
	openers := OpenerOptions()
	closers := CloserOptions()

	// Initialize maps if they are still nil.
	if l.PreviousTitles == nil {
		l.PreviousTitles = make(map[string]string)
	}
	if l.PreviousOrigins == nil {
		l.PreviousOrigins = make(map[string]string)
	}
	if l.State.ProxyHostnameEditPatterns == nil {
		l.State.ProxyHostnameEditPatterns = make(map[string]*regexp.Regexp)
	}
	if l.State.StanzaOrigins == nil {
		l.State.StanzaOrigins = make(map[string]string)
	}

	// Does the line end in a space or tab character?
	if l.Whitespace && TrailingSpaceOrTabCheck(line) {
		m = append(m, "Line ends in a space or tab character (L5002)")
	}

	// Trim leading and trailing spaces to ensure the rest of the linting
	// is uniform.
	line = strings.TrimSpace(line)

	// Is the line empty, or an empty comment?
	// If so, we're at the end of the stanza.
	if line == "" || line == "#" {
		if l.State.Title != "" && l.State.URL == "" && !l.State.IsSeparator {
			m = append(m, fmt.Sprintf("Stanza %q has Title but no URL (L4003)", l.State.Title))
		}
		if l.State.AddUserHeaderNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza %q uses AddUserHeader but doesn't have a corresponding \"AddUserHeader\" "+
				"line at the end of the stanza (L4005)", l.State.Title))
		}
		if l.State.AnonymousURLNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza %q has AnonymousURL but doesn't have a corresponding \"AnonymousURL -*\" "+
				"line at the end of the stanza (L4001)", l.State.Title))
		}
		if len(l.State.OpenOptions) != 0 {
			for _, option := range l.State.OpenOptions {
				m = append(m, fmt.Sprintf("Stanza %q has %q but doesn't have a "+
					"corresponding %q line at the end of the stanza (L4002)", l.State.Title, option, optionPairs[option]))
			}
		}

		// If present, add the stored URL origin to the PreviousOrigins map.
		if l.State.URLOrigin != "" {
			l.PreviousOrigins[l.State.URLOrigin] = l.State.URLAt
		}

		// Copy the origins from this stanza to the PreviousOrigins map.
		maps.Copy(l.PreviousOrigins, l.State.StanzaOrigins)

		// Reset the stanza state.
		l.State = State{LastLineEmpty: true}

		return m
	}

	l.State.LastLineEmpty = false

	// Is the line a comment?
	if strings.HasPrefix(line, "#") {
		if l.Source && strings.HasPrefix(line, "# Source - ") {
			source, oclcTitle, err := processSourceLine(line)
			if err != nil {
				m = append(m, fmt.Sprintf("Error processsing Source line (L9003): %v", err))
			} else {
				l.State.Source = source
				l.State.OCLCTitle = oclcTitle
			}
		}
		return m
	}

	// Is the line part of a multiline string?
	if strings.HasSuffix(line, "\\") {
		l.State.PreviousMultilineSegments += strings.TrimSuffix(line, "\\")
		l.State.InMultiline = true
		return m
	} else if l.State.InMultiline {
		line = l.State.PreviousMultilineSegments + line
		l.State.PreviousMultilineSegments = ""
	}

	// Line isn't a comment or empty.

	// Reset the IsSeparator flag to false.
	l.State.IsSeparator = false

	// Split the line by spaces to find the label.
	split := strings.Split(line, " ")
	label := split[0]

	// Option directives have two parts.
	if label == "Option" {
		if len(split) != 2 {
			m = append(m, "Option directive not in the form \"Option OPTIONNAME\" (L3008)")
			return m
		}
		label = line
	}

	// Find the Directive which matches this label.
	directive, ok := LabelToDirective[label]
	if !ok {
		directive, ok = LowercaseLabelToDirective[strings.ToLower(label)]
		if !ok {
			m = append(m, fmt.Sprintf("Unknown directive %q (L9001)", label))
			return m
		}
		if l.DirectiveCase {
			m = append(m, fmt.Sprintf("%q directive does not have the right letter casing. It should be replaced by %q (L5001)", label, directive))
		}
	}
	l.State.Current = directive
	l.State.Label = label

	// Short-circuit check for Find/Replace pairs.
	// Without this, we would need to check that the previous
	// directive was not Find on every directive other than Replace.
	if l.State.Previous == Find && directive != Replace {
		m = append(m, "\"Find\" directive must be immediately proceeded with a \"Replace\" directive (L4004)")
	}

	// Special case for defensive OptionCookie.
	// Return early if we see an OptionCookie prior to other opening directives
	// that require an OptionCookie closer.
	if directive == OptionCookie && l.State.Title == "" {
		returnEarly := true

		// This is not very efficient, but hopefully this is not a hot path.
		opprs := OptionPairs()

		for _, v := range l.State.OpenOptions {
			if opprs[v] == OptionCookie {
				returnEarly = false
				break
			}
		}
		if returnEarly {
			l.State.Previous = OptionCookie
			return
		}
	}

	// Process Option Pair directives.
	if slices.Contains(openers, directive) {
		m = append(m, l.ProcessOptionOpener(line)...)
	} else if slices.Contains(closers, directive) {
		m = append(m, l.ProcessOptionCloser(line)...)
	}

	// Process other directives.
	switch directive {
	case ProxyHostnameEdit:
		m = append(m, l.ProcessProxyHostnameEdit(line)...)
	case AddUserHeader:
		m = append(m, l.ProcessAddUserHeader(line)...)
	case AnonymousURL:
		m = append(m, l.ProcessAnonymousURL(line)...)
	case Title:
		m = append(m, l.ProcessTitle(line, at)...)
	case Description:
		m = append(m, l.ProcessDescription(line, at)...)
	case URL:
		m = append(m, l.ProcessURL(line, at)...)
	case Host, HostJavaScript:
		m = append(m, l.ProcessHostAndHostJavaScript(line, at)...)
	case Domain, DomainJavaScript:
		m = append(m, l.ProcessDomainAndDomainJavaScript(line)...)
	}
	l.State.Previous = directive
	return m
}

// ProcessOptionOpener processes the line containing an Option which will need to be closed later.
func (l *Linter) ProcessOptionOpener(line string) (m []string) {
	allowedPreviousDirectives := []Directive{
		Undefined,
		Group,
		DbVar,
		DbVar0,
		DbVar1,
		DbVar2,
		DbVar3,
		DbVar4,
		DbVar5,
		DbVar6,
		DbVar7,
		DbVar8,
		DbVar9,
		HTTPMethod,
		AddUserHeader,
		AnonymousURL,
		OptionCookie,
	}
	allowedPreviousDirectives = append(allowedPreviousDirectives, OpenerOptions()...)
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("%q directive is out of order, previous directive: %q (L1005)", l.State.Current, l.State.Previous))
	}
	l.State.OpenOptions = append(l.State.OpenOptions, l.State.Current)
	return m
}

// ProcessOptionCloser processes the line containing an Option which closes an 'Opener' option.
func (l *Linter) ProcessOptionCloser(line string) (m []string) {
	optionPairs := OptionPairs()
	allowedPreviousDirectives := []Directive{
		DbVar,
		URL,
		Host,
		HostJavaScript,
		Domain,
		DomainJavaScript,
		Replace,
		AddUserHeader,
		AnonymousURL,
		NeverProxy,
	}
	allowedPreviousDirectives = append(allowedPreviousDirectives, CloserOptions()...)
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("%q directive is out of order, previous directive: %q (L1006)", l.State.Current, l.State.Previous))
	}
	l.State.OpenOptions = slices.DeleteFunc(l.State.OpenOptions, func(d Directive) bool {
		return optionPairs[d] == l.State.Current
	})
	return m
}

// ProcessProxyHostnameEdit processes the line containing the ProxyHostnameEdit directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/ProxyHostnameEdit
func (l *Linter) ProcessProxyHostnameEdit(line string) (m []string) {
	allowedPreviousDirectives := []Directive{
		Undefined,
		Group,
		HTTPMethod,
		Cookie,
		DbVar,
		DbVar0,
		DbVar1,
		DbVar2,
		DbVar3,
		DbVar4,
		DbVar5,
		DbVar6,
		DbVar7,
		DbVar8,
		DbVar9,
		AddUserHeader,
		AnonymousURL,
		OptionCookie,
		ProxyHostnameEdit,
	}
	allowedPreviousDirectives = append(allowedPreviousDirectives, OpenerOptions()...)
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("\"ProxyHostnameEdit\" directive is out of order, previous directive: %q (L1008)", l.State.Previous))
	}

	// Does the ProxyHostnameEdit line have both a find and replace?
	findReplacePair := strings.Split(TrimLabel(line, l.State.Label), " ")
	if len(findReplacePair) != 2 {
		m = append(m, "\"ProxyHostnameEdit\" directive must have both a find and replace qualifier (L3001)")
		return m
	}

	if l.AdditionalPHEChecks {
		find, found := strings.CutSuffix(findReplacePair[0], "$")
		if !found {
			m = append(m, "Find part of \"ProxyHostnameEdit\" directive should end with a $ (L3002)")
		}

		if strings.ReplaceAll(find, ".", "-") != findReplacePair[1] {
			m = append(m, "Replace part of \"ProxyHostnameEdit\" directive is malformed (L3003)")
		}

		for pattern, re := range l.State.ProxyHostnameEditPatterns {
			if re.MatchString(find) {
				m = append(m, fmt.Sprintf("\"ProxyHostnameEdit\" domains should be placed in deepest-to-shallowest order, previous pattern: %q (L1009)", pattern))
			}
		}

		// For every pattern we see, create a regexp to match any subdomains.
		re := regexp.MustCompile(`[.]` + regexp.QuoteMeta(find) + `$`)
		l.State.ProxyHostnameEditPatterns[find] = re
	}
	return m
}

// ProcessAddUserHeader processes the line containing the AddUserHeader directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/AddUserHeader
func (l *Linter) ProcessAddUserHeader(line string) (m []string) {
	if TrimLabel(line, l.State.Label) == "" {
		allowedPreviousDirectives := []Directive{
			URL,
			Host,
			HostJavaScript,
			Domain,
			DomainJavaScript,
			Replace,
			AnonymousURL,
			NeverProxy,
		}
		allowedPreviousDirectives = append(allowedPreviousDirectives, CloserOptions()...)
		if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
			m = append(m, fmt.Sprintf("\"AddUserHeader\" directive with no qualifiers is out of order, previous directive: %q (L1011)", l.State.Previous))
		}
		l.State.AddUserHeaderNeedsClosing = false
	} else {
		allowedPreviousDirectives := []Directive{
			Undefined,
			Group,
			HTTPMethod,
			Cookie,
			DbVar,
			DbVar0,
			DbVar1,
			DbVar2,
			DbVar3,
			DbVar4,
			DbVar5,
			DbVar6,
			DbVar7,
			DbVar8,
			DbVar9,
			AddUserHeader,
			AnonymousURL,
			OptionCookie,
			ProxyHostnameEdit,
		}
		allowedPreviousDirectives = append(allowedPreviousDirectives, OpenerOptions()...)
		if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
			m = append(m, fmt.Sprintf("\"AddUserHeader\" directive is out of order, previous directive: %q (L1012)", l.State.Previous))
		}
		l.State.AddUserHeaderNeedsClosing = true
	}
	return m
}

// ProcessAnonymousURL processes the line containing the AnonymousURL directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/AnonymousURL
func (l *Linter) ProcessAnonymousURL(line string) (m []string) {
	if TrimLabel(line, l.State.Label) == "-*" {
		allowedPreviousDirectives := []Directive{
			URL,
			Host,
			HostJavaScript,
			Domain,
			DomainJavaScript,
			Replace,
			AddUserHeader,
			NeverProxy,
		}
		allowedPreviousDirectives = append(allowedPreviousDirectives, CloserOptions()...)
		if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
			m = append(m, fmt.Sprintf("\"AnonymousURL -*\" directive is out of order, previous directive: %q (L1003)", l.State.Previous))
		}
		l.State.AnonymousURLNeedsClosing = false
	} else {
		allowedPreviousDirectives := []Directive{
			Undefined,
			Group,
			HTTPMethod,
			Cookie,
			DbVar,
			DbVar0,
			DbVar1,
			DbVar2,
			DbVar3,
			DbVar4,
			DbVar5,
			DbVar6,
			DbVar7,
			DbVar8,
			DbVar9,
			AddUserHeader,
			AnonymousURL,
			ProxyHostnameEdit,
		}
		allowedPreviousDirectives = append(allowedPreviousDirectives, OpenerOptions()...)
		if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
			m = append(m, fmt.Sprintf("\"AnonymousURL\" directive is out of order, previous directive: %q (L1004)", l.State.Previous))
		}
		l.State.AnonymousURLNeedsClosing = true
	}
	return m
}

// ProcessTitle processes the line containing the Title directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Title
func (l *Linter) ProcessTitle(line, at string) (m []string) {
	allowedPreviousDirectives := []Directive{
		Undefined,
		Group,
		HTTPMethod,
		AddUserHeader,
		AnonymousURL,
		ProxyHostnameEdit,
		Referer,
		Cookie,
		DbVar,
		DbVar0,
		DbVar1,
		DbVar2,
		DbVar3,
		DbVar4,
		DbVar5,
		DbVar6,
		DbVar7,
		DbVar8,
		DbVar9,
		OptionEbraryUnencodedTokens,
		OptionCookie,
	}
	allowedPreviousDirectives = append(allowedPreviousDirectives, OpenerOptions()...)
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("\"Title\" directive is out of order, previous directive: %q (L1001)", l.State.Previous))
	}
	// If the previous AnonymousURL directive was `AnonymousURL -*`, that's a problem.
	if !l.State.AnonymousURLNeedsClosing && l.State.Previous == AnonymousURL {
		m = append(m, fmt.Sprintf("\"Title\" directive is out of order, previous directive: %q (L1001)", l.State.Previous))
	}

	if l.State.Title != "" {
		m = append(m, "Duplicate \"Title\" directive in stanza (L2001)")
	}
	l.State.Title = TrimLabel(line, l.State.Label)
	titleSeenAt, titleSeen := l.PreviousTitles[l.State.Title]
	if titleSeen {
		m = append(m, fmt.Sprintf("\"Title\" directive value already seen at %q (L2004)", titleSeenAt))
	} else {
		l.PreviousTitles[l.State.Title] = at
	}

	titleWithHideRemoved := strings.TrimPrefix(l.State.Title, "-Hide ")
	if l.State.OCLCTitle != "" && l.State.Title != l.State.OCLCTitle && titleWithHideRemoved != l.State.OCLCTitle {
		m = append(m, "Source title doesn't match, you might need to update this stanza (L9002)")
	}
	return m
}

// ProcessDescription processes the line containing a description directive.
// OCLC documention:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Description
func (l *Linter) ProcessDescription(line, at string) (m []string) {
	allowedPreviousDirectives := []Directive{
		Title,
		Description,
	}
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("\"Description\" directive is out of order, previous directive: %q (L1013)", l.State.Previous))
	}

	// From the documentation: "EZproxy supports a special database stanza comprised of only a
	// single Title directive and one or more Description directives."
	// That special stanza designation is stored in l.State.IsSeparator.
	l.State.IsSeparator = true
	return m
}

// ProcessHostandHostJavaScript processes the line containing a Host or HostJavaScript directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Host_H
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/HostJavaScript_HJ
func (l *Linter) ProcessHostAndHostJavaScript(line, at string) (m []string) {
	trimmed := TrimLabel(line, l.State.Label)
	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		m = append(m, fmt.Sprintf("Unable to parse URL, might be malformed: %v (L3005)", err))
		return
	}
	if parsedURL.Host == "" {
		// This H/HJ line did not have a scheme.
		// Per the EZproxy docs, http:// is assumed.
		parsedURL, err = url.Parse("http://" + trimmed)
		if err != nil {
			m = append(m, fmt.Sprintf("Unable to parse URL, might be malformed: %v (L3005)", err))
			return
		}
	}
	origin := fmt.Sprintf("%v://%v", parsedURL.Scheme, parsedURL.Host)
	// Check the origin against origins seen in other stanzas.
	originSeenAt, originSeen := l.PreviousOrigins[origin]
	if originSeen {
		m = append(m, fmt.Sprintf("Origin already seen at %q (L2002)", originSeenAt))
	}
	// Check the origin against origins seen in the current stanza.
	originSeenAt, originSeen = l.State.StanzaOrigins[origin]
	if l.Origins && originSeen {
		m = append(m, fmt.Sprintf("Origin already seen at %q (L2005)", originSeenAt))
	}
	if !originSeen {
		l.State.StanzaOrigins[origin] = at
	}

	return m
}

// ProcessDomainAndDomainJavaScript processes the line containing a Domain or DomainJavaScript directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Domain_D
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/DomainJavaScript_DJ
func (l *Linter) ProcessDomainAndDomainJavaScript(line string) (m []string) {
	parsedURL, err := url.Parse(TrimLabel(line, l.State.Label))
	if err != nil {
		m = append(m, fmt.Sprintf("Unable to parse URL, might be malformed: %v (L3005)", err))
		return
	}
	if parsedURL.Scheme != "" || strings.Contains(parsedURL.Path, "/") {
		m = append(m, "Domain and DomainJavaScript directives should only specify domains (L3004)")
	}
	return m
}

// ProcessURL processes the line containing a URL directive.
// OCLC documention:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_1
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_2
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_3
func (l *Linter) ProcessURL(line, at string) (m []string) {
	allowedPreviousDirectives := []Directive{
		AllowVars,
		Description,
		EBLSecret,
		EbrarySite,
		EncryptVar,
		HTTPHeader,
		MimeFilter,
		Title,
	}
	if !slices.Contains(allowedPreviousDirectives, l.State.Previous) {
		m = append(m, fmt.Sprintf("\"URL\" directive is out of order, previous directive: %q (L1002)", l.State.Previous))
	}

	if l.State.Title == "" {
		m = append(m, "\"URL\" directive is before \"Title\" directive (L1010)")
	}
	if l.State.URL != "" {
		m = append(m, "Duplicate \"URL\" directive in stanza (L2003)")
	}

	urlFromLine := FindURLFromLine(line)
	if urlFromLine == "" {
		m = append(m, "\"URL\" directive is not in the right format. (L3009)")
	}
	l.State.URL = urlFromLine
	parsedURL, err := url.Parse(l.State.URL)
	if err != nil {
		m = append(m, fmt.Sprintf("Unable to parse URL, might be malformed: %v (L3005)", err))
		return
	}
	if parsedURL.Host == "" {
		m = append(m, "URL does not start with http or https (L3006)")
		return
	}
	if l.HTTPS && parsedURL.Scheme != "https" {
		m = append(m, "URL is not using HTTPS scheme (L3007)")
	}
	// According to the EZproxy docs at
	// https://help.oclc.org/Library_Management/EZproxy/EZproxy_configuration/Starting_point_URLs_and_config_txt,
	// URL, Host, and HostJavaScript directives are checked for starting point URLs.
	// So many stanzas duplicate the URL in an HJ or H line that adding the URL's
	// origin to PreviousOrigins would add a lot of noise to the output.
	// Instead, we add the URL's origin and the filename/line combination (the 'at')
	// to the Linter's State so that we can add it to PreviousOrigins when we're done
	// processing the stanza.
	l.State.URLOrigin = fmt.Sprintf("%v://%v", parsedURL.Scheme, parsedURL.Host)
	l.State.URLAt = at
	originSeenAt, originSeen := l.PreviousOrigins[l.State.URLOrigin]
	if originSeen {
		m = append(m, fmt.Sprintf("Origin already seen at %q (L2002)", originSeenAt))
	}
	return m
}

func FindURLFromLine(line string) (url string) {
	urlV1Match := URLV1Regex.FindStringSubmatch(line)
	if urlV1Match != nil {
		return urlV1Match[len(urlV1Match)-1]
	}
	urlV2Match := URLV2Regex.FindStringSubmatch(line)
	if urlV2Match != nil {
		return urlV2Match[len(urlV2Match)-1]
	}
	urlV3Match := URLV3Regex.FindStringSubmatch(line)
	if urlV3Match != nil {
		return urlV3Match[len(urlV3Match)-1]
	}
	return ""
}

func TrailingSpaceOrTabCheck(line string) bool {
	if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
		return true
	}
	return false
}

func TrimLabel(line, label string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, label))
}

func TrimDirective(line string, directiveToTrim Directive) string {
	for label, directive := range LabelToDirective {
		if directive == directiveToTrim {
			line = strings.TrimPrefix(line, label+" ")
		}
	}
	return strings.TrimSpace(line)
}

func processSourceLine(sourceLine string) (string, string, error) {
	oclcTitle := ""
	splitSourceLine := strings.Split(sourceLine, " ")
	if len(splitSourceLine) != 4 {
		return "", "", errors.New("source line is malformed")
	}
	source := splitSourceLine[3]
	parsedSourceURL, err := url.Parse(source)
	if err != nil {
		return "", "", err
	}
	if parsedSourceURL.Scheme != "https" {
		return "", "", errors.New("source line isn't using https")
	}
	if parsedSourceURL.Host != "help.oclc.org" {
		return "", "", errors.New("source line isn't pointing to OCLC")
	}
	// Make a GET request, waiting no more than 10 second for the results.
	ctx, cancel := context.WithTimeout(context.Background(), OCLCHTTPTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedSourceURL.String(), nil)
	if err != nil {
		return "", "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	time.Sleep(OCLCRequestDelay)
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", "", err
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				buf := make([]byte, DefaultBufferSize)
				scanner := bufio.NewScanner(strings.NewReader(n.FirstChild.Data))
				scanner.Buffer(buf, MaxBufferSize)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, "Title ") || strings.HasPrefix(line, "T ") {
						oclcTitle = TrimDirective(line, Title)
						break
					}
				}
				if err := scanner.Err(); err != nil {
					log.Printf("Error scanning OCLC stanza source: %v\n", err)
				}
			}
		}
		for c := n.FirstChild; c != nil && oclcTitle == ""; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return source, oclcTitle, nil
}

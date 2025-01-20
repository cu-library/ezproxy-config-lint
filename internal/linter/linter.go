// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package linter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/html"
)

type State struct {
	AnonymousURLNeedsClosing  bool
	CookieOptionNeedsClosing  bool
	InMultiline               bool
	LastLineEmpty             bool
	OCLCTitle                 string
	Current                   Directive
	Previous                  Directive
	PreviousMultilineSegments string
	Source                    string
	Title                     string
	URL                       string
	ProxyHostnameEditDepth    int
}

type Linter struct {
	Annotate             bool
	Verbose              bool
	Whitespace           bool
	DirectiveCase        bool
	AdditionalPHEChecks  bool
	HTTPS                bool
	FollowIncludeFile    bool
	IncludeFileDirectory string
	State                State
	Output               io.Writer
	PreviousTitles       map[string]string
	PreviousOrigins      map[string]string
}

var URLV1Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(\S+)$`)
var URLV2Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(-Refresh )?\s*(-Redirect )?\s*(-Append -Encoded )?\s*(\S+)\s+(\S+)$`)
var URLV3Regex = regexp.MustCompile(`(?i)^U(RL)?\s+(-Form)=([A-Za-z]+ )\s*(-RewriteHost )?\s*(\S+)\s+(\S+)$`)

func (l *Linter) ProcessFile(filePath string) (warningCount int, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return warningCount, err
	}
	defer f.Close()

	// Make a buffer of about 1 MB in size.
	buf := make([]byte, 1048576)
	// Make a scanner to go through the file line by line.
	scanner := bufio.NewScanner(f)
	// Use the buffer to store each line, but grow the buffer to about 5MB if required.
	// 5MB line is a huge line.
	scanner.Buffer(buf, 5242880)

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
			fmt.Fprintf(l.Output, "%+v\n", l.State)
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
				return warningCount, fmt.Errorf("unable to find IncludeFile path on line \"%v\"", line)
			}
			includeFilePath := splitLine[1]
			help := ""
			// If the file path for the included file is not absolute, we should
			// join it with the IncludeFileDirectory or the parent directory
			// of the config file.
			if !filepath.IsAbs(includeFilePath) {
				if l.IncludeFileDirectory != "" {
					includeFilePath = filepath.Join(l.IncludeFileDirectory, includeFilePath)
					help = fmt.Sprintf("The '-includefile-directory' option was used, joined %v with %v", l.IncludeFileDirectory, includeFilePath)
				} else {
					filePathDir := filepath.Dir(filePath)
					includeFilePath = filepath.Join(filePathDir, includeFilePath)
					help += fmt.Sprintf("This IncludeFile directive is in a config file at this path: %v\n", filePath)
					help += fmt.Sprintf("            The IncludeFile directive resolves to this path: %v", includeFilePath)
				}
			}

			includeFileWarningCount, err := l.ProcessFile(includeFilePath)
			if err != nil {
				// Help people debug errors with IncludeFile parent directories.
				log.Printf("Error encountered when processing line \"%v\".\n", line)
				log.Println(help)
				if l.IncludeFileDirectory == "" {
					log.Println("You might want to try the '-includefile-directory' option.")
				}
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
	// Initialize maps if they are still nil.
	if l.PreviousTitles == nil {
		l.PreviousTitles = make(map[string]string)
	}
	if l.PreviousOrigins == nil {
		l.PreviousOrigins = make(map[string]string)
	}

	// Does the line end in a space or tab character?
	if l.Whitespace && TrailingSpaceOrTabCheck(line) {
		m = append(m, "Line ends in a space or tab character (L5002)")
	}

	// Trim leading and trailing spaces to ensure the rest of the linting
	// is uniform.
	line = strings.TrimSpace(line)

	// Is the line empty, or an empty comment?
	if line == "" || line == "#" {
		if l.State.Title != "" && l.State.URL == "" {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has Title but no URL (L4003)", l.State.Title))
		}
		if l.State.AnonymousURLNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has AnonymousURL but doesn't have a corresponding \"AnonymousURL -*\" "+
				"line at the end of the stanza (L4001)", l.State.Title))
		}
		if l.State.CookieOptionNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has \"Option DomainCookieOnly\" or \"Option CookiePassthrough\" "+
				"but doesn't have a corresponding \"Option Cookie\" line at the end of the stanza (L4002)", l.State.Title))
		}
		// Reset the stanza state.
		l.State = State{LastLineEmpty: true}

		return m
	} else {
		l.State.LastLineEmpty = false
	}

	// Is the line a comment?
	if strings.HasPrefix(line, "#") {
		if strings.HasPrefix(line, "# Source - ") {
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
	// Split the line by spaces to find the label.
	split := strings.Split(line, " ")
	label := split[0]

	// Option directives have two parts.
	if label == "Option" {
		if len(split) != 2 {
			m = append(m, "Option directive not in the form Option OPTIONNAME (L3008)")
			return m
		}
		label = line
	}

	// Find the Directive which matches this label.
	directive, ok := LabelToDirective[label]
	if !ok {
		directive, ok = LowercaseLabelToDirective[strings.ToLower(label)]
		if !ok {
			m = append(m, fmt.Sprintf("Unknown directive \"%v\" (L9001)", label))
			return m
		}
		if l.DirectiveCase {
			m = append(m, fmt.Sprintf("\"%v\" directive does not have the right letter casing. It should be replaced by \"%v\" (L5001)", label, directive))
		}
	}
	l.State.Current = directive

	// Short-circuit check for Find/Replace pairs.
	// Without this, we would need to check that the previous
	// directive was not Find on every directive other than Replace.
	if l.State.Previous == Find && directive != Replace {
		m = append(m, "Find directive must be immediately proceeded with a Replace directive (L4004)")
	}

	// Process each directive.
	switch directive {
	case OptionCookie:
		m = append(m, l.ProcessOptionCookie(line)...)
	case OptionCookiePassThrough:
		m = append(m, l.ProcessOptionCookiePassThrough(line)...)
	case OptionDomainCookieOnly:
		m = append(m, l.ProcessOptionDomainCookieOnly(line)...)
	case ProxyHostnameEdit:
		m = append(m, l.ProcessProxyHostnameEdit(line)...)
	case AnonymousURL:
		m = append(m, l.ProcessAnonymousURL(line)...)
	case Title:
		m = append(m, l.ProcessTitle(line, at)...)
	case URL:
		m = append(m, l.ProcessURL(line)...)
	case Host, HostJavaScript:
		m = append(m, l.ProcessHostAndHostJavaScript(line, at)...)
	case Domain, DomainJavaScript:
		m = append(m, l.ProcessDomainAndDomainJavaScript(line)...)
	}
	l.State.Previous = directive
	return m
}

// ProcessOptionCookie processes the line containing the Option Cookie directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Option_Cookie_Option_DomainCookieOnly_Option_NoCookie_Option_CookiePassThrough
func (l *Linter) ProcessOptionCookie(line string) (m []string) {
	if l.State.CookieOptionNeedsClosing {
		switch l.State.Previous {
		case URL, Host, HostJavaScript, Domain, DomainJavaScript, Replace, AnonymousURL, OptionNoXForwardedFor:
			// OptionCookie, when closing a stanza, is allowed after these directives.
		default:
			m = append(m, fmt.Sprintf("Option Cookie directive is out of order, previous directive: \"%v\" (L1011)", l.State.Previous))
		}
		l.State.CookieOptionNeedsClosing = false
	} else {
		switch l.State.Previous {
		case Undefined, Group, HTTPMethod, OptionXForwardedFor, AnonymousURL:
			// OptionCookie is allowed after these directives.
		default:
			m = append(m, fmt.Sprintf("Option Cookie directive is out of order, previous directive: \"%v\" (L1005)", l.State.Previous))
		}
	}
	return m
}

// ProcessOptionCookiePassThrough processes the line containing the Option CookiePassThrough directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Option_Cookie_Option_DomainCookieOnly_Option_NoCookie_Option_CookiePassThrough
func (l *Linter) ProcessOptionCookiePassThrough(line string) (m []string) {
	switch l.State.Previous {
	case Undefined, Group, HTTPMethod, OptionXForwardedFor, AnonymousURL:
		// OptionCookiePassThrough is allowed after these directives.
	default:
		m = append(m, fmt.Sprintf("Option CookiePassThrough directive is out of order, previous directive: \"%v\" (L1006)", l.State.Previous))
	}
	l.State.CookieOptionNeedsClosing = true
	return m
}

// ProcessOptionDomainCookieOnly processes the line containing the Option DomainCookieOnly directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Option_Cookie_Option_DomainCookieOnly_Option_NoCookie_Option_CookiePassThrough
func (l *Linter) ProcessOptionDomainCookieOnly(line string) (m []string) {
	switch l.State.Previous {
	case Undefined, Group, HTTPMethod, OptionXForwardedFor, AnonymousURL:
		// OptionDomainCookieOnly is allowed after these directives.
	default:
		m = append(m, fmt.Sprintf("Option DomainCookieOnly directive is out of order, previous directive: \"%v\" (L1007)", l.State.Previous))
	}
	l.State.CookieOptionNeedsClosing = true
	return m
}

// ProcessProxyHostnameEdit processes the line containing the ProxyHostnameEdit directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/ProxyHostnameEdit
func (l *Linter) ProcessProxyHostnameEdit(line string) (m []string) {
	switch l.State.Previous {
	case Undefined, Group, HTTPMethod,
		OptionCookie, OptionCookiePassThrough, OptionDomainCookieOnly, OptionXForwardedFor,
		AnonymousURL, ProxyHostnameEdit:
		// ProxyHostnameEdit is allowed after these directives.
	default:
		m = append(m, fmt.Sprintf("ProxyHostnameEdit directive is out of order, previous directive: \"%v\" (L1008)", l.State.Previous))
	}
	// Does the ProxyHostnameEdit line have both a find and replace?
	findReplacePair := strings.Split(TrimDirective(line, l.State.Current), " ")
	if len(findReplacePair) != 2 {
		m = append(m, "ProxyHostnameEdit directive must have both a find and replace qualifier (L3001)")
		return m
	}
	if l.AdditionalPHEChecks {
		find, found := strings.CutSuffix(findReplacePair[0], "$")
		if !found {
			m = append(m, "Find part of ProxyHostnameEdit directive should end with a $ (L3002)")
		}
		if strings.ReplaceAll(find, ".", "-") != findReplacePair[1] {
			m = append(m, "Replace part of ProxyHostnameEdit directive is malformed (L3003)")
		}
		depth := strings.Count(find, ".") + 1
		if l.State.ProxyHostnameEditDepth != 0 && l.State.ProxyHostnameEditDepth < depth {
			m = append(m, "ProxyHostnameEdit domains should be placed in deepest-to-shallowest order (L1009)")
		}
		l.State.ProxyHostnameEditDepth = depth
	}
	return m
}

// ProcessAnonymousURL processes the line containing the AnonymousURL directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/AnonymousURL
func (l *Linter) ProcessAnonymousURL(line string) (m []string) {
	if TrimDirective(line, l.State.Current) == "-*" {
		switch l.State.Previous {
		case URL, Host, HostJavaScript, Domain, DomainJavaScript, Replace,
			NeverProxy, OptionCookie, OptionNoXForwardedFor:
			// AnonymousURL is allowed after these directives.
		default:
			m = append(m, fmt.Sprintf("AnonymousURL -* directive is out of order, previous directive: \"%v\" (L1003)", l.State.Previous))
		}
		l.State.AnonymousURLNeedsClosing = false
	} else {
		switch l.State.Previous {
		case Undefined, Group, HTTPMethod,
			OptionCookie, OptionCookiePassThrough, OptionDomainCookieOnly,
			OptionXForwardedFor, ProxyHostnameEdit, AnonymousURL:
		default:
			m = append(m, fmt.Sprintf("AnonymousURL directive is out of order, previous directive: \"%v\" (L1004)", l.State.Previous))
		}
		l.State.AnonymousURLNeedsClosing = true
	}
	return m
}

// ProcessTitle processes the line containing the Title directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Title
func (l *Linter) ProcessTitle(line, at string) (m []string) {
	switch l.State.Previous {
	case Undefined, Group, HTTPMethod,
		OptionCookiePassThrough, OptionDomainCookieOnly, OptionXForwardedFor, Cookie,
		ProxyHostnameEdit, Referer, AddUserHeader, OptionEbraryUnencodedTokens:
		// Title is allowed after these directives.
	case OptionCookie:
		if !l.State.CookieOptionNeedsClosing {
			m = append(m, fmt.Sprintf("Title directive is out of order, previous directive: \"%v\" (L1001)", l.State.Previous))
		}
	case AnonymousURL:
		if !l.State.AnonymousURLNeedsClosing {
			m = append(m, fmt.Sprintf("Title directive is out of order, previous directive: \"%v\" (L1001)", l.State.Previous))
		}
	default:
		m = append(m, fmt.Sprintf("Title directive is out of order, previous directive: \"%v\" (L1001)", l.State.Previous))
	}
	if l.State.Title != "" {
		m = append(m, "Duplicate Title directive in stanza (L2001)")
	}
	l.State.Title = TrimDirective(line, l.State.Current)
	titleSeenAt, titleSeen := l.PreviousTitles[l.State.Title]
	if titleSeen {
		m = append(m, fmt.Sprintf("Title value already seen at \"%v\" (L2004)", titleSeenAt))
	} else {
		l.PreviousTitles[l.State.Title] = at
	}
	if l.State.OCLCTitle != "" && l.State.Title != l.State.OCLCTitle {
		m = append(m, "Source title doesn't match, you might need to update this stanza (L9002)")
	}
	return m
}

// ProcessHostandHostJavaScript processes the line containing a Host or HostJavaScript directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Host_H
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/HostJavaScript_HJ
func (l *Linter) ProcessHostAndHostJavaScript(line, at string) (m []string) {
	trimmed := TrimDirective(line, l.State.Current)
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
	originSeenAt, originSeen := l.PreviousOrigins[origin]
	if originSeen {
		m = append(m, fmt.Sprintf("Origin already seen at \"%v\" (L2002)", originSeenAt))
	} else {
		l.PreviousOrigins[origin] = at
	}
	return m
}

// ProcessDomainAndDomainJavaScript processes the line containing a Domain or DomainJavaScript directive.
// OCLC documentation:
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Domain_D
// https://help.oclc.org/Library_Management/EZproxy/Configure_resources/DomainJavaScript_DJ
func (l *Linter) ProcessDomainAndDomainJavaScript(line string) (m []string) {
	parsedURL, err := url.Parse(TrimDirective(line, l.State.Current))
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
func (l *Linter) ProcessURL(line string) (m []string) {
	switch l.State.Previous {
	case Title, HTTPHeader, MimeFilter, AllowVars, EncryptVar, EBLSecret, EbrarySite:
		// URL is allowed after these directives.
	default:
		m = append(m, fmt.Sprintf("URL directive is out of order, previous directive: \"%v\" (L1002)", l.State.Previous))
	}
	if l.State.Title == "" {
		m = append(m, "URL directive is before Title directive (L1010)")
	}
	if l.State.URL != "" {
		m = append(m, "Duplicate URL directive in stanza (L2003)")
	}

	urlFromLine := FindURLFromLine(line)
	if urlFromLine == "" {
		m = append(m, "URL directive is not in the right format. (L3009)")
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
	// According to the EZproxy docs, 'Starting point URLs and config.txt',
	// URL, Host, and HostJavaScript directives are checked for starting point URLs.
	// URL origins should be checked against or added to PreviousOrigins.
	// However, so many stanzas duplicate the URL in an HJ or H line that
	// enabling the check below will add a lot of noise to the output.
	// Possible to add behind a 'pedantic' flag later.
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	time.Sleep(300 * time.Millisecond)
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", "", err
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				buf := make([]byte, 1048576)
				scanner := bufio.NewScanner(strings.NewReader(n.FirstChild.Data))
				scanner.Buffer(buf, 5242880)
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
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return source, oclcTitle, nil
}

// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/html"
)

// A version flag, which should be overwritten when building using ldflags.
var version = "devel"

type State struct {
	AnonymousURLNeedsClosing  bool
	CookieOptionNeedsClosing  bool
	InMultiline               bool
	LastLineEmpty             bool
	OCLCTitle                 string
	Previous                  Directive
	PreviousMultilineSegments string
	Source                    string
	Title                     string
	URL                       string
}

type Linter struct {
	Annotate   bool
	Verbose    bool
	Whitespace bool
	State      State
}

func main() {
	annotate := flag.Bool("annotate", false, "Print all lines, not just lines that create warnings")
	verbose := flag.Bool("verbose", false, "Print internal state before each line is processed")
	whitespace := flag.Bool("whitespace", false, "Report on trailing space or tab characters")
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "ezproxy-config-lint: Lint config files for EZproxy\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Version %v\n", version)
		fmt.Fprintf(flag.CommandLine.Output(), "Compiled with %v\n", runtime.Version())
		flag.PrintDefaults()
	}

	// Process the flags.
	flag.Parse()

	// Set the logger to not include timestamp.
	log.SetFlags(0)

	// Create a Linter struct to hold configuration options.
	linter := &Linter{
		Annotate:   *annotate,
		Verbose:    *verbose,
		Whitespace: *whitespace,
	}

	// Final exit code, after processing all files.
	exitCode := 0

	for _, arg := range flag.Args() {
		fileExitCode, err := linter.processFile(arg)
		if err != nil {
			log.Fatalf("Error processing %v: %v", arg, err)
		}
		if exitCode == 0 && fileExitCode != 0 {
			exitCode = fileExitCode
		}
	}

	os.Exit(exitCode)
}

func (l *Linter) processFile(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 1, err
	}
	defer f.Close()

	// Color output.
	yellow := color.New(color.FgYellow).SprintFunc()

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

	// What exit code should be used.
	exit := 0

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

		if l.Verbose {
			fmt.Printf("%#v\n", l.State)
		}

		warnings := l.ProcessLine(line)
		if len(warnings) > 0 {
			exit = 2
			if l.State.LastLineEmpty {
				for range len(strconv.Itoa(lineNum)) {
					fmt.Printf(" ")
				}
				fmt.Printf("  %v\n", yellow(fmt.Sprintf("!!! %v", strings.Join(warnings, ", "))))
				if more {
					fmt.Printf("%v:\n", lineNum)
				}
			} else {
				fmt.Printf("%v: %v %v\n", lineNum, line, yellow(fmt.Sprintf("!!! %v", strings.Join(warnings, ", "))))
			}
		} else if l.Annotate && more {
			fmt.Printf("%v: %v\n", lineNum, line)
		}
	}

	// If the scanner encountered any errors, report them to the caller.
	if err := scanner.Err(); err != nil {
		return 1, err
	}
	return exit, nil
}

func (l *Linter) ProcessLine(line string) (m []string) {
	// Does the line end in a space or tab character?
	if l.Whitespace && TrailingSpaceOrTabCheck(line) {
		m = append(m, "Line ends in a space or tab character")
	}

	// Trim leading and trailing spaces to ensure the rest of the linting
	// is uniform.
	line = strings.TrimSpace(line)

	// Is the line empty?
	if line == "" {
		if l.State.Title != "" && l.State.URL == "" {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has Title but no URL", l.State.Title))
		}
		if l.State.AnonymousURLNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has AnonymousURL but doesn't have a corresponding \"AnonymousURL -*\" "+
				"line at the end of the stanza", l.State.Title))
		}
		if l.State.CookieOptionNeedsClosing {
			m = append(m, fmt.Sprintf("Stanza \"%v\" has \"Option DomainCookieOnly\" or \"Option CookiePassthrough\" "+
				"but doesn't have a corresponding \"Option Cookie\" line at the end of the stanza", l.State.Title))
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
				m = append(m, fmt.Sprintf("Error processsing Source line: %v", err))
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
			m = append(m, "Option directive not in the form Option OPTIONNAME")
			return m
		}
		label = line
	}

	// Find the Directive which matches this label.
	directive, ok := LabelToDirective[label]
	if !ok {
		m = append(m, "Unknown directive")
		return m
	}

	// Check each directive.
	switch directive {
	case OptionCookie:
		if l.State.CookieOptionNeedsClosing {
			switch l.State.Previous {
			case URL, Host, HostJavaScript, Domain, DomainJavaScript, Replace:
			case AnonymousURL:
				if l.State.AnonymousURLNeedsClosing {
					m = append(m, "Option Cookie directive is out of order")
				}
			default:
				m = append(m, "Option Cookie directive is out of order")
			}
			l.State.CookieOptionNeedsClosing = false
		} else {
			switch l.State.Previous {
			case Undefined, HTTPMethod:
			default:
				m = append(m, "Option Cookie directive is out of order")
			}
		}
	case OptionCookiePassThrough:
		switch l.State.Previous {
		case Undefined, Group, HTTPMethod:
		default:
			m = append(m, "Option CookiePassThrough directive is out of order")
		}
		l.State.CookieOptionNeedsClosing = true
	case OptionDomainCookieOnly:
		switch l.State.Previous {
		case Undefined, Group, HTTPMethod:
		default:
			m = append(m, "Option DomainCookieOnly directive is out of order")
		}
		l.State.CookieOptionNeedsClosing = true
	case AnonymousURL:
		if l.State.AnonymousURLNeedsClosing {
			if strings.TrimPrefix(line, "AnonymousURL ") == "-*" {
				switch l.State.Previous {
				case URL, Host, HostJavaScript, Domain, DomainJavaScript, Replace:
				default:
					m = append(m, "AnonymousURL directive is out of order")
				}
				l.State.AnonymousURLNeedsClosing = false
			} else {
				switch l.State.Previous {
				case AnonymousURL:
				default:
					m = append(m, "AnonymousURL directive is out of order")
				}
			}
		} else {
			switch l.State.Previous {
			case Undefined, Group, HTTPMethod, OptionCookie, OptionCookiePassThrough, OptionDomainCookieOnly, ProxyHostnameEdit:
			default:
				m = append(m, "AnonymousURL directive is out of order")
			}
			l.State.AnonymousURLNeedsClosing = true
		}
	case Title:
		if l.State.Title != "" {
			m = append(m, "Duplicate Title directive")
		}
		l.State.Title = TrimTitlePrefix(line)
		if l.State.OCLCTitle != "" && l.State.Title != l.State.OCLCTitle {
			m = append(m, "Source title doesn't match, you might need to update this stanza.")
		}
		switch l.State.Previous {
		case Undefined, Group, HTTPMethod, OptionCookiePassThrough, OptionDomainCookieOnly, ProxyHostnameEdit, OptionEbraryUnencodedTokens:
		case OptionCookie:
			if !l.State.CookieOptionNeedsClosing {
				m = append(m, "Title directive is out of order")
			}
		case AnonymousURL:
			if !l.State.AnonymousURLNeedsClosing {
				m = append(m, "Title directive is out of order")
			}
		default:
			m = append(m, "Title directive is out of order")
		}
	case URL:
		if l.State.Title == "" {
			m = append(m, "URL directive is before Title directive")
		}
		if l.State.URL != "" {
			m = append(m, "Duplicate URL directive")
		}
		l.State.URL = TrimURLPrefix(line)
		switch l.State.Previous {
		case Title, HTTPHeader, MimeFilter, EbrarySite:
		default:
			m = append(m, "URL directive is out of order")
		}
		parsedURL, err := url.Parse(l.State.URL)
		if err != nil {
			m = append(m, "Unable to prase URL")
		}
		if err == nil {
			if parsedURL.Scheme != "https" {
				m = append(m, "URL is not using HTTPS scheme")
			}
		}
	}
	l.State.Previous = directive
	return m
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
						oclcTitle = TrimTitlePrefix(line)
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

func TrailingSpaceOrTabCheck(line string) bool {
	if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
		return true
	}
	return false
}

func TrimTitlePrefix(line string) string {
	if strings.HasPrefix(line, "Title ") {
		return strings.TrimPrefix(line, "Title ")
	}
	if strings.HasPrefix(line, "T ") {
		return strings.TrimPrefix(line, "T ")
	}
	return line
}

func TrimURLPrefix(line string) string {
	if strings.HasPrefix(line, "URL ") {
		return strings.TrimPrefix(line, "URL ")
	}
	if strings.HasPrefix(line, "U ") {
		return strings.TrimPrefix(line, "U ")
	}
	return line
}

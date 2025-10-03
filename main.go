// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/cu-library/ezproxy-config-lint/internal/linter"
)

type ExitCode int

const (
	Failure = iota + 1 // Linting was successful, and issues were found.
	Error              // Linting was unsuccessful.
)

// A version flag, which should be overwritten when building using ldflags.
var version = "devel"

func main() {
	annotate := flag.Bool("annotate", false, "Print all lines, not just lines that create warnings.")
	verbose := flag.Bool("verbose", false, "Print internal state before each line is processed.")
	additionalPHEChecks := flag.Bool("phe", false, "Perform additional checks on ProxyHostnameEdit directives.")
	directiveCase := flag.Bool("case", false, "Report on directives having the wrong case.")
	https := flag.Bool("https", false, "Report on URL directives which do not use the HTTPS scheme.")
	source := flag.Bool("source", true, "Use source comments to check against OCLC stanzas.")
	pedantic := flag.Bool("pedantic", false, "Enable pedantic checks.")
	whitespace := flag.Bool("whitespace", false, "Report on trailing space or tab characters.")
	followIncludeFile := flag.Bool("follow-includefile", true, "Also process files referenced by IncludeFile directives.")
	includeFileDirectory := flag.String("includefile-directory", "", "The directory from which the IncludeFile paths will be resolved. "+
		"By default, IncludeFile paths are resolved from the parent directory of each of the file arguments, unless they are absolute paths.")
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "ezproxy-config-lint: Lint config files for EZproxy\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  Version %v\n", version)
		fmt.Fprintf(flag.CommandLine.Output(), "  Compiled with %v\n", runtime.Version())
		fmt.Fprint(flag.CommandLine.Output(), "Usage:\n  ezproxy-config-lint [options] <file>...\n")
		fmt.Fprint(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}

	// Process the flags.
	flag.Parse()

	// Set the logger to not include timestamp.
	log.SetFlags(0)

	// Create a Linter struct to hold configuration options.
	linter := &linter.Linter{
		Annotate:             *annotate,
		Verbose:              *verbose,
		AdditionalPHEChecks:  *additionalPHEChecks,
		DirectiveCase:        *directiveCase,
		HTTPS:                *https,
		Source:               *source,
		Pedantic:             *pedantic,
		Whitespace:           *whitespace,
		FollowIncludeFile:    *followIncludeFile,
		IncludeFileDirectory: *includeFileDirectory,
		Output:               os.Stdout,
	}

	warningCount := 0

	for _, arg := range flag.Args() {
		fileWarningCount, err := linter.ProcessFile(arg)
		if err != nil {
			log.Printf("Error processing %v: %v", arg, err)
			os.Exit(Error)
		}
		warningCount += fileWarningCount
		// ProcessFile() recursively processes files referenced
		// by IncludeFile directives.
		// If includeFileDirectory is not set by a CLI option,
		// ProcessFile() will set the linter's IncludeFileDirectory
		// to the parent directory of the first file is processes.
		// That is done because IncludeFile directives are processed
		// as though they were in the file that was processed first.
		// There might be multiple files passed as CLI arguments,
		// which might not be in the same parent directory.
		// The IncludeFileDirectory is reset here so that it does not
		// potentially remain set to the parent directory of the first
		// filePath in the argument list.
		linter.IncludeFileDirectory = *includeFileDirectory
	}

	if warningCount > 0 {
		if warningCount == 1 {
			fmt.Printf("\n%v issue found.\n", warningCount)
		} else {
			fmt.Printf("\n%v issues found.\n", warningCount)
		}
		os.Exit(Failure)
	}
}

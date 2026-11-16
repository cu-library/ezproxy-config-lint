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

	"github.com/cu-library/ezproxy-config-lint/v5/internal/linter"
)

type ExitCode int

const (
	Failure = iota + 1 // Linting was successful, and issues were found.
	Error              // Linting was unsuccessful.
)

// A version flag, which should be overwritten when building using ldflags.
var version = "devel"

func main() {
	config := linter.DefaultConfig()
	flag.BoolVar(&config.Annotate, "annotate", config.Annotate,
		"Print all lines, not just lines that create warnings.")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose,
		"Print internal state before each line is processed.")
	flag.BoolVar(&config.AdditionalPHEChecks, "phe", config.AdditionalPHEChecks,
		"Perform additional checks on ProxyHostnameEdit directives.")
	flag.BoolVar(&config.DirectiveCase, "case", config.DirectiveCase,
		"Report on directives having the wrong case.")
	flag.BoolVar(&config.HTTPS, "https", config.HTTPS,
		"Report on URL directives which do not use the HTTPS scheme.")
	flag.BoolVar(&config.Origins, "origins", config.Origins,
		"Report on duplicate origins in H or HJ directives within a stanza.")
	flag.BoolVar(&config.Source, "source", config.Source,
		"Use source comments to check against OCLC stanzas.")
	flag.BoolVar(&config.Whitespace, "whitespace", config.Whitespace,
		"Report on trailing space or tab characters.")
	flag.BoolVar(&config.FollowIncludeFile, "follow-includefile", config.FollowIncludeFile,
		"Also process files referenced by IncludeFile directives.")
	flag.StringVar(&config.IncludeFileDirectory, "includefile-directory", config.IncludeFileDirectory,
		"The directory from which the IncludeFile paths will be resolved. "+
			"By default, IncludeFile paths are resolved from the parent directory "+
			"of each of the file arguments, unless they are absolute paths.")
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

	linter := linter.NewLinter(config)
	warningCount := 0
	for _, arg := range flag.Args() {
		fileWarningCount, err := linter.ProcessFile(arg)
		if err != nil {
			log.Printf("Error processing %v: %v", arg, err)
			os.Exit(Error)
		}
		warningCount += fileWarningCount
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

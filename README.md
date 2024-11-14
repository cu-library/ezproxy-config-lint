# ezproxy-config-lint

## Description

`ezproxy-config-lint` is a linter for EZproxy config files. It checks config files for common issues like:

- Misspelled directives.
- Ensuring that stanzas have an `Option Cookie` directive if an `Option DomainCookieOnly` or `Option CookiePassthrough` directive is used.
- Ensuring that stanzas have an `AnonymousURL -*` directive after `AnonymousURL` directives are used.
- Ensuring that stanzas have one, and only one, `URL` and `Title` directive.

The `-annotate` flag makes the tool print the whole file, not just lines which raise warnings. 
The tool uses non-zero exit codes to indicate problems: `1` means an error occured during linting, `2` means at least one warning was printed.

## Checking for updates with 'Source'

The linter has a built-in way to check the OCLC website for updates to some database stanzas. If a comment is seen which matches the pattern "# Source - https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/...", the tool will check the stanza at the provided URL and pull out the `Title` directive. The tool will report if the stanza title in the config file does not match the stanza title from the OCLC website.

For example, for the resource Docuseek2, if the staza begins with a Source comment:

```
# Source - https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/Database_stanzas_D/Docuseek2
Title Docuseek2 (updated 20180101)
...
```
the linter will visit https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/Database_stanzas_D/Docuseek2, and pull out the Title directive, which might be:

```
Title Docuseek2 (updated 20191015)
...
```

Because the title directives do not match, the tool will report that you might want to update the stanza from the source.

## Status

This software is still an early prototype, with lots of false positives, false negatives, and missing features. Please feel free to submit [issues](https://github.com/cu-library/ezproxy-config-lint/issues)!

## Getting Started

1. Download the latest release from https://github.com/cu-library/ezproxy-config-lint/releases. The tool is compiled for Windows, Linux, and Darwin (macOS).
    - If you're on Windows, you probably want this release: https://github.com/cu-library/ezproxy-config-lint/releases/latest/download/ezproxy-config-lint_Windows_x86_64.zip 
    - If you're on Linux, you probably want this release: https://github.com/cu-library/ezproxy-config-lint/releases/latest/download/ezproxy-config-lint_Linux_x86_64.tar.gz
    - If you're on macOS, you probably want this release: https://github.com/cu-library/ezproxy-config-lint/releases/latest/download/ezproxy-config-lint_Darwin_x86_64.tar.gz
2. Unzip or untar the archive.
    - On Linux: Navigate to your download directory, then run `tar xzvf ezproxy-config-lint_Linux_x86_64.tar.gz`
    - On Windows: Right-click on the .zip file in your download directory, and select "Extract All..."
3. On the command line, nagivate to where the `ezproxy-config-lint` tool was extracted.
    - On Windows: If you've installed [Windows Terminal](https://aka.ms/terminal), you can right-click on the new folder and select "Open in Terminal".
4. Run the tool, passing the config file you want to lint as a argument.

## Example

```
$ ls
config.txt
$ cat config.txt
Option DomainCookieOnly
Title EB Medicine
HJ http://www.ebmedicine.net
URL https://www.ebmedicine.net
DJ ebmedicine.net
NeverProxy cdnjs.cloudflare.com
$ wget --quiet https://github.com/cu-library/ezproxy-config-lint/releases/latest/download/ezproxy-config-lint_Linux_x86_64.tar.gz
$ tar xzvf ezproxy-config-lint_Linux_x86_64.tar.gz
$ ./ezproxy-config-lint -help
ezproxy-config-lint: Lint config files for EZproxy
  -annotate
        Print all lines, not just lines that create warnings
  -verbose
        Print internal state before each line is processed
  -whitespace
        Report on trailing space or tab characters
$ ./ezproxy-config-lint config.txt
4: URL https://www.ebmedicine.net !!! URL directive is out of order
   !!! Stanza "EB Medicine (updated 20190614)" has "Option DomainCookieOnly" or "Option CookiePassthrough" but doesn't have a corresponding "Option Cookie" line at the end of the stanza
$ ./ezproxy-config-lint -annotate config.txt
1: Option DomainCookieOnly
2: Title EB Medicine (updated 20190614)
3: HJ http://www.ebmedicine.net
4: URL https://www.ebmedicine.net !!! URL directive is out of order
5: DJ ebmedicine.net
6: NeverProxy cdnjs.cloudflare.com
   !!! Stanza "EB Medicine (updated 20190614)" has "Option DomainCookieOnly" or "Option CookiePassthrough" but doesn't have a corresponding "Option Cookie" line at the end of the stanza
```
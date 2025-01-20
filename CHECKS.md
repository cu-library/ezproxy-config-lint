# Checks

Explanations of all checks in `ezproxy-config-lint`.

| Check           | Short description
|-----------------|------------------
| **L1**          | **Ordering Issues**
| [L1001](#l1001) | `Title` directive is out of order
| [L1002](#l1002) | `URL` directive is out of order
| [L1003](#l1003) | `AnonymousURL -*` directive is out of order
| [L1004](#l1004) | `AnonymousURL` directive is out of order
| [L1005](#l1005) | `Option Cookie` directive is out of order
| [L1006](#l1006) | `Option CookiePassThrough` directive is out of order
| [L1007](#l1007) | `Option DomainCookieOnly` directive is out of order
| [L1008](#l1008) | `ProxyHostnameEdit` directive is out of order
| [L1009](#l1009) | `ProxyHostnameEdit` domains should be placed in deepest-to-shallowest order
| [L1010](#l1010) | `URL` directive is before `Title` directive
| [L1011](#l1011) | `Option Cookie` directive should not preceed closing AnonymousURL
| **L2**          | **Duplication Issues**
| [L2001](#l2001) | Duplicate `Title` directive in stanza
| [L2002](#l2002) | Origin already seen
| [L2003](#l2003) | Duplicate `URL` directive in stanza
| [L2004](#l2004) | `Title` value already seen
| **L3**          | **Malformation Issues**
| [L3001](#l3001) | `ProxyHostnameEdit` directive must have two values
| [L3002](#l3002) | Find part of `ProxyHostnameEdit` directive should end with a `$`
| [L3003](#l3003) | Replace part of `ProxyHostnameEdit` directive is malformed
| [L3004](#l3004) | `Domain` and `DomainJavaScript` directives should only specify domains
| [L3005](#l3005) | Unable to parse `URL`
| [L3006](#l3006) | `URL` does not start with `http` or `https`
| [L3007](#l3007) | `URL` is not using HTTPS scheme
| [L3008](#l3008) | `Option` directive not in the form `Option OPTIONNAME`
| [L3009](#l3009) | `URL` directive is not in the right format
| **L4**          | **Missing Directive Issues**
| [L4001](#l4001) | Missing `AnonymousURL -*` clearing at end of stanza
| [L4002](#l4002) | Missing `Option Cookie` at end of stanza
| [L4003](#l4003) | Stanza has `Title` but no `URL`
| [L4004](#l4004) | `Find` directive must be immediately proceeded with a `Replace` directive
| **L5**          | **Styling Issues**
| [L5001](#l5001) | Directive uses the wrong case.
| [L5002](#l5002) | Line ends in a space or tab character.
| **L9**          | **Other Issues**
| [L9001](#l9001) | Unknown directive
| [L9002](#l9002) | Source title doesn't match
| [L9003](#l9003) | Error processing Source line

## L1 - Ordering Issues

### L1001

#### `Title` directive is out of order

The `Title` directive are only allowed to follow these directives:

* `AddUserHeader`
* `Group`
* `HTTPMethod`
* `Option CookiePassThrough`
* `Option DomainCookieOnly`
* `Option X-Forwarded-For`
* `Cookie`
* `Option ebraryUnencodedTokens`
* `ProxyHostnameEdit`
* `Referer`

Additionally, a `Title` will be considered out of order when
[L4001](#l4001) or [L4002](#l4002) are detected.

---------

### L1002

#### `URL` directive is out of order

The `URL` directive are only allowed to follow these directives:

* `AllowVars`
* `EBLSecret`
* `EbrarySite`
* `EncryptVar`
* `HTTPHeader`
* `MimeFilter`
* `Title`

---------

### L1003

#### `AnonymousURL -*` directive is out of order

The `AnonymousURL -*` directive is only allowed to follow these directives:

* `DomainJavaScript`
* `Domain`
* `HostJavaScript`
* `Host`
* `Replace`
* `URL`
* `NeverProxy`
* `Option Cookie`
* `Option NoX-Forwarded-For`

---------

### L1004

#### `AnonymousURL` directive is out of order

Except for the ending `AnonymousURL -*` usage (see [L1003](#l1003]),
the `AnonymousURL` directive is only allowed to follow these directives:

* `AnonymousURL`
* `Group`
* `HTTPMethod`
* `Option CookiePassThrough`
* `Option Cookie`
* `Option DomainCookieOnly`
* `Option X-Forwarded-For`
* `ProxyHostnameEdit`

---------

### L1005

#### `Option Cookie` directive is out of order

The `Option Cookie` directive is only allowed to follow these directives at the beginning of the stanza:

* `Group`
* `HTTPMethod`
* `Option X-Forwarded-For`
* `AnonymousURL`

The directive should also be used at the end of the stanza to ensure other cookie handling options do not impact other stanzas.

---------

### L1006

#### `Option CookiePassThrough` directive is out of order

The `Option CookiePassThrough` directive is only allowed to follow these directives:

* `Group`
* `HTTPMethod`
* `Option X-Forwarded-For`
* `AnonymousURL`

---------

### L1007

#### `Option DomainCookieOnly` directive is out of order

The `Option DomainCookieOnly` directive is only allowed to follow these
directives:

* `Group`
* `HTTPMethod`
* `Option X-Forwarded-For`
* `AnonymousURL`

---------

### L1008

#### `ProxyHostnameEdit` directive is out of order

The `ProxyHostnameEdit` directive is only allowed to follow these directives:

* `Group`
* `HTTPMethod`
* `Option Cookie`
* `Option CookiePassThrough`
* `Option DomainCookieOnly`
* `Option X-Forwarded-For`
* `AnonymousURL`
* `ProxyHostnameEdit`

---------

### L1009

#### `ProxyHostnameEdit` domains should be placed in deepest-to-shallowest order

This check is enabled with the `-phe=true` option.

It is best explained with an example.
Lets say you have these two lines in your stanza:

```
ProxyHostnameEdit heinonline.org$ heinonline-org
ProxyHostnameEdit home.heinonline.org$ home-heinonline-org
```

Assume EZproxy was processing the domain name `home.heinonline.org`.
It would start with the first line, and the resulting find and replace action would generate the domain name `home.heinonline-org`.
Because that domain name has one period and one hyphen, it would not match the second `ProxyHostnameEdit` line.
You always want to start with more specific, deeper subdomains.

We use "deeper" instead of "longer" here, because areallylongdomainhere.com is only two components deep,
but a.short.domain.ca is four components deep.

---------

### L1010

#### `URL` directive is before `Title` directive

The `URL` directive should always come after the `Title` is a given stanza.

---------

### L1011

#### `Option Cookie` directive

The `Option Cookie` directive, when used at the end of a stanza to prevent other cookie handling options from impacting other stanzas, is only allowed to follow the following directives:

* `URL`
* `Host`
* `HostJavaScript`
* `Domain`
* `DomainJavaScript`
* `Replace`
* `AnonymousURL`
* `Option NoX-Forwarded-For`


## L2 - Duplication Issues

### L2001

#### Duplicate `Title` directive in stanza

More than one `Title` directive present in a stanza.

---------

### L2002

#### Origin already seen

EZproxy [only reads origins and does not read paths](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Groups).
The origin is the combination of the scheme (http or https), the host (google.com, help.oclc.org), and the port.
EZproxy does not care about paths (/astronomy, /login).
The linter will report if you've already used an origin, so that you can ensure that limiting access via Groups works as you expect.

---------

### L2003

#### Duplicate `URL` directive in stanza

More than one `URL` directive present in a stanza.

---------

### L2004

#### `Title` value already seen

The linter tracks stanza `Title` values and reports when a value has been seen more than
once.


## L3 - Malformation Issues

### L3001

#### `ProxyHostnameEdit` directive must have two values

The `ProxyHostnameEdit` directive requires two qualifiers.
See [OCLC Documentation](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/ProxyHostnameEdit) for more details.

---------

### L3002

#### Find part of `ProxyHostnameEdit` directive should end with a `$`

This check is enabled with the `-phe=true` option. The first qualifier to `ProxyHostnameEdit`, *find*, is a regular expression and should end in a dollar sign (`$`) to match the end of the string.

---------

### L3003

#### Replace part of `ProxyHostnameEdit` directive is malformed

This check is enabled with the `-phe=true` option. To make redirects from HTTP to HTTPS links more reliable, this check ensures that the *find* and *replace* qualifiers correspond with each other: *replace* should just be *find*, but with the trailing `$` removed and the periods replaced with hypens.

---------

### L3004

#### `Domain` and `DomainJavaScript` directives should only specify domains

The `Domain` and `DomainJavaScript` directives should only specify domains.
No URLs or path components should be used.

---------

### L3005

#### Unable to parse `URL`

The URL part of this directive could not be parsed, check to see if it is malformed.

---------

### L3006

#### `URL` does not start with `http` or `https`

The scheme/protocol of the URL should be specified, and it should either be `http` or `https`.

---------

### L3007

#### `URL` is not using HTTPS scheme

This check is enabled with the `-https=true` option. The URL should use the `https` scheme/protocol.

---------

### L3008

#### `Option` directive not in the form `Option OPTIONNAME`

The `Option` directive needs a second part. You can see the list of options in the [OCLC documentation](https://help.oclc.org/Library_Management/EZproxy/Configure_resources).

---------

### L3009

#### `URL` directive is not in the right format

The `URL` directive must be in the [URL (version 1)](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_1), [URL (version 2)](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_2), or [URL (version 3)](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/URL_version_3) format. Ensure line is not malformed.


## L4 - Missing Directive Issues

### L4001

#### Missing `AnonymousURL -*` at end of stanza

Stanzas which use a AnonymousURL directive should include a `AnonymousURL -*` line at the end of the stanza so that other stanzas in the config file are not impacted.

---------

### L4002

#### Missing `Option Cookie` at end of stanza

Stanzas which use a Option DomainCookieOnly or Option CookiePassthrough directive should include a `Option Cookie` line at the end of the stanza so that other stanzas in the config file are not impacted.

---------

### L4003

#### Stanza has `Title` but no `URL`

Stanzas should have both a Title and a URL directive.

---------

### L4004

#### `Find` directive must be immediately proceeded with a `Replace` directive

From the [OCLC documentation](https://help.oclc.org/Library_Management/EZproxy/Configure_resources/Find_Replace):

"These directives always appear paired together, with Find appearing directly before its corresponding Replace."


## L5 - Styling Issues

### L5001

#### Directive uses the wrong case.

This check is enabled with the `-case=true` option. The directive was found, but it does not use the normal case. For example, TITLE instead of Title, or HTTPheader instead of HTTPHeader.

### L5002

#### Line ends in a space or tab character.

This check is enabled with the `-whitespace=true` option. Trailing whitespace characters space or tab were found on this line.

---------


## L9 - Other Issues

### L9001

#### Unknown directive

An unknown directive was encountered. This can be caused by a typo in the directive. You can check the list of possible directives in the [OCLC documentation](https://help.oclc.org/Library_Management/EZproxy/Configure_resources).

---------

### L9002

#### Source title doesn't match

The linter has a built-in way to check the OCLC website for updates to some database stanzas.
If a comment is seen which matches the pattern `# Source - https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/...`,
the tool will check the stanza at the provided URL and pull out the `Title` directive.
The tool will report if the stanza title in the config file does not match the stanza title from the OCLC website.

For example, for the resource Docuseek2, if the stanza begins with a `Source` comment:

```
# Source - https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/Database_stanzas_D/Docuseek2
Title Docuseek2 (updated 20180101)
...
```
the linter will visit https://help.oclc.org/Library_Management/EZproxy/EZproxy_database_stanzas/Database_stanzas_D/Docuseek2,
and pull out the `Title` directive, which might be:

```
Title Docuseek2 (updated 20191015)
...
```

Because the `Title` directives do not match, the tool will report that you might want to update the stanza from the source.

### L9003

#### Error processing Source line

There was some problem processing the Source line. The URL might be malformed, or there was an HTTP request issue.
// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package linter

// optionPairs maps each opening Option directive to the directive that closes it.
var optionPairs = map[Directive]Directive{ //nolint:gochecknoglobals
	OptionCookiePassThrough:    OptionCookie,
	OptionDomainCookieOnly:     OptionCookie,
	OptionHideEZproxy:          OptionNoHideEZproxy,
	OptionMetaEZproxyRewriting: OptionNoMetaEZproxyRewriting,
	OptionNoCookie:             OptionCookie,
	OptionNoHttpsHyphens:       OptionHttpsHyphens,
	OptionProxyFTP:             OptionNoProxyFTP,
	OptionUTF16:                OptionNoUTF16,
	OptionXForwardedFor:        OptionNoXForwardedFor,
}

// titleAllowedPrevious is the set of directives that may appear immediately before Title.
// It is built once. Openers are added from optionPairs in init.
var titleAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:               {},
	AnonymousURL:                {},
	Cookie:                      {},
	DbVar:                       {},
	DbVar0:                      {},
	DbVar1:                      {},
	DbVar2:                      {},
	DbVar3:                      {},
	DbVar4:                      {},
	DbVar5:                      {},
	DbVar6:                      {},
	DbVar7:                      {},
	DbVar8:                      {},
	DbVar9:                      {},
	Group:                       {},
	HTTPMethod:                  {},
	OptionCookie:                {},
	OptionEbraryUnencodedTokens: {},
	ProxyHostnameEdit:           {},
	Referer:                     {},
	Undefined:                   {},
}

// descriptionAllowedPrevious is the set of directives that may appear immediately before Description.
var descriptionAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	Description: {},
	Title:       {},
}

// proxyHostnameEditAllowedPrevious is the set of directives that may appear immediately before ProxyHostnameEdit.
// Openers are added from optionPairs in init.
var proxyHostnameEditAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:     {},
	AnonymousURL:      {},
	Cookie:            {},
	DbVar:             {},
	DbVar0:            {},
	DbVar1:            {},
	DbVar2:            {},
	DbVar3:            {},
	DbVar4:            {},
	DbVar5:            {},
	DbVar6:            {},
	DbVar7:            {},
	DbVar8:            {},
	DbVar9:            {},
	Group:             {},
	HTTPMethod:        {},
	OptionCookie:      {},
	ProxyHostnameEdit: {},
	Undefined:         {},
}

// urlAllowedPrevious is the set of directives that may appear immediately before URL.
var urlAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AllowVars:   {},
	ByteServe:   {},
	Description: {},
	EBLSecret:   {},
	EbrarySite:  {},
	EncryptVar:  {},
	HTTPHeader:  {},
	MimeFilter:  {},
	Title:       {},
}

// optionOpenerAllowedPrevious is the set of directives that may appear immediately before an opening Option.
// Openers are added from optionPairs in init.
var optionOpenerAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:     {},
	AnonymousURL:      {},
	DbVar:             {},
	DbVar0:            {},
	DbVar1:            {},
	DbVar2:            {},
	DbVar3:            {},
	DbVar4:            {},
	DbVar5:            {},
	DbVar6:            {},
	DbVar7:            {},
	DbVar8:            {},
	DbVar9:            {},
	Group:             {},
	HTTPMethod:        {},
	OptionCookie:      {},
	ProxyHostnameEdit: {},
	Undefined:         {},
}

// addUserHeaderAllowedPrevious is the set of directives that may appear immediately before AddUserHeader with qualifiers.
// Openers are added from optionPairs in init.
var addUserHeaderAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:     {},
	AnonymousURL:      {},
	Cookie:            {},
	DbVar:             {},
	DbVar0:            {},
	DbVar1:            {},
	DbVar2:            {},
	DbVar3:            {},
	DbVar4:            {},
	DbVar5:            {},
	DbVar6:            {},
	DbVar7:            {},
	DbVar8:            {},
	DbVar9:            {},
	Group:             {},
	HTTPMethod:        {},
	OptionCookie:      {},
	ProxyHostnameEdit: {},
	Undefined:         {},
}

// addUserHeaderBareAllowedPrevious is the set of directives that may appear immediately before AddUserHeader with no qualifiers.
// Closers are added from optionPairs in init.
var addUserHeaderBareAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AnonymousURL:     {},
	Domain:           {},
	DomainJavaScript: {},
	Host:             {},
	HostJavaScript:   {},
	NeverProxy:       {},
	Replace:          {},
	URL:              {},
}

// anonymousURLAllowedPrevious is the set of directives that may appear immediately before AnonymousURL.
// Openers are added from optionPairs in init.
var anonymousURLAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:     {},
	AnonymousURL:      {},
	Cookie:            {},
	DbVar:             {},
	DbVar0:            {},
	DbVar1:            {},
	DbVar2:            {},
	DbVar3:            {},
	DbVar4:            {},
	DbVar5:            {},
	DbVar6:            {},
	DbVar7:            {},
	DbVar8:            {},
	DbVar9:            {},
	Domain:            {},
	DomainJavaScript:  {},
	Group:             {},
	HTTPMethod:        {},
	Host:              {},
	HostJavaScript:    {},
	NeverProxy:        {},
	ProxyHostnameEdit: {},
	Replace:           {},
	URL:               {},
	Undefined:         {},
}

// anonymousURLStarAllowedPrevious is the set of directives that may appear immediately before "AnonymousURL -*".
// Closers are added from optionPairs in init.
var anonymousURLStarAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:    {},
	AnonymousURL:     {},
	Domain:           {},
	DomainJavaScript: {},
	Host:             {},
	HostJavaScript:   {},
	NeverProxy:       {},
	Replace:          {},
	URL:              {},
}

// optionCloserAllowedPrevious is the set of directives that may appear immediately before a closing Option.
// Closers are added from optionPairs in init.
var optionCloserAllowedPrevious = map[Directive]struct{}{ //nolint:gochecknoglobals
	AddUserHeader:    {},
	AnonymousURL:     {},
	DbVar:            {},
	Domain:           {},
	DomainJavaScript: {},
	Host:             {},
	HostJavaScript:   {},
	NeverProxy:       {},
	Replace:          {},
	URL:              {},
}

// openerDirectives is the set of Option directives that open a pair.
// It is filled from optionPairs in init.
var openerDirectives = map[Directive]struct{}{} //nolint:gochecknoglobals

// closerDirectives is the set of Option directives that close a pair.
// It is filled from optionPairs in init.
var closerDirectives = map[Directive]struct{}{} //nolint:gochecknoglobals

func init() {
	for _, opener := range OpenerOptions() {
		openerDirectives[opener] = struct{}{}

		addUserHeaderAllowedPrevious[opener] = struct{}{}
		anonymousURLAllowedPrevious[opener] = struct{}{}
		optionOpenerAllowedPrevious[opener] = struct{}{}
		proxyHostnameEditAllowedPrevious[opener] = struct{}{}
		titleAllowedPrevious[opener] = struct{}{}
	}
	for _, closer := range CloserOptions() {
		closerDirectives[closer] = struct{}{}

		addUserHeaderBareAllowedPrevious[closer] = struct{}{}
		anonymousURLStarAllowedPrevious[closer] = struct{}{}
		optionCloserAllowedPrevious[closer] = struct{}{}
	}
}

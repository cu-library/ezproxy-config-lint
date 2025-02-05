// Copyright Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package linter

import (
	"strings"
)

//go:generate stringer -type Directive --linecomment
type Directive int

//nolint:godot
const (
	Undefined Directive = iota
	AddUserHeader
	AllowIP
	AllowVars
	AnonymousURL
	Audit
	AuditPurge
	AutoLoginIP
	AutoLoginIPBanner
	BinaryTimeout
	Books24x7Site
	ByteServe
	CASServiceURL
	ChargeSetLatency
	Charset
	ClientTimeout
	ConnectWindow
	Cookie
	CookieFilter
	DbVar
	DenyIfRequestHeader
	Description
	DNS
	Domain
	DomainJavaScript
	EBLSecret
	EbrarySite
	EncryptVar
	ExcludeIP
	ExcludeIPBanner
	ExtraLoginCookie
	Find
	FirstPort
	FormSelect
	FormSubmit
	FormVariable
	Gartner
	Group
	HAName
	HAPeer
	Host
	HostJavaScript
	HTTPHeader
	HTTPMethod
	Identifier
	IncludeFile
	IncludeIP
	Interface
	IntruderIPAttempts
	IntruderLog
	IntruderUserAttempts
	IntrusionAPI
	LBPeer
	Location
	LogFile
	LogFilter
	LogFormat
	LoginCookieDomain
	LoginCookieName
	LoginMenu
	LoginPort
	LoginPortSSL
	LogSPU
	MaxConcurrentTransfers
	MaxLifetime
	MaxSessions
	MaxVirtualHosts
	MessagesFile
	MetaFind
	MimeFilter
	Name
	NeverProxy
	OptionAcceptXForwardedFor                                       // Option AcceptX-Forwarded-For
	OptionAllowSendGZip                                             // Option AllowSendGZip
	OptionAllowWebSubdirectories                                    // Option AllowWebSubdirectories
	OptionAnyDNSHostname                                            // Option AnyDNSHostname
	OptionBlockCountryChange                                        // Option BlockCountryChange
	OptionCookie                                                    // Option Cookie
	OptionCookiePassThrough                                         // Option CookiePassThrough
	OptionCSRFToken                                                 // Option CSRFToken
	OptionDisableSSL40bit                                           // Option DisableSSL40bit
	OptionDisableSSL56bit                                           // Option DisableSSL56bit
	OptionDisableSSLv2                                              // Option DisableSSLv2
	OptionDomainCookieOnly                                          // Option DomainCookieOnly
	OptionEbraryUnencodedTokens                                     // Option ebraryUnencodedTokens
	OptionExcludeIPMenu                                             // Option ExcludeIPMenu
	OptionForceHTTPSAdmin                                           // Option ForceHTTPSAdmin
	OptionForceHTTPSLogin                                           // Option ForceHTTPSLogin
	OptionForceWildcardCertificate                                  // Option ForceWildcardCertificate
	OptionHideEZproxy                                               // Option HideEZproxy
	OptionHttpsHyphens                                              // Option HttpsHyphens
	OptionIChooseToUseDomainLinesThatThreatenTheSecurityOfMyNetwork // Option I choose to use Domain lines that threaten the security of my network
	OptionIgnoreWildcardCertificate                                 // Option IgnoreWildcardCertificate
	OptionIPv6                                                      // Option IPv6
	OptionLoginReplaceGroups                                        // Option LoginReplaceGroups
	OptionLogReferer                                                // Option LogReferer
	OptionLogSAML                                                   // Option LogSAML
	OptionLogSession                                                // Option LogSession
	OptionLogSPUEdit                                                // Option LogSPUEdit
	OptionLogUser                                                   // Option LogUser
	OptionMenuByGroups                                              // Option MenuByGroups
	OptionMetaEZproxyRewriting                                      // Option MetaEZproxyRewriting
	OptionNoCookie                                                  // Option NoCookie
	OptionNoHideEZproxy                                             // Option NoHideEZproxy
	OptionNoHttpsHyphens                                            // Option NoHttpsHyphens
	OptionNoMetaEZproxyRewriting                                    // Option NoMetaEZproxyRewriting
	OptionNoProxyFTP                                                // Option NoProxyFTP
	OptionNoUTF16                                                   // Option NoUTF16
	OptionNoXForwardedFor                                           // Option NoX-Forwarded-For
	OptionProxyByHostname                                           // Option ProxyByHostname
	OptionProxyFTP                                                  // Option ProxyFTP
	OptionRecordPeaks                                               // Option RecordPeaks
	OptionRedirectUnknown                                           // Option RedirectUnknown
	OptionReferInHostname                                           // Option ReferInHostname
	OptionRelaxedRADIUS                                             // Option RelaxedRADIUS
	OptionRequireAuthenticate                                       // Option RequireAuthenticate
	OptionSafariCookiePatch                                         // Option SafariCookiePatch
	OptionStatusUser                                                // Option StatusUser
	OptionTicketIgnoreExcludeIP                                     // Option TicketIgnoreExcludeIP
	OptionUnsafeRedirectUnknown                                     // Option UnsafeRedirectUnknown
	OptionUsernameCaretN                                            // Option UsernameCaretN
	OptionUTF16                                                     // Option UTF16
	OptionXForwardedFor                                             // Option X-Forwarded-For
	OverDriveSite
	PDFRefresh
	PDFRefreshPost
	PDFRefreshPre
	PidFile
	Proxy
	ProxyHostnameEdit
	ProxySSL
	RADIUSRetry
	RedirectSafe
	Referer
	RejectIP
	RemoteIPHeader
	RemoteIPInternalProxy
	RemoteIPTrustedProxy
	RemoteTimeout
	Replace
	RunAs
	ShibbolethDisable
	ShibbolethMetadata
	SkipPort
	SPUEdit
	SPUEditVar
	SQLiteTempDir
	SSLCipherSuite
	SSLHonorCipherOrder
	SSLOpenSSLConfCmd
	SSOUsername
	Title
	TokenKey
	TokenSignatureKey
	UMask
	URL
	URLAppendEncoded
	URLRedirect
	URLRedirectAppend
	URLRedirectAppendEncoded
	UsageLimit
	Validate
	XDebug
)

var LabelToDirective = map[string]Directive{ //nolint:gochecknoglobals
	"A":                               AutoLoginIP,
	"AddUserHeader":                   AddUserHeader,
	"AllowIP":                         AllowIP,
	"AllowVars":                       AllowVars,
	"AnonymousURL":                    AnonymousURL,
	"Audit":                           Audit,
	"AuditPurge":                      AuditPurge,
	"AutoLoginIP":                     AutoLoginIP,
	"AutoLoginIPBanner":               AutoLoginIPBanner,
	"BinaryTimeout":                   BinaryTimeout,
	"Books24x7Site":                   Books24x7Site,
	"ByteServe":                       ByteServe,
	"CASServiceURL":                   CASServiceURL,
	"ChargeSetLatency":                ChargeSetLatency,
	"Charset":                         Charset,
	"ClientTimeout":                   ClientTimeout,
	"ConnectWindow":                   ConnectWindow,
	"Cookie":                          Cookie,
	"CookieFilter":                    CookieFilter,
	"D":                               Domain,
	"DbVar":                           DbVar,
	"DenyIfRequestHeader":             DenyIfRequestHeader,
	"Description":                     Description,
	"DJ":                              DomainJavaScript,
	"DNS":                             DNS,
	"Domain":                          Domain,
	"DomainJavaScript":                DomainJavaScript,
	"E":                               ExcludeIP,
	"EBLSecret":                       EBLSecret,
	"ebrarySite":                      EbrarySite,
	"EncryptVar":                      EncryptVar,
	"ExcludeIP":                       ExcludeIP,
	"ExcludeIPBanner":                 ExcludeIPBanner,
	"ExtraLoginCookie":                ExtraLoginCookie,
	"Find":                            Find,
	"FirstPort":                       FirstPort,
	"FormSelect":                      FormSelect,
	"FormSubmit":                      FormSubmit,
	"FormVariable":                    FormVariable,
	"Gartner":                         Gartner,
	"Group":                           Group,
	"H":                               Host,
	"HAName":                          HAName,
	"HAPeer":                          HAPeer,
	"HJ":                              HostJavaScript,
	"Host":                            Host,
	"HostJavaScript":                  HostJavaScript,
	"HTTPHeader":                      HTTPHeader,
	"HTTPMethod":                      HTTPMethod,
	"I":                               IncludeIP,
	"Identifier":                      Identifier,
	"IncludeFile":                     IncludeFile,
	"IncludeIP":                       IncludeIP,
	"Interface":                       Interface,
	"IntruderIPAttempts":              IntruderIPAttempts,
	"IntruderLog":                     IntruderLog,
	"IntruderUserAttempts":            IntruderUserAttempts,
	"IntrusionAPI":                    IntrusionAPI,
	"LBPeer":                          LBPeer,
	"Location":                        Location,
	"LogFile":                         LogFile,
	"LogFilter":                       LogFilter,
	"LogFormat":                       LogFormat,
	"LoginCookieDomain":               LoginCookieDomain,
	"LoginCookieName":                 LoginCookieName,
	"LoginMenu":                       LoginMenu,
	"LoginPort":                       LoginPort,
	"LoginPortSSL":                    LoginPortSSL,
	"LogSPU":                          LogSPU,
	"MaxConcurrentTransfers":          MaxConcurrentTransfers,
	"MaxLifetime":                     MaxLifetime,
	"MaxSessions":                     MaxSessions,
	"MaxVirtualHosts":                 MaxVirtualHosts,
	"MC":                              MaxConcurrentTransfers,
	"MessagesFile":                    MessagesFile,
	"MetaFind":                        MetaFind,
	"MimeFilter":                      MimeFilter,
	"ML":                              MaxLifetime,
	"MS":                              MaxSessions,
	"MV":                              MaxVirtualHosts,
	"Name":                            Name,
	"NeverProxy":                      NeverProxy,
	"Option AcceptX-Forwarded-For":    OptionAcceptXForwardedFor,
	"Option AllowSendGZip":            OptionAllowSendGZip,
	"Option AllowWebSubdirectories":   OptionAllowWebSubdirectories,
	"Option AnyDNSHostname":           OptionAnyDNSHostname,
	"Option BlockCountryChange":       OptionBlockCountryChange,
	"Option Cookie":                   OptionCookie,
	"Option CookiePassThrough":        OptionCookiePassThrough,
	"Option CSRFToken":                OptionCSRFToken,
	"Option DisableSSL40bit":          OptionDisableSSL40bit,
	"Option DisableSSL56bit":          OptionDisableSSL56bit,
	"Option DisableSSLv2":             OptionDisableSSLv2,
	"Option DomainCookieOnly":         OptionDomainCookieOnly,
	"Option ebraryUnencodedTokens":    OptionEbraryUnencodedTokens,
	"Option ExcludeIPMenu":            OptionExcludeIPMenu,
	"Option ForceHTTPSAdmin":          OptionForceHTTPSAdmin,
	"Option ForceHTTPSLogin":          OptionForceHTTPSLogin,
	"Option ForceWildcardCertificate": OptionForceWildcardCertificate,
	"Option HideEZproxy":              OptionHideEZproxy,
	"Option HttpsHyphens":             OptionHttpsHyphens,
	"Option I choose to use Domain lines that threaten the security of my network": OptionIChooseToUseDomainLinesThatThreatenTheSecurityOfMyNetwork,
	"Option IgnoreWildcardCertificate":                                             OptionIgnoreWildcardCertificate,
	"Option IPv6":                                                                  OptionIPv6,
	"Option LoginReplaceGroups":                                                    OptionLoginReplaceGroups,
	"Option LogReferer":                                                            OptionLogReferer,
	"Option LogSAML":                                                               OptionLogSAML,
	"Option LogSession":                                                            OptionLogSession,
	"Option LogSPUEdit":                                                            OptionLogSPUEdit,
	"Option LogUser":                                                               OptionLogUser,
	"Option MenuByGroups":                                                          OptionMenuByGroups,
	"Option MetaEZproxyRewriting":                                                  OptionMetaEZproxyRewriting,
	"Option NoCookie":                                                              OptionNoCookie,
	"Option NoHideEZproxy":                                                         OptionNoHideEZproxy,
	"Option NoHttpsHyphens":                                                        OptionNoHttpsHyphens,
	"Option NoMetaEZproxyRewriting":                                                OptionNoMetaEZproxyRewriting,
	"Option NoProxyFTP":                                                            OptionNoProxyFTP,
	"Option NoUTF16":                                                               OptionNoUTF16,
	"Option NoX-Forwarded-For":                                                     OptionNoXForwardedFor,
	"Option ProxyByHostname":                                                       OptionProxyByHostname,
	"Option ProxyFTP":                                                              OptionProxyFTP,
	"Option RecordPeaks":                                                           OptionRecordPeaks,
	"Option RedirectUnknown":                                                       OptionRedirectUnknown,
	"Option ReferInHostname":                                                       OptionReferInHostname,
	"Option RelaxedRADIUS":                                                         OptionRelaxedRADIUS,
	"Option RequireAuthenticate":                                                   OptionRequireAuthenticate,
	"Option SafariCookiePatch":                                                     OptionSafariCookiePatch,
	"Option StatusUser":                                                            OptionStatusUser,
	"Option TicketIgnoreExcludeIP":                                                 OptionTicketIgnoreExcludeIP,
	"Option UnsafeRedirectUnknown":                                                 OptionUnsafeRedirectUnknown,
	"Option UsernameCaretN":                                                        OptionUsernameCaretN,
	"Option UTF16":                                                                 OptionUTF16,
	"Option X-Forwarded-For":                                                       OptionXForwardedFor,
	"OverDriveSite":                                                                OverDriveSite,
	"PDFRefresh":                                                                   PDFRefresh,
	"PDFRefreshPost":                                                               PDFRefreshPost,
	"PDFRefreshPre":                                                                PDFRefreshPre,
	"PHE":                                                                          ProxyHostnameEdit,
	"PidFile":                                                                      PidFile,
	"PIDFile":                                                                      PidFile,
	"Proxy":                                                                        Proxy,
	"ProxyHostnameEdit":                                                            ProxyHostnameEdit,
	"ProxySSL":                                                                     ProxySSL,
	"RADIUSRetry":                                                                  RADIUSRetry,
	"RedirectSafe":                                                                 RedirectSafe,
	"Referer":                                                                      Referer,
	"RejectIP":                                                                     RejectIP,
	"RemoteIPHeader":                                                               RemoteIPHeader,
	"RemoteIPInternalProxy":                                                        RemoteIPInternalProxy,
	"RemoteIPTrustedProxy":                                                         RemoteIPTrustedProxy,
	"RemoteTimeout":                                                                RemoteTimeout,
	"Replace":                                                                      Replace,
	"RunAs":                                                                        RunAs,
	"ShibbolethDisable":                                                            ShibbolethDisable,
	"ShibbolethMetadata":                                                           ShibbolethMetadata,
	"SkipPort":                                                                     SkipPort,
	"SPUEdit":                                                                      SPUEdit,
	"SPUEditVar":                                                                   SPUEditVar,
	"SQLiteTempDir":                                                                SQLiteTempDir,
	"SSLCipherSuite":                                                               SSLCipherSuite,
	"SSLHonorCipherOrder":                                                          SSLHonorCipherOrder,
	"SSLOpenSSLConfCmd":                                                            SSLOpenSSLConfCmd,
	"SSOUsername":                                                                  SSOUsername,
	"T":                                                                            Title,
	"Title":                                                                        Title,
	"TokenKey":                                                                     TokenKey,
	"TokenSignatureKey":                                                            TokenSignatureKey,
	"U":                                                                            URL,
	"UMask":                                                                        UMask,
	"URL":                                                                          URL,
	"URLAppendEncoded ":                                                            URLAppendEncoded,
	"URLRedirect":                                                                  URLRedirect,
	"URLRedirectAppend":                                                            URLRedirectAppend,
	"URLRedirectAppendEncoded":                                                     URLRedirectAppendEncoded,
	"UsageLimit":                                                                   UsageLimit,
	"Validate":                                                                     Validate,
	"XDebug":                                                                       XDebug,
}

var LowercaseLabelToDirective = map[string]Directive{} //nolint:gochecknoglobals

func init() {
	for label, directive := range LabelToDirective {
		LowercaseLabelToDirective[strings.ToLower(label)] = directive
	}
}

func (d Directive) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

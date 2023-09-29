package main

import (
	"encoding/json"
	"net/url" // For QueryUnescape
	"regexp"
	"strings"
)

// Profile holds the fields of a TOTP QR Code URI plus a custom title.
type Profile struct {
	URI       string `json:"URI,omitempty"`
	Title     string `json:"title,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
	Account   string `json:"account,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	Digits    string `json:"digits,omitempty"`
	Period    string `json:"period,omitempty"`
}

// String is a Stringer to make a (JSON) string representation of a Profile.
// JSON format is indented and omits blank fields.
func (p Profile) String() string {
	bytes, err := json.MarshalIndent(p, "    ", " ")
	if err != nil {
		// This shouldn't happen
		return "{\"error\":\"something went wrong\"}"
	}
	return strings.TrimLeft(string(bytes), " ")
}

// NewProfileFromURI attempts to initialize a new Profile struct from the label
// and query parameters of a TOTP QR Code URI. The expected URI format is:
//
//	otpauth://totp/<issuer>:<account>?<query-parameters>
//
// Any Profile fields that cannot be initialized from the URI input string will
// be left blank. But, at minimum, the URI field will be set with a copy of the
// URI input string. This intentionally uses lax input validation to allow for
// maximum interoperability. The goal is to help you enroll in 2FA and manage
// the related data. If a service you need to interact with makes weirdly
// formatted URI's, that's okay. After scanning the weird QR code, you can edit
// the resulting temporary profile to fix up the mess.
func NewProfileFromURI(uri string) (p Profile) {
	p = Profile{}
	p.URI = uri
	var issuer1, issuer2 string
	// Remove prefix and split URI into path and query, separated by "?"
	totpQRCodeRE := regexp.MustCompile(`^otpauth://totp/([^?]*)\?(.*)`)
	submatches := totpQRCodeRE.FindStringSubmatch(uri)
	if len(submatches) < 3 {
		// URI does not match the form of otpauth://totop/<label>?<query>
		return
	}
	path := submatches[1]
	// Split query into key=value pairs separated by "&"
	query := strings.Split(submatches[2], "&")
	// Split path into ((issuer)(?:$3A|:)(?:%20)*)(account=user@domain)
	pathRE := regexp.MustCompile(`((.*)(?:%3A|:)(?:%20)*)?(.*)`)
	pathSubmatches := pathRE.FindStringSubmatch(path)
	issuer1 = pathSubmatches[2]
	if unesc, err := url.QueryUnescape(issuer1); err == nil {
		issuer1 = unesc
	}
	p.Account = pathSubmatches[3]
	// Extract query parameter values (secret=, ...)
	for _, v := range query {
		switch {
		case strings.HasPrefix(v, "secret="):
			p.Secret = v[len("secret="):]
		case strings.HasPrefix(v, "issuer="):
			issuer2 = v[len("issuer="):]
			if unesc, err := url.QueryUnescape(issuer2); err == nil {
				issuer2 = unesc
			}
		case strings.HasPrefix(v, "algorithm="):
			p.Algorithm = strings.ToUpper(v[len("algorithm="):])
		case strings.HasPrefix(v, "digits="):
			p.Digits = v[len("digits="):]
		case strings.HasPrefix(v, "period="):
			p.Period = v[len("period="):]
		}
	}
	// According to this wiki in the archived google-authenticator repo,
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format the
	// recommended QR Code TOTP enrollment URI structure includes an an issuer
	// query parameter and a matching issue prefix in the label that comes
	// after `...//totp/`. But, sometimes it might be the case that there's
	// only a prefix in the label. Newer method is to use the parameter but
	// provide the label prefix for backward compatibility.
	if issuer2 != "" {
		p.Issuer = issuer2
	} else {
		p.Issuer = issuer1
	}
	return
}

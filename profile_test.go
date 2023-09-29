package main

import "testing"

// Empty URI should produce empty Profile
func TestURIEmpty(t *testing.T) {
	uri := ""
	ref := Profile{}
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURIScheme(t *testing.T) {
	uri := "otpauth://"
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURISchemeTotp(t *testing.T) {
	uri := "otpauth://totp/"
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURISchemeHotp(t *testing.T) {
	// HOTP is not supported, but it shouldn't cause any errors at this stage
	uri := "otpauth://hotp/"
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURISchemePathQuery1(t *testing.T) {
	uri := "otpauth://totp/?"
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURISchemePathQuery2(t *testing.T) {
	uri := "otpauth://totp/label?query"
	ref := Profile{}
	ref.URI = uri
	ref.Account = "label"
	// The query is missing relevant fields, so it won't set anything
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURIAccount1(t *testing.T) {
	uri := "otpauth://totp/account"
	ref := Profile{}
	ref.URI = uri
	// This won't set the account because URI has no ? query delimiter
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set the account
func TestURIAccount2(t *testing.T) {
	uri := "otpauth://totp/account?"
	ref := Profile{}
	ref.URI = uri
	ref.Account = "account"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// URI with no recognizeable fields should get stored as Profile.URI
func TestURIIssuerAccount1(t *testing.T) {
	uri := "otpauth://totp/issuer:account"
	ref := Profile{}
	ref.URI = uri
	// This won't set the account because URI has no ? query delimiter
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set the issuer and account
func TestURIIssuerAccount2(t *testing.T) {
	uri := "otpauth://totp/issuer:account?"
	ref := Profile{}
	ref.URI = uri
	ref.Account = "account"
	ref.Issuer = "issuer"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set secret
func TestURIQuerySecret(t *testing.T) {
	uri := "otpauth://totp/?secret=JBSWY3DPEHPK3PXP"
	ref := Profile{}
	ref.URI = uri
	ref.Secret = "JBSWY3DPEHPK3PXP"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set issuer (test for issuer in query only)
func TestURIQueryIssuer1(t *testing.T) {
	uri := "otpauth://totp/?issuer=Example"
	ref := Profile{}
	ref.URI = uri
	ref.Issuer = "Example"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set issuer with priority given to the query
func TestURIQueryIssuer2(t *testing.T) {
	// The issuer mismatch is malformed, but the point here is to test that the
	// parser prioritizes the query over the label prefix
	uri := "otpauth://totp/issuer1:account?issuer=issuer2"
	ref := Profile{}
	ref.URI = uri
	ref.Account = "account"
	ref.Issuer = "issuer2"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set issuer (test for matching issuers)
func TestURIQueryIssuer3(t *testing.T) {
	// The issuer mismatch is malformed, but the point here is to test that the
	// parser prioritizes the query over the label prefix
	uri := "otpauth://totp/issuer:account?issuer=issuer"
	ref := Profile{}
	ref.URI = uri
	ref.Account = "account"
	ref.Issuer = "issuer"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set digits to 6
func TestURIQueryDigits6(t *testing.T) {
	uri := "otpauth://totp/?digits=6"
	ref := Profile{}
	ref.URI = uri
	ref.Digits = "6"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set digits to 8
func TestURIQueryDigits8(t *testing.T) {
	uri := "otpauth://totp/?digits=8"
	ref := Profile{}
	ref.URI = uri
	ref.Digits = "8"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should not set digits
func TestURIQueryDigitsNope(t *testing.T) {
	uri := "otpauth://totp/?digits="
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set algorithm to SHA1
func TestURIQueryAlgoritmSHA1(t *testing.T) {
	uri := "otpauth://totp/?algorithm=SHA1"
	ref := Profile{}
	ref.URI = uri
	ref.Algorithm = "SHA1"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set algorithm to SHA256
func TestURIQueryAlgoritmSHA256(t *testing.T) {
	uri := "otpauth://totp/?algorithm=SHA256"
	ref := Profile{}
	ref.URI = uri
	ref.Algorithm = "SHA256"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should not set algorithm
func TestURIQueryAlgoritmNope(t *testing.T) {
	uri := "otpauth://totp/?algorithm="
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set period to 30
func TestURIQueryPeriod30(t *testing.T) {
	uri := "otpauth://totp/?period=30"
	ref := Profile{}
	ref.URI = uri
	ref.Period = "30"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set period to 60
func TestURIQueryPeriod60(t *testing.T) {
	uri := "otpauth://totp/?period=60"
	ref := Profile{}
	ref.URI = uri
	ref.Period = "60"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should not set period
func TestURIQueryPeriodNope(t *testing.T) {
	uri := "otpauth://totp/?period="
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should not set title because it's a custom field, not part of URI spec
func TestURIQueryTitleNope(t *testing.T) {
	uri := "otpauth://totp/?title=not-happening"
	ref := Profile{}
	ref.URI = uri
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set secret, but not period, digits, nor algorithm
func TestURIWellFormed1(t *testing.T) {
	uri := "otpauth://totp/Example:Alice@example?secret=JBSWY3DPEHPK3PXP&issuer=Example"
	ref := Profile{}
	ref.URI = uri
	ref.Issuer = "Example"
	ref.Account = "Alice@example"
	ref.Secret = "JBSWY3DPEHPK3PXP"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

// This should set issuer, secret, period, digits, and algorithm. Note the
// query-unescape on issuer.
func TestURIWellFormed2(t *testing.T) {
	uri := "otpauth://totp/ACME%20Co:john.doe@acme?" +
		"secret=HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ&issuer=ACME%20Co&algorithm=SHA1" +
		"&digits=6&period=30"
	ref := Profile{}
	ref.URI = uri
	ref.Issuer = "ACME Co"
	ref.Account = "john.doe@acme"
	ref.Secret = "HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ"
	ref.Algorithm = "SHA1"
	ref.Digits = "6"
	ref.Period = "30"
	got := NewProfileFromURI(uri)
	if ref != got {
		t.Error("\nwanted:", ref, "\ngot:", got)
	}
}

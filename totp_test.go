package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"strings"
	"testing"
)

type TestVector struct {
	Time int64
	T    string
	TOTP string
	Mode string
}

// TestVectors holds RFC6238 Appendix B SHA1 and SHA256 test vectors. I'm
// ignoring SHA512 because it is not commonly supported. Appendix B states that
// time step is 30 seconds and shared secret is ASCII "12345678901234567890".
// But, the code in Appendix B uses an extended (repeats first 12 bytes)
// version of that key. So, I'm not sure yet about the keys.
var TestVectors []TestVector = []TestVector{
	// Time (sec), Value of T (hex), TOTP, Mode
	{59, "0000000000000001", "94287082", "SHA1"},
	{59, "0000000000000001", "46119246", "SHA256"},
	{1111111109, "00000000023523EC", "07081804", "SHA1"},
	{1111111109, "00000000023523EC", "68084774", "SHA256"},
	{1111111111, "00000000023523ED", "14050471", "SHA1"},
	{1111111111, "00000000023523ED", "67062674", "SHA256"},
	{1234567890, "000000000273EF07", "89005924", "SHA1"},
	{1234567890, "000000000273EF07", "91819424", "SHA256"},
	{2000000000, "0000000003F940AA", "69279037", "SHA1"},
	{2000000000, "0000000003F940AA", "90698825", "SHA256"},
	{20000000000, "0000000027BC86AA", "65353130", "SHA1"},
	{20000000000, "0000000027BC86AA", "77737706", "SHA256"},
}

var key1 string = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
var key256 string = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA===="

func Test_RFC6238_Appendix_B_test_vectors(t *testing.T) {
	for i, v := range TestVectors {
		var secret string
		switch v.Mode {
		case "SHA1":
			secret = key1
		case "SHA256":
			secret = key256
		default:
			t.Error("Unsupported mode:", v.Mode, "i:", i)
		}
		digits := "8"
		algorithm := v.Mode
		period := "30"
		totp, err := NewTotp(secret, digits, algorithm, period)
		if err != nil {
			t.Error(err)
		}
		code, _, err := totp.CodeAtTime(v.Time)
		if err != nil {
			t.Error(err)
		}
		if v.TOTP != code {
			t.Error("\ni:", i, "\nwanted:", v.TOTP, "\ngot:", code)
		}
	}
}

type RFC2202 struct {
	Key    []byte
	Data   []byte
	Digest []byte
}

var RFC2202_Cases []RFC2202 = []RFC2202{
	{[]byte{0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b},
		[]byte("Hi There"),
		[]byte{0xb6, 0x17, 0x31, 0x86, 0x55, 0x05, 0x72, 0x64, 0xe2, 0x8b,
			0xc0, 0xb6, 0xfb, 0x37, 0x8c, 0x8e, 0xf1, 0x46, 0xbe, 0x00},
	},
	{[]byte("Jefe"),
		[]byte("what do ya want for nothing?"),
		[]byte{0xef, 0xfc, 0xdf, 0x6a, 0xe5, 0xeb, 0x2f, 0xa2, 0xd2, 0x74,
			0x16, 0xd5, 0xf1, 0x84, 0xdf, 0x9c, 0x25, 0x9a, 0x7c, 0x79},
	},
}

// Make we're using HMAC-SHA-1 properly according to RFC2202 test vectors
func Test_RFC2202_HMAC_SHA1_test_cases(t *testing.T) {
	for i, v := range RFC2202_Cases {
		h := hmac.New(sha1.New, v.Key)
		h.Write(v.Data)
		digest := h.Sum(nil)
		if !bytes.Equal(digest, v.Digest) {
			t.Error("\ni:", i, "\nwanted:", v.Digest, "\ngot:", digest)
		}
	}
}

// Attempting to use a blank secret should fail
func Test_blank_secret(t *testing.T) {
	secret := ""
	digits := ""
	algorithm := ""
	period := ""
	wantError := "Secret value is blank."
	_, err := NewTotp(secret, digits, algorithm, period)
	if err != nil && !strings.Contains(err.Error(), wantError) {
		t.Error("\nwanted:", wantError, "\ngot:", err.Error())
	}
}

// Attempting to use an unsupported period should fail
func Test_unsupported_period(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	period := "29"
	wantError := "Period should be empty, \"30\", or \"60\"."
	_, err := NewTotp(secret, "", "", period)
	if err != nil && !strings.Contains(err.Error(), wantError) {
		t.Error("\nwanted:", wantError, "\ngot:", err.Error())
	}
}

// Attempting to use an unsupported algorithm should fail
func Test_unsupported_algorithm(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	algorithm := "MD5"
	wantError := "Algorithm should be empty, \"SHA1\", or \"SHA256\"."
	_, err := NewTotp(secret, "", algorithm, "")
	if err != nil && !strings.Contains(err.Error(), wantError) {
		t.Error("\nwanted:", wantError, "\ngot:", err.Error())
	}
}

// Attempting to use an unsupported digits value should fail
func Test_unsupported_digits(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	digits := "7"
	wantError := "Digits should be empty, \"6\", or \"8\"."
	_, err := NewTotp(secret, digits, "", "")
	if err != nil && !strings.Contains(err.Error(), wantError) {
		t.Error("\nwanted:", wantError, "\ngot:", err.Error())
	}
}

// Attempting to use supported digits values should work
func Test_supported_digits(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	for _, digits := range []string{"6", "8"} {
		_, err := NewTotp(secret, digits, "", "")
		if err != nil {
			t.Error("\ntried:", digits, "\ngot:", err.Error())
		}
	}
}

// Attempting to use supported algorithms should work
func Test_supported_algorithms(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	for _, algorithm := range []string{"SHA1", "SHA256"} {
		_, err := NewTotp(secret, "", algorithm, "")
		if err != nil {
			t.Error("\ntried:", algorithm, "\ngot:", err.Error())
		}
	}
}

// Attempting to use supported periods should work
func Test_supported_periods(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	for _, period := range []string{"30", "60"} {
		_, err := NewTotp(secret, "", "", period)
		if err != nil {
			t.Error("\ntried:", period, "\ngot:", err.Error())
		}
	}
}

// Lowercase secret should work
func Test_lowercase_secret(t *testing.T) {
	secret := "jbswy3dpehpk3pxp"
	_, err := NewTotp(secret, "", "", "")
	if err != nil {
		t.Error("\ntried:", secret, "\ngot:", err.Error())
	}
}

// Attempting to use non-multiple-of-8-length secret values should work, as
// long as the values are valid base32. Note that you can't just make an
// arbitrary length string from base32 symbols and expect the result will parse
// as a valid base32 string. Go's base32 decoder seems to want the strings to
// be padded with "=" at the end if they aren't otherwise a multiple of 8
// characters long. If you give the decoder a string that's an exact multiple
// of 8 base32 symbols long, probably it will be fine. If you give a "=" padded
// string that is a multiple of 8 long, but with an odd number of base32
// symbols before the "=", it may give an error. If you give weird stuff like
// "AB" that an encoder wouldn't emit, it still may parse without errors.
func Test_weird_but_valid_secrets(t *testing.T) {
	notMultEightButWellFormed := []string{
		"AA", "AE", "AI", "AM", "AQ", "AU", "AY", "A4======", "ABCDEFGHAA",
		"ABAA====", "ACAA===="}
	for _, secret := range notMultEightButWellFormed {
		_, err := NewTotp(secret, "", "", "")
		if err != nil {
			t.Error("\ntried:", secret, "\ngot:", err.Error())
		}
	}
	// Is this a bug in Go's decoder? These have some dangling bits on the end
	// but the decoder doesn't seem to mind?
	badlyEncodedButStillPassing := []string{
		"AB", "AC", "AD"}
	for _, secret := range badlyEncodedButStillPassing {
		_, err := NewTotp(secret, "", "", "")
		if err != nil {
			t.Error("\ntried:", secret, "\ngot:", err.Error())
		}
	}
	// The decoder does not like odd numbers of base32 symbols
	badlyEncodedFailing := []string{
		"ABA", "ACA", "ADA"}
	for _, secret := range badlyEncodedFailing {
		totp, err := NewTotp(secret, "", "", "")
		if err == nil {
			t.Error("\nWanted decoder error:", secret, "\ngot:", totp.Secret)
		}
	}
}

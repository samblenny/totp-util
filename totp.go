package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"net/url" // for QueryUnescape()
	"strings"
	"time"
)

// HmacAlgo is an enum representing supported HMAC hash algorithms
type HmacAlgo int

const (
	HmacSha1 HmacAlgo = iota
	HmacSha256
)

func (h HmacAlgo) String() string {
	switch h {
	case HmacSha1:
		return "SHA1"
	case HmacSha256:
		return "SHA256"
	}
	return "ERROR"
}

// Struct Totp holds the parameters needed to compute a TOTP code
type Totp struct {
	Secret    []byte
	Digits    int
	Algorithm HmacAlgo
	Period    int
}

// NewTotp attempts to create a Totp instance with the requested parameters.
// If the parameters fail the validation checks, NewTotp returns an error.
// In case of unspecified parameters, defaults are:
//   - digits = 6
//   - algorithm = SHA1
//   - period = 30
//
// Those values are intended to match the defaults documented on the wiki at
// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
func NewTotp(secret, digits, algorithm, period string) (*Totp, error) {
	t := Totp{}
	msg := ""
	// Validate digits=...
	switch digits {
	case "", "6":
		t.Digits = 6
	case "8":
		t.Digits = 8
	default:
		msg += " Digits should be empty, \"6\", or \"8\"."
	}
	// Validate algorithm=...
	switch algorithm {
	case "", "SHA1":
		t.Algorithm = HmacSha1
	case "SHA256":
		t.Algorithm = HmacSha256
	default:
		msg += " Algorithm should be empty, \"SHA1\", or \"SHA256\"."
	}
	// Validate period=...
	switch period {
	case "", "30":
		t.Period = 30
	case "60":
		t.Period = 60
	default:
		msg += " Period should be empty, \"30\", or \"60\"."
	}
	// Validate base32 secret
	// Secrets ideally shouldn't end with "=", and they really shouldn't end
	// with a "%3D" url-escaped "=". But, I've seen authenticator app bug
	// reports about TOTP QR Code URI parsing failures for secrets that do end
	// with some "=" or "%3D". So, be cautious and start with a url-unescape.
	// See previously mentioned documentation wiki page and RFC3548 §2.2:
	//  https://datatracker.ietf.org/doc/html/rfc3548#section-2.2
	if unescapedSecret, err := url.QueryUnescape(secret); err != nil {
		msg += fmt.Sprintf(
			" Secret value is weird (query unescape failed: \"%v\", %v).",
			secret, err.Error())
	} else if secret == "" {
		msg += " Secret value is blank."
	} else {
		// The wiki URI docs say the "=" suffix padding is not needed, but Go's
		// base32 decoder seems to want padding for strings that are not an
		// exact multiple of eight characters. So, add padding.
		padLen := 8 - (len(unescapedSecret) % 8)
		if padLen == 8 {
			padLen = 0
		}
		for i := 0; i < padLen; i++ {
			unescapedSecret += "="
		}
		// Base32 decoder wants uppercase, but some TOTP QR Codes use lowercase
		unescapedSecret := strings.ToUpper(unescapedSecret)
		// Now decode the padded base32
		t.Secret, err = base32.StdEncoding.DecodeString(unescapedSecret)
		if err != nil {
			msg += fmt.Sprintf(
				" Secret value is weird (base32 decode failed: \"%v\", %v).",
				unescapedSecret, err.Error())
		}
	}
	// Bail out with an error if any of the validation checks failed
	if msg != "" {
		return nil, errors.New(msg)
	}
	// Yay, all good...
	return &t, nil
}

// TotpCode returns a string with the TOTP code for the given Unix timestamp.
// The configurable timestamp is meant to support tests that check for the
// text vectors from RFC6238 Appendix B.
func (t Totp) CodeAtTime(unixTime int64) (
	code string, validSeconds int, err error) {
	// Because Go doesn't support proper enum types, this switch combines type
	// validation with setting up the HMAC hashers using crypto/hmac,
	// crypto/sha1, and crypto/sha256. The related docs are a little thin, but
	// you can read them here:
	//  - https://pkg.go.dev/crypto/hmac
	//  - https://pkg.go.dev/hash#Hash
	//  - https://pkg.go.dev/crypto/sha1@go1.21.1
	//  - https://pkg.go.dev/crypto/sha256@go1.21.1
	var h hash.Hash
	switch t.Algorithm {
	case HmacSha1:
		h = hmac.New(sha1.New, t.Secret)
	case HmacSha256:
		h = hmac.New(sha256.New, t.Secret)
	default:
		err = errors.New("Unsupported algorithm value")
		return
	}

	// Notes from RFC6238 (TOTP):
	//  - Summarizing §4.1 and §4.2: The time (T) to be fed into the HMAC
	//    hasher is calculated as:
	//     T = current_unix_time / time_step
	//    where / is integer division that acts like floor()
	//  - §4.2: On the 2038 problem:
	//     > The implementation of this algorithm MUST support a time value T
	//     > larger than a 32-bit integer when it is beyond the year 2038.
	//  - Appendix A: Timestamp is converted to 16-digit 0-padded hex string
	//    before being converted to bytes and then fed to the HMAC hash. Not
	//    sure why they took a detour through the whole hex string thing
	//    instead of just spelling out clearly that you should feed the time
	//    into the hash as big-endian bytes from an int64. Oh well.
	//  - Appendix A: Converting HMAC output bytes to a decimal code selects a
	//    31 bit integer from the output according to a window offset that is
	//    calculated from the high byte of the HMAC output:
	//     > int offset = hash[hash.length - 1] & 0xf;
	//     > int binary =
	//     >     ((hash[offset] & 0x7f) << 24) |
	//     >     ((hash[offset + 1] & 0xff) << 16) |
	//     >     ((hash[offset + 2] & 0xff) << 8) |
	//     >     (hash[offset + 3] & 0xff);
	//   - RFC4226 (HOTP) §5.3 and §5.4 give a longer explanation of the HMAC
	//     and "dynamic truncation" algorithm. RFC6238 (TOTP) seems to assume
	//     you already know how that stuff worked from reading about HOTP.
	//

	// Floor the Unix timestamp with the 30 or 60 second period ("time-step").
	// The RFC calls it time-step. The Google Authenticator QR code URI format
	// cals it period. Whatever. I'll call it both.
	var timeStep int64
	switch t.Period {
	case 30, 60:
		timeStep = int64(t.Period)
	default:
		err = errors.New("Unsupported period value")
		return
	}
	floorTime := unixTime / timeStep
	validSeconds = int(timeStep - (unixTime % timeStep))

	// HMAC hash the floored timestamp as a big-endian int64. Do not be fooled
	// by the hex timestamp stuff in the RFC6238 sample code. You're not
	// supposed to hash the hex strings. Big-endian int64 is the way.
	binary.Write(h, binary.BigEndian, floorTime)
	hmacOut := h.Sum(nil)

	// Do the dynamic truncation thing from the RFC
	offset := int(hmacOut[len(hmacOut)-1]) & 0xf
	// Be paranoid and redundantly assert that the offset window falls inside
	// the HMAC buffer's length. The algorithm should be SHA1 or SHA256. SHA1
	// should output 20 bytes and SHA256 should output 32 bytes, offset should
	// be in range 0..15, and 15+3=18 fits in 20 and 32. But, that's a lot of
	// should. Perhaps a bug invalidated one of those assumed truths, so check.
	if offset < 0 || offset >= len(hmacOut) || 0xf+3 >= len(hmacOut) {
		err = errors.New("HMAC output buffer selection window OOR")
		return
	}
	// Extract the 31-bit integer like in the RFC (note the & 0x7f)
	n := int64(hmacOut[offset]&0x7f) << 24
	n |= int64(hmacOut[offset+1]&0xff) << 16
	n |= int64(hmacOut[offset+2]&0xff) << 8
	n |= int64(hmacOut[offset+3] & 0xff)
	// Do % (10^digits) so the final code is the right number of digits
	switch t.Digits {
	case 6:
		n %= 1000000
		code = fmt.Sprintf("%06d", n)
	case 8:
		n %= 100000000
		code = fmt.Sprintf("%08d", n)
	default:
		err = errors.New("Unsupported digits value")
		return
	}
	return
}

// TotpCode returns a string with the TOTP code for the current Unix time
func (t Totp) CurrentCode() (string, int, error) {
	return t.CodeAtTime(time.Now().Unix())
}

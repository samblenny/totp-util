package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"testing"
)

type TestVector struct {
	Time int64
	T    string
	TOTP string
	Mode string
}

// TestVectors holds RFC6238 Appendix B SHA1 and SHA256 test vectors. I'm
// ignoring SHA512 because it is not commonly supoorted. Appendix B states that
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

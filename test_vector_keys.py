#!/usr/bin/python3
# Convert keys for test vectors from RFC6238 Appendix B into base32.

import base64

k1_int     = 0x3132333435363738393031323334353637383930
k256_int   = 0x3132333435363738393031323334353637383930313233343536373839303132
k1_ascii   = k1_int.to_bytes(length=20, byteorder='big')
k256_ascii = k256_int.to_bytes(length=32, byteorder='big')
k1_b32     = base64.b32encode(k1_ascii)
k256_b32   = base64.b32encode(k256_ascii)

print("k1_int:    ", hex(k1_int))
print("k256_int:  ", hex(k256_int))
print("k1_ascii:  ", k1_ascii)
print("k256_ascii:", k256_ascii)
print("k1_b32:    ", k1_b32)
print("k256_b32:  ", k256_b32)

# k1_int:     0x3132333435363738393031323334353637383930
# k256_int:   0x3132333435363738393031323334353637383930313233343536373839303132
# k1_ascii:   b'12345678901234567890'
# k256_ascii: b'12345678901234567890123456789012'
# k1_b32:     b'GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ'
# k256_b32:   b'GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA===='

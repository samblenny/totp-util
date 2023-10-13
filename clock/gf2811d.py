#!/usr/bin/python3
"""
Experiment with arithemetic in Galois Field GF(2^8), prime polynomial 0x11d
"""

class GF2811d:
    """Galois Field GF(2^8), generator 𝛼=2, prime polynomial=0x11d"""

    def __init__(self):
        """Compute logarithm and exponential tables"""
        self.log_lut = [0] * 256
        self.exp_lut = [1] * 256
        n = 1
        for i in range(1,256):
            n = (n << 1) ^ (((n>>7) & 1) * 0x11d)  # Multiply n by 𝛼=2 mod 11d
            self.exp_lut[i] = n
            self.log_lut[n] = i

    def log(self, x):
        """Logarithm for domain 0..255 with log(0) = 0.
        Defining log(0) = 0 lets calling code be simpler by avoiding the need
        for conditional logic to handle 0 as a special case.
        """
        if (x < 0) or (x > 255):
            raise Exception(f"x={x} is out of range for log(x)")
        return self.log_lut[x]

    def exp(self, x):
        """Exponential for domain 0..510 to allow for exp(log(a) + log(b))"""
        if (x < 0) or (x > 510):
            raise Exception(f"x={x} is out of range for exp(x)")
        return self.exp_lut[(x & 255) + ((x>>8)&1)]

    def mul(self, x, y):
        """Multiply using logarithms"""
        if (x == 0) or (y == 0):
            return 0
        return self.exp(self.log(x) + self.log(y))

    def hw_mul(self, x, y):
        """Multiply approximating hardware AND and XOR gate style"""
        n = 0
        for i in range(8):
            n ^= ((y>>i)&1) * (x<<i)
        for i in range(7,-1,-1):
            n ^= ((n>>(8+i))&1) * (0x11d << i)
        return n


class ReedSolomon:
    """
    Implement Reed-Solomon encoder using GF(2^8), 𝛼=2, prime polynomial 0x11d,
    and generator polynomials from Table A.1 of ISO/IEC 18004:2015 Annex A.
    """

    def __init__(self, version):
        """
        Initialize encoder for QR Code version 1 with 26 total codewords. The
        codewords are split between data and error correction according to the
        L, M, Q, and H error correction levels. The generator polynomial
        coefficients are exponents (the base 𝛼=2 log over GF(2^8) of the
        actual coefficients) to match Table A.1 of ISO/IEC 18004:2005. The
        coefficients of the highest order terms are always 1 (𝛼^0=1), so they
        are are omitted.
        """
        if version == "1-L":
            self.ecc_len = 7
            self.gen_poly = [0, 87, 229, 146, 149, 238, 102, 21]
        elif version == "1-M":
            self.ecc_len = 10
            self.gen_poly = [0, 251, 67, 46, 61, 118, 70, 64, 94, 32, 45]
        elif version == "1-Q":
            self.ecc_len = 13
            self.gen_poly = [0, 74, 152, 176, 100, 86, 100, 106, 104, 130, 218,
                206, 140, 78]
        elif version == "1-H":
            self.ecc_len = 17
            self.gen_poly = [0, 43, 139, 206, 78, 43, 239, 123, 206, 214, 147,
                24, 99, 150, 39, 243, 163, 136]
        else:
            raise Exception(f"version \"{version}\" is not supported")
        self.gf = GF2811d()
        self.v = version
        self.d_len = 26 - self.ecc_len  # Version 1 total bytes = 26
        if self.d_len + self.ecc_len != 26:
            raise Exception("d_len does not match ecc_len")

    def encode(self, data):
        """
        Encode data as a Reed-Solomon message. Data should be the exact length
        of codewords for the selected QR code version and ECC level. Returned
        array will be data + ecc codewords.
        """
        dl = len(data)
        if self.d_len != dl:
            e = f"{self.v} needs {self.d_len} data bytes, got {dl}"
            raise Exception(e)
        # Create dividend array with data + zeros for the ECC bytes
        m = data + ([0] * self.ecc_len)
        # Divide m by the generator polynomial to get ECC bytes (the remainder)
        for i in range(dl):
            # Leading coefficient of generator polynomial g(x) is always 1, so
            # each iteration, we subtract (xor) m[i]*g(x)*x^(n-i) from m
            log_a = self.gf.log(m[i])
            if log_a == 0:
                continue
            for (j, g_coeff) in enumerate(self.gen_poly):
                # Since g(x) coefficients are already in log form, we can save
                # some work by inlining mul(a, b) = exp(log(a) + log(b)).
                # Remember that log(leading coefficient) = 0.
                m[i+j] ^= self.gf.exp(log_a + g_coeff)
        # Now that the ECC remainder has been computed, copy the data bytes back
        # over the front of the array to form a complete message
        for (i, x) in enumerate(data):
            m[i] = x
        return m


if __name__ == "__main__":
    import unittest

    class TestGF2811d(unittest.TestCase):
        def test_lookup_tables(self):
            gf = GF2811d()
            self.assertEqual(gf.exp_lut,
                [1, 2, 4, 8, 16, 32, 64, 128, 29, 58, 116, 232, 205, 135, 19,
                38, 76, 152, 45, 90, 180, 117, 234, 201, 143, 3, 6, 12, 24, 48,
                96, 192, 157, 39, 78, 156, 37, 74, 148, 53, 106, 212, 181, 119,
                238, 193, 159, 35, 70, 140, 5, 10, 20, 40, 80, 160, 93, 186,
                105, 210, 185, 111, 222, 161, 95, 190, 97, 194, 153, 47, 94,
                188, 101, 202, 137, 15, 30, 60, 120, 240, 253, 231, 211, 187,
                107, 214, 177, 127, 254, 225, 223, 163, 91, 182, 113, 226, 217,
                175, 67, 134, 17, 34, 68, 136, 13, 26, 52, 104, 208, 189, 103,
                206, 129, 31, 62, 124, 248, 237, 199, 147, 59, 118, 236, 197,
                151, 51, 102, 204, 133, 23, 46, 92, 184, 109, 218, 169, 79,
                158, 33, 66, 132, 21, 42, 84, 168, 77, 154, 41, 82, 164, 85,
                170, 73, 146, 57, 114, 228, 213, 183, 115, 230, 209, 191, 99,
                198, 145, 63, 126, 252, 229, 215, 179, 123, 246, 241, 255, 227,
                219, 171, 75, 150, 49, 98, 196, 149, 55, 110, 220, 165, 87,
                174, 65, 130, 25, 50, 100, 200, 141, 7, 14, 28, 56, 112, 224,
                221, 167, 83, 166, 81, 162, 89, 178, 121, 242, 249, 239, 195,
                155, 43, 86, 172, 69, 138, 9, 18, 36, 72, 144, 61, 122, 244,
                245, 247, 243, 251, 235, 203, 139, 11, 22, 44, 88, 176, 125,
                250, 233, 207, 131, 27, 54, 108, 216, 173, 71, 142, 1])
            self.assertEqual(gf.log_lut,
                [0, 255, 1, 25, 2, 50, 26, 198, 3, 223, 51, 238, 27, 104, 199,
                75, 4, 100, 224, 14, 52, 141, 239, 129, 28, 193, 105, 248, 200,
                8, 76, 113, 5, 138, 101, 47, 225, 36, 15, 33, 53, 147, 142,
                218, 240, 18, 130, 69, 29, 181, 194, 125, 106, 39, 249, 185,
                201, 154, 9, 120, 77, 228, 114, 166, 6, 191, 139, 98, 102, 221,
                48, 253, 226, 152, 37, 179, 16, 145, 34, 136, 54, 208, 148,
                206, 143, 150, 219, 189, 241, 210, 19, 92, 131, 56, 70, 64, 30,
                66, 182, 163, 195, 72, 126, 110, 107, 58, 40, 84, 250, 133,
                186, 61, 202, 94, 155, 159, 10, 21, 121, 43, 78, 212, 229, 172,
                115, 243, 167, 87, 7, 112, 192, 247, 140, 128, 99, 13, 103, 74,
                222, 237, 49, 197, 254, 24, 227, 165, 153, 119, 38, 184, 180,
                124, 17, 68, 146, 217, 35, 32, 137, 46, 55, 63, 209, 91, 149,
                188, 207, 205, 144, 135, 151, 178, 220, 252, 190, 97, 242, 86,
                211, 171, 20, 42, 93, 158, 132, 60, 57, 83, 71, 109, 65, 162,
                31, 45, 67, 216, 183, 123, 164, 118, 196, 23, 73, 236, 127, 12,
                111, 246, 108, 161, 59, 82, 41, 157, 85, 170, 251, 96, 134,
                177, 187, 204, 62, 90, 203, 89, 95, 176, 156, 169, 160, 81, 11,
                245, 22, 235, 122, 117, 44, 215, 79, 174, 213, 233, 230, 231,
                173, 232, 116, 214, 244, 234, 168, 80, 88, 175])

        def test_exp(self):
            gf = GF2811d()
            x_values = [0, 1, 2, 3, 254, 255, 256, 257, 258, 509, 510]
            y_values = [gf.exp(x) for x in x_values]
            self.assertEqual(y_values, [1, 2, 4, 8, 142, 1, 2, 4, 8, 142, 1])
            for x in [-1, 511]:
                self.assertRaises(Exception, gf.exp, x)

        def test_log(self):
            gf = GF2811d()
            x_values = [0, 1, 2, 3, 4, 253, 254, 255]
            y_values = [gf.log(x) for x in x_values]
            self.assertEqual(y_values, [0, 255, 1, 25, 2, 80, 88, 175])
            for x in [-1, 256]:
                self.assertRaises(Exception, gf.log, x)

        def test_mul_implementations_agree(self):
            """Log multiply and hardware-gate-style multiply should agree"""
            gf = GF2811d()
            for a in range(256):
                for b in range(a,256):
                    self.assertEqual(gf.hw_mul(a, b), gf.mul(a, b))
            wanted = [0, 0, 0]
            got = [gf.mul(0, x) for x in [0, 5, 99]]
            self.assertEqual(wanted, got)
            pairs = [(2, 5), (3, 3), (3, 7), (5, 5)]
            wanted = [10, 5, 9, 17]
            got = [gf.mul(a, b) for (a, b) in pairs]
            self.assertEqual(wanted, got)


    class TestReedSolomon(unittest.TestCase):
        def test_init(self):
            with self.assertRaises(Exception):
                rs = ReedSolomon("2-Q")
            rs = ReedSolomon("1-L")
            self.assertEqual(rs.ecc_len, 7)
            self.assertEqual(rs.ecc_len, len(rs.gen_poly)-1)
            self.assertEqual(rs.d_len, 19)
            rs = ReedSolomon("1-M")
            self.assertEqual(rs.ecc_len, 10)
            self.assertEqual(rs.ecc_len, len(rs.gen_poly)-1)
            self.assertEqual(rs.d_len, 16)
            rs = ReedSolomon("1-Q")
            self.assertEqual(rs.ecc_len, 13)
            self.assertEqual(rs.ecc_len, len(rs.gen_poly)-1)
            self.assertEqual(rs.d_len, 13)
            rs = ReedSolomon("1-H")
            self.assertEqual(rs.ecc_len, 17)
            self.assertEqual(rs.ecc_len, len(rs.gen_poly)-1)
            self.assertEqual(rs.d_len, 9)

        def test_encode_1L(self):
            rs = ReedSolomon("1-L")
            # Convert bytes to list of bytes
            msg = "this is a message__"
            data = [b for b in msg.encode()]  # Convert bytes to list of bytes
            encoded = rs.encode(data)
            enc_data = encoded[:len(data)]
            enc_ecc = encoded[len(data):]
            self.assertEqual(enc_data, data)
            self.assertEqual(enc_ecc, [102, 171, 195, 197, 71, 73, 243])

        def test_encode_1H(self):
            rs = ReedSolomon("1-H")
            msg = "short msg"
            data = [b for b in msg.encode()]  # Convert bytes to list of bytes
            encoded = rs.encode(data)
            enc_data = encoded[:len(data)]
            enc_ecc = encoded[len(data):]
            self.assertEqual(enc_data, data)
            self.assertEqual(enc_ecc, [186, 198, 18, 201, 195, 189, 127, 28,
                146, 122, 203, 121, 139, 46, 52, 177, 235])


    print("Testing exp and log for GF(2^8), 𝛼=2, prime polynomial=0x11d")
    print("Testing Reed-Solomon encoding")
    unittest.main()

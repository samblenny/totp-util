#!/usr/bin/python3

class ReedSolomon:
    """
    Reed-Solomon encoder using GF(2^8), 𝛼=2, prime polynomial 0x11d, and
    generator polynomial from ISO/IEC 18004:2015 Annex A.
    """

    def __init__(self):
        """Compute logarithm and exponential tables for GF(2^8), poly 0x11d"""
        self.log_lut = [0] * 256
        self.exp_lut = [1] * 256
        n = 1
        for i in range(1,256):
            n = (n << 1) ^ (((n>>7) & 1) * 0x11d)  # Multiply n by 𝛼=2 mod 11d
            self.exp_lut[i] = n
            self.log_lut[n] = i

    def log(self, x):
        """Logarithm for domain 0..255 with log(0) = 0."""
        if (x < 0) or (x > 255):
            raise Exception(f"x={x} is out of range for log(x)")
        return self.log_lut[x]

    def exp(self, x):
        """Exponential for domain 0..510 to allow for exp(log(a) + log(b))"""
        if (x < 0) or (x > 510):
            raise Exception(f"x={x} is out of range for exp(x)")
        return self.exp_lut[(x & 255) + ((x>>8)&1)]

    def ecc_remainder(self, data):
        """
        Encode Reed-Solomon ECC remainder for QR Code version 1-M data.
        Generator polynomial coefficients are in logarithm form (exponents of
        base 𝛼=2 over GF(2^8)) to match Table A.1 of ISO/IEC 18004:2005.

        Notes on dividing message m (data + ecc) by generator polynomial g(x):
        1. Leading coefficient of g(x) is always 1 (0 in log form), so each
           iteration, we subtract (xor) m[i]*g(x)*x^(n-i) from m
        2. Since g(x) coefficients are already in log form, we can inline and
           simplify the multiplication m[i]*g(x) = exp(log(m[i]) + log(g(x)))
        """
        version = "1-M"
        ecc_bytes = 10
        data_bytes = 16
        generator_polynomial = [0, 251, 67, 46, 61, 118, 70, 64, 94, 32, 45]
        if len(data) != data_bytes:
            e = f"{version} needs {data_bytes} data bytes, got {len(data)}"
            raise Exception(e)
        # Create dividend polynomial
        m = data + ([0] * ecc_bytes)
        for i in range(len(data)):
            log_m = self.log(m[i])
            if log_m == 0:
                continue
            for (j, log_coefficient) in enumerate(generator_polynomial):
                m[i+j] ^= self.exp(log_m + log_coefficient)
        # Return the ECC remainder bytes
        return m[data_bytes:]


if __name__ == "__main__":
    import unittest

    class TestReedSolomon(unittest.TestCase):
        def setUp(self):
            self.rs = ReedSolomon()

        def test_lookup_tables(self):
            self.assertEqual(self.rs.exp_lut,
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
            self.assertEqual(self.rs.log_lut,
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
            x_values = [0, 1, 2, 3, 254, 255, 256, 257, 258, 509, 510]
            y_values = [self.rs.exp(x) for x in x_values]
            self.assertEqual(y_values, [1, 2, 4, 8, 142, 1, 2, 4, 8, 142, 1])
            for x in [-1, 511]:
                self.assertRaises(Exception, self.rs.exp, x)

        def test_log(self):
            x_values = [0, 1, 2, 3, 4, 253, 254, 255]
            y_values = [self.rs.log(x) for x in x_values]
            self.assertEqual(y_values, [0, 255, 1, 25, 2, 80, 88, 175])
            for x in [-1, 256]:
                self.assertRaises(Exception, self.rs.log, x)

        def test_encode(self):
            # Convert bytes to list of bytes
            msg = "a short message."
            data = [b for b in msg.encode()]  # Convert bytes to list of bytes
            ecc_bytes = self.rs.ecc_remainder(data)
            self.assertEqual(ecc_bytes,
                [71, 193, 109, 98, 162, 8, 33, 119, 105, 74])


    print("Testing Reed-Solomon encoding")
    unittest.main()

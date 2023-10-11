#!/usr/bin/python3
"""
Experiment with Galois Field GF(2^8) logarithmic multiplier
"""

class GF2811d:
    """Galois Field GF(2^8), base 𝛼=2, prime polynomial=0x11d"""

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
        """Logarithm for domain 1..255"""
        return self.log_lut[x]

    def exp(self, x):
        """Exponential for domain 0..510 to allow for exp(log(a) + log(b))"""
        return self.exp_lut[(x & 255) + ((x>>8)&1)]

    def mul(self, x, y):
        return self.exp(self.log(x) + self.log(y))


print("Exponentials and logarithms for GF(2^8), 𝛼=2, prime polynomial=0x11d")
gf = GF2811d()
print("\nexp:", gf.exp_lut)
print("\nlog:", gf.log_lut)
print()
for a in range(1,10):
    for b in range(a,10):
        print(f"{a} × {b} = {gf.mul(a, b)}")
print()
for i in [1,2,3,254,255,256,257,258,509,510]:
    print(f"exp({i:3d}) = {gf.exp(i):3d}")
print()
for i in [1,2,3,4,254,255]:
    print(f"log({i:3d}) = {gf.log(i)}")

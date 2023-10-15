#!/usr/bin/python3
"""
QR Code Mask Pattern Generator

This script generates Uint32 arrays representing the 8 possible mask patterns
for a version 1 QR code (21 x 21 modules, not counting quiet zone).

Each row of the QR code is encoded as 21 bits. The leftmost QR code module of
a row goes in the least significant bit of a Uint32. This is meant to make it
easy to do things like `whatever |= (bit << x)`. The top row of the QR code
goes in the first Uint32 of each mask array.
"""

# This mask contains 1 for modules in the encoding region (symbol characters)
clipMask = [
    0x001e00, 0x001e00, 0x001e00, 0x001e00,
    0x001e00, 0x001e00, 0x000000, 0x001e00, 0x001e00,
    0x1fffbf, 0x1fffbf, 0x1fffbf, 0x1fffbf,
    0x1ffe00, 0x1ffe00, 0x1ffe00, 0x1ffe00,
    0x1ffe00, 0x1ffe00, 0x1ffe00, 0x1ffe00]

# Compute a mask pattern for the given mask function (lambda)
def makeMask(num, mask_function):
    print(f"    static mask{num} = [\n        ", end="")
    for i in range(21):
        mask = 0
        for j in range(21):
            maskBit = 1 if mask_function(i, j) else 0
            mask |= maskBit << j
        mask &= clipMask[i]
        print(f" 0x{mask:06x},", end="")
        if i % 6 == 5:
            print("\n        ", end="")
    print(" ];")

# Compute mask arrays for the 8 QR code mask patterns according to the
# conditions specified in ISO/IEC 18004:2015 §7.8.2 "Data mask patterns"
makeMask(0, lambda i, j: (i+j)%2 == 0)
makeMask(1, lambda i, j: i%2 == 0)
makeMask(2, lambda i, j: j%3 == 0)
makeMask(3, lambda i, j: (i+j)%3 == 0)
makeMask(4, lambda i, j: ((i//2)+(j//3)) % 2 == 0)
makeMask(5, lambda i, j: ((i*j)%2) + ((i*j)%3) == 0)
makeMask(6, lambda i, j: ( ((i*j)%2) + ((i*j)%3) ) % 2 == 0)
makeMask(7, lambda i, j: ( ((i+j)%2) + ((i*j)%3) ) % 2 == 0)

"""
    static mask0 = [
         0x001400, 0x000a00, 0x001400, 0x000a00, 0x001400, 0x000a00,
         0x000000, 0x000a00, 0x001400, 0x0aaaaa, 0x155515, 0x0aaaaa,
         0x155515, 0x0aaa00, 0x155400, 0x0aaa00, 0x155400, 0x0aaa00,
         0x155400, 0x0aaa00, 0x155400, ];
    static mask1 = [
         0x001e00, 0x000000, 0x001e00, 0x000000, 0x001e00, 0x000000,
         0x000000, 0x000000, 0x001e00, 0x000000, 0x1fffbf, 0x000000,
         0x1fffbf, 0x000000, 0x1ffe00, 0x000000, 0x1ffe00, 0x000000,
         0x1ffe00, 0x000000, 0x1ffe00, ];
    static mask2 = [
         0x001200, 0x001200, 0x001200, 0x001200, 0x001200, 0x001200,
         0x000000, 0x001200, 0x001200, 0x049209, 0x049209, 0x049209,
         0x049209, 0x049200, 0x049200, 0x049200, 0x049200, 0x049200,
         0x049200, 0x049200, 0x049200, ];
    static mask3 = [
         0x001200, 0x000800, 0x000400, 0x001200, 0x000800, 0x000400,
         0x000000, 0x000800, 0x000400, 0x049209, 0x124924, 0x092492,
         0x049209, 0x124800, 0x092400, 0x049200, 0x124800, 0x092400,
         0x049200, 0x124800, 0x092400, ];
    static mask4 = [
         0x001000, 0x001000, 0x000e00, 0x000e00, 0x001000, 0x001000,
         0x000000, 0x000e00, 0x001000, 0x1c7187, 0x038e38, 0x038e38,
         0x1c7187, 0x1c7000, 0x038e00, 0x038e00, 0x1c7000, 0x1c7000,
         0x038e00, 0x038e00, 0x1c7000, ];
    static mask5 = [
         0x001e00, 0x001000, 0x001200, 0x001400, 0x001200, 0x001000,
         0x000000, 0x001000, 0x001200, 0x155515, 0x049209, 0x041001,
         0x1fffbf, 0x041000, 0x049200, 0x155400, 0x049200, 0x041000,
         0x1ffe00, 0x041000, 0x049200, ];
    static mask6 = [
         0x001e00, 0x001000, 0x001600, 0x001400, 0x001a00, 0x001c00,
         0x000000, 0x001000, 0x001600, 0x155515, 0x16db2d, 0x071c31,
         0x1fffbf, 0x1c7000, 0x0db600, 0x155400, 0x16da00, 0x071c00,
         0x1ffe00, 0x1c7000, 0x0db600, ];
    static mask7 = [
         0x001400, 0x000e00, 0x001c00, 0x000a00, 0x001000, 0x000200,
         0x000000, 0x000e00, 0x001c00, 0x0aaaaa, 0x1c7187, 0x18e38e,
         0x155515, 0x038e00, 0x071c00, 0x0aaa00, 0x1c7000, 0x18e200,
         0x155400, 0x038e00, 0x071c00, ];
"""

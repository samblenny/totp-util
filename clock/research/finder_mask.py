#!/usr/bin/python3
"""
QR code finder pattern and timing pattern mask generator

This generates a bit mask for the 3 finder patterns in the corners of a version
1 QR code along with the timing patterns connecting the finder patterns. The
Pixels are packed into an array of 21 Uint32's, with the low 21 bits of each
Uint32 representing the pixels (modules) of the QR code. The top left pixel of
the QR code goes in the least significant bit of the first Uint32.
"""

mask = [
    0b111111100000001111111,
    0b100000100000001000001,
    0b101110100000001011101,
    0b101110100000001011101,
    0b101110100000001011101,
    0b100000100000001000001,
    0b111111101010101111111,
    0b000000000000000000000,
    0b000000000000001000000,
    0b000000000000000000000,
    0b000000000000001000000,
    0b000000000000000000000,
    0b000000000000001000000,
    0b000000000000000000000,
    0b000000000000001111111,
    0b000000000000001000001,
    0b000000000000001011101,
    0b000000000000001011101,
    0b000000000000001011101,
    0b000000000000001000001,
    0b000000000000001111111,
    ]

print("    static finderTimingMask = [\n        ", end="")
for i in range(21):
    if i % 6 == 5:
        print(f"0x{mask[i]:06x},\n        ", end="")
    else:
        print(f"0x{mask[i]:06x}, ", end="")
print("];")

"""
    static finderTimingMask = [
        0x1fc07f, 0x104041, 0x17405d, 0x17405d, 0x17405d, 0x104041,
        0x1fd57f, 0x000000, 0x000040, 0x000000, 0x000040, 0x000000,
        0x000040, 0x000000, 0x00007f, 0x000041, 0x00005d, 0x00005d,
        0x00005d, 0x000041, 0x00007f, ];
"""

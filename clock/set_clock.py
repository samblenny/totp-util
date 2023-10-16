#!/usr/bin/python3
"""
Clock setter:

This is meant to be run on an airgapped Linux system, such as a Raspberry Pi,
to set the system time to within +/- 2s of UTC using a 2d barcode scanner to
read an NTP synchronized QR code clock.

Expected QR code timestamp format is MMDDhhmmCCYYss in UTC. This is almost the
format for  setting time with the GNU date command line tool (MMDDhhmmCCYY.ss).
Omitting the "." before seconds makes it easier to encode the full timestamp in
a small QR code.

Usage:
1. ./set_clock.py
2. Scan QR code clock with USB HID 2d barcode scanner (scanner needs to be
   configured to send an Enter key press at the end of the code)
"""
import os
import re
import sys

pattern = re.compile(r'[0-9]{14}')
print("If you proceed, this will run `sudo date --utc ...`.")
print("You can scan the QR code clock now.")
string = input("time?> ")
if pattern.match(string):
    date_cmd = f"sudo date --utc {string[:12]}.{string[12:]}"
    if sys.platform.startswith('linux'):
        print(date_cmd)
        os.system(date_cmd)
    else:
        print(f"On a Linux system, this would have run:\n {date_cmd}")
else:
    print("unexpected input... bye.")

#!/usr/bin/python3
import os,re,sys
r = re.compile(r'[0-9]{14}')
s = input("Scan QR clock> ")
if r.match(s):
  c = f"sudo date --utc {s[:12]}.{s[12:14]}"
  print(c)
  if sys.platform.startswith('linux'):
    os.system(c)
else:
  print("???")

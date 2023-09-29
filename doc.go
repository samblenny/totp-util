/*
Totp_util implements a text-mode utility to assist with TOTP 2FA signups and to
securely manage TOTP profiles for hardware tokens.

Totp_util is built to work on an airgapped workstation (barcode scanner but no
network connection) and to avoid writing plaintext secrets to disk.

Key Features:
  - TOTP QR Code URI parsing for use with USB HID barcode scanners
  - 10 slot organizer to help program multi-profile hardware tokens
  - TOTP profile editor for manual data entry or QR Code URI cleanup
  - Source code is simple so you can read it in a reasonable time

Intended Uses:
  - Sign up for 2FA with online services that expect you to use a QR Code and
    an authenticator app. But, instead of a smartphone, use a USB HID barcode
    scanner, totp_util, and a programmable TOTP hardware token.
  - Generate OTP codes from the command line without leaving unencrypted TOTP
    secrets laying around on your disk.

Limitations:
  - For convenient QR code scanning, you need a USB HID 2D barcode scanner.
  - If you want to use totp_util fully airgapped, you will need to manually set
    your workstation's clock because NTP won't be available.
  - The text-mode menu system uses line-buffered input to reduce code complexity
    and improve auditability. That means arrow key editing doesn't work, which
    is annoying. The tradeoff is simple code with easy to follow control flow.
  - Go's runtime makes it difficult to zeroize buffers that have been used to
    hold key material. Strings are heap allocated and immutable. Using raw-mode
    terminal IO with byte buffers might work, but that would require a lot of
    additional code complexity. If you care about secrets remaining in freed RAM
    pages after totp_util exits, the solution is to power down your computer for
    long enough to clear the RAM.
*/
package main

/*
Totp_util helps you sign up for QR code "authenticator app" 2FA using a USB HID
2D barcode scanner and airgapped workstation.

Key Features:
  - Parse TOTP QR Code URIs (note: this assumes you have a USB barcode scanner)
  - Allow TOTP profile editing for manual data entry or URI cleanup
  - Generate TOTP login codes
  - Source code is short, focused, and hopefully easy to audit
  - Ephemerality: totp_util works out of RAM and does not save any data to disk

Limitations:
  - For convenient QR code scanning, you need a USB HID 2D barcode scanner.
  - If you want to use totp_util fully airgapped, you will need to manually set
    your workstation's clock because NTP won't be available.
  - The text-mode interactive menu system uses line-buffered input to reduce
    code complexity. That means arrow-key editing is not implemented.
  - Go's runtime makes it difficult to sanitize buffers that have been used to
    hold key material. If you care about clearing secrets out of RAM after
    totp_util exits, your best option is to power down your computer.

Notes on Backups and Data Hygiene:
  - Totp_util is built with the assumption that you already use some kind of
    password manager, encrypted backup disk, or physical lockbox, and that you
    will keep a copy of your TOTP QR codes or decoded TOTP URIs in that place.
  - Totp_util uses an interactive prompt rather than command line arguments so
    that TOTP secrets don't get written to your shell history file.
*/
package main

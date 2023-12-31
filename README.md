# totp-util

TOTP setup utility for use with USB 2d barcode scanner


## Usage

To run `totp-util`, do `make run`. You will need a Go compiler.

At the `totp-util` interactive prompt, you can enter TOTP URI's with a 2d
barcode scanner, type commands by hand, or copy and paste URI's from a password
manager. Once a URI is successfully parsed, `totp-util` will begin showing OTP
codes until you tell it to stop.

`totp-util` does not save secrets or other information to disk. From
`totp-util`'s perspective, you are responsible for managing backups of
enrollment QR codes or URI's on your own (password manager, encrypted disk
volume, printouts, or whatever... totally up to you).

The interactive menu looks like this:

```
$ make run
totp-util v0.4.1
 ?             - Show menu
 p             - Print profile
 otpauth://... - Parse TOTP QR Code URI into profile
 secret=<s>    - Set secret to <s> (must be base32 string)
 algorithm=<s> - Set algorithm to <s> (can be empty, "SHA1" or "SHA256")
 digits=<s>    - Set digits to <s> (can be empty, "6", or "8")
 period=<s>    - Set period to <s> (can be empty, "30", or "60")
 clr           - Clear profile
 t             - Show updating TOTP code (press Enter key to stop)
 q             - Quit
> otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example
{
 "issuer": "Example",
 "account": "alice@google.com",
 "secret": "JBSWY3DPEHPK3PXP"
}
To stop displaying TOTP codes, use the Enter key.

(28s)  302134  

> 
```

Also, check out the build targets in [Makefile](Makefile)


## Design

The main design principle for `totp-util` was to keep the implementation as
simple and straightforward as possible. The point is to have fewer moving parts
to break and to have readable code that is reasonable to audit with a realistic
amount of effort.


## Tests

To run the tests, do `make test` or `go test`.


## QR Code Clock

There is also a QR code clock to help set time without NTP on an
airgapped Linux box using a USB 2d barcode scanner. For details,
see [clock/README.md](clock/README.md).

The QR code clock thing is a single static html file, so it works
great offline. But, there is also a copy hosted here at
https://samblenny.github.io/totp-util/clock/

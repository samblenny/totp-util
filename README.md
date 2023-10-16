# totp-util

TOTP hardware token setup utility for use with USB 2d barcode scanner


## Usage

To run `totp_util`, do `make run`

For more detail, check out the build targets in [Makefile](Makefile)


## Tests

To run the tests, do `make test` or `go test`.


## QR Code Clock

There is also a QR code clock to help set time without NTP on an
airgapped Linux box using a USB 2d barcode scanner. For details,
see [clock/README.md](clock/README.md).

The QR code clock thing is a single static html file, so it works
great offline. But, there is also a copy hosted here at
https://samblenny.github.io/totp-util/clock/

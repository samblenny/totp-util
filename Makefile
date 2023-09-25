.PHONY: run test clean

totp-util:
	go build -buildvcs=false -ldflags "-s -w" -trimpath

run: totp-util
	./totp-util

test:
	go test

clean:
	go clean

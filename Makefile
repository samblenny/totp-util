.PHONY: run test clean
SRC_FILES=go.mod main.go profile.go doc.go totp.go

totp-util: Makefile $(SRC_FILES)
	@go build -buildvcs=false -ldflags "-s -w" -trimpath

run: totp-util
	@./totp-util

test:
	go test

clean:
	go clean

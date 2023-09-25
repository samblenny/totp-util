.PHONY: run test clean

totp-util: go.mod main.go
	@go build -buildvcs=false -ldflags "-s -w" -trimpath

run: totp-util
	@./totp-util

test:
	go test

clean:
	go clean

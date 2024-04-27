.PHONY: shooter web

shooter:
	go build ./cmd/shooter

web:
	go run github.com/hajimehoshi/wasmserve@latest ./cmd/shooter

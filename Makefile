.PHONY: shooter run-web

shooter:
	go build ./cmd/shooter

run-web:
	go run github.com/hajimehoshi/wasmserve@latest ./cmd/shooter

build-web:
	env GOOS=js GOARCH=wasm go build -o shooter.wasm ./cmd/shooter

itch.io: build-web
	cp shooter.wasm itch.io
	butler push ./itch.io kazzmir/simple-shooter:html


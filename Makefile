.PHONY: shooter signaling-server run-web windows

shooter:
	go build ./cmd/shooter

signaling-server:
	go build ./cmd/signaling-server

run-web:
	go run github.com/hajimehoshi/wasmserve@latest ./cmd/shooter

build-web:
	env GOOS=js GOARCH=wasm go build -o shooter.wasm ./cmd/shooter

itch.io: build-web
	cp shooter.wasm itch.io
	butler push ./itch.io kazzmir/simple-shooter:html

windows:
	GOOS=windows go build -o shooter-win.exe ./cmd/shooter

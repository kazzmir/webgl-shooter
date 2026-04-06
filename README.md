![build](https://github.com/kazzmir/webgl-shooter/actions/workflows/build.yml/badge.svg)

Simple 2d overhead shootem-up game

Play online: https://kazzmir.itch.io/simple-shooter

Written in golang and ebitten
https://github.com/hajimehoshi/ebiten

![gameplay](./screenshots/shooter-2025-12-22-231354.png)

## Peer connection

The game now includes a **Multiplayer** menu option in both the desktop and wasm builds. Inside that menu you can set **Peer server**, **Peer room**, and **Connect to peer** to establish a WebRTC data channel to another running instance of the game. Gameplay state sync is not wired up yet.

To try it locally:

1. Start the signaling server with `make signaling-server` and run the binary from `./signaling-server` or `go run ./cmd/signaling-server`.
2. Start the desktop game or serve the wasm build with `make run-web` / `make build-web`.
3. In each game instance, open **Multiplayer**, set the same **Peer server** and **Peer room** values, then choose **Connect to peer**.

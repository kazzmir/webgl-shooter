package main

import (
    "log"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
}

func (game *Game) Update() error {
    return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
    ebitenutil.DebugPrint(screen, "debugging")
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return 320, 240
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(640, 480)
    ebiten.SetWindowTitle("Hello!")

    log.Printf("Running")

    err := ebiten.RunGame(&Game{})
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }
}

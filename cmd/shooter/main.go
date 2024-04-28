package main

import (
    "log"
    "os"
    "fmt"

    "image"
    "image/color"
    _ "image/png"

    "github.com/hajimehoshi/ebiten/v2"
    _ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Player struct {
    x, y float64
    velocityX, velocityY float64
    pic *ebiten.Image
}

func (player *Player) Draw(screen *ebiten.Image) {
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(player.x, player.y)
    screen.DrawImage(player.pic, options)
}

func loadPng(path string) (image.Image, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    defer file.Close()

    img, _, err := image.Decode(file)
    return img, err
}

func MakePlayer(x, y float64) (*Player, error) {

    playerImage, err := loadPng("images/player/player.png")

    if err != nil {
        return nil, err
    }

    return &Player{x: x, y: y, pic: ebiten.NewImageFromImage(playerImage)}, nil
}

type Game struct {
    Player *Player
}

func (game *Game) Update() error {

    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)
    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            game.Player.y -= 2
        } else if key == ebiten.KeyArrowDown {
            game.Player.y += 2
        } else if key == ebiten.KeyArrowLeft {
            game.Player.x -= 2
        } else if key == ebiten.KeyArrowRight {
            game.Player.x += 2
        // FIXME: make ebiten understand key mapping
        } else if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return fmt.Errorf("quit")
        }
    }

    return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
    // ebitenutil.DebugPrint(screen, "debugging")
    game.Player.Draw(screen)
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return 320, 240
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(640, 480)
    ebiten.SetWindowTitle("Hello!")

    log.Printf("Loading player")

    player, err := MakePlayer(50, 50)
    if err != nil {
        log.Printf("Failed to make player: %v", err)
        return
    }

    log.Printf("Running")

    err = ebiten.RunGame(&Game{
        Player: player,
    })
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }
}

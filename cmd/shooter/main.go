package main

import (
    "log"
    "os"
    "fmt"
    "math/rand"

    "image"
    "image/color"
    _ "image/png"

    "github.com/hajimehoshi/ebiten/v2"
    _ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 800
const ScreenHeight = 600

type Bullet struct {
    x, y float64
    velocityX, velocityY float64
    pic *ebiten.Image
}

func (bullet *Bullet) Draw(screen *ebiten.Image) {
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(bullet.x, bullet.y)
    screen.DrawImage(bullet.pic, options)
}

func (bullet *Bullet) Move(){
    bullet.x += bullet.velocityX
    bullet.y += bullet.velocityY
}

func (bullet *Bullet) IsAlive() bool {
    return bullet.y > 0
}

type StarPosition struct {
    x, y float64
    dx, dy float64
}

type Background struct {
    Star *ebiten.Image
    Stars []*StarPosition
}

func randomFloat(min float64, max float64) float64 {
    return min + rand.Float64() * (max - min)
}

func MakeBackground() (*Background, error) {
    starImage, err := loadPng("images/background/star1.png")
    if err != nil {
        return nil, err
    }

    stars := make([]*StarPosition, 0)
    for i := 0; i < 50; i++ {
        x := randomFloat(0, float64(ScreenWidth))
        y := randomFloat(0, float64(ScreenHeight))
        dx := 0.0
        dy := randomFloat(0.6, 1.1)
        stars = append(stars, &StarPosition{x: x, y: y, dx: dx, dy: dy})
    }

    return &Background{
        Star: ebiten.NewImageFromImage(starImage),
        Stars: stars,
    }, nil
}

func (background *Background) Update(){
    for _, star := range background.Stars {
        star.y += star.dy
        if star.y > ScreenHeight + 50 {
            star.y = -50
        }
    }
}

func (background *Background) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x27, 0x38, 0x3e, 0xff})

    for _, star := range background.Stars {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(star.x, star.y)
        screen.DrawImage(background.Star, options)
    }
}

type Player struct {
    x, y float64
    velocityX, velocityY float64
    bulletCounter int
    pic *ebiten.Image
    bullet *ebiten.Image
}

func (player *Player) Move() {
    player.x += player.velocityX
    player.y += player.velocityY

    accel := 0.23

    if player.velocityX < -accel {
        player.velocityX += accel
    } else if player.velocityX > accel {
        player.velocityX -= accel
    } else {
        player.velocityX = 0
    }

    if player.velocityY < -accel {
        player.velocityY += accel
    } else if player.velocityY > accel {
        player.velocityY -= accel
    } else {
        player.velocityY = 0
    }

    if player.x < 0 {
        player.x = 0
    } else if player.x > ScreenWidth {
        player.x = ScreenWidth
    }

    if player.y < 0 {
        player.y = 0
    } else if player.y > ScreenHeight {
        player.y = ScreenHeight
    }

    if player.bulletCounter > 0 {
        player.bulletCounter -= 1
    }
}

func (player *Player) MakeBullet() *Bullet {

    velocityY := player.velocityY-2
    if velocityY > -1 {
        velocityY = -1
    }

    velocityY = -2.5

    return &Bullet{
        x: player.x + 27,
        y: player.y,
        velocityX: 0,
        velocityY: velocityY,
        pic: player.bullet,
    }
}

func (player *Player) Draw(screen *ebiten.Image) {
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(player.x, player.y)
    screen.DrawImage(player.pic, options)
}

func (player *Player) HandleKeys(game *Game) error {
    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)
    playerAccel := 3.8
    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            game.Player.velocityY = -playerAccel;
        } else if key == ebiten.KeyArrowDown {
            game.Player.velocityY = playerAccel;
        } else if key == ebiten.KeyArrowLeft {
            game.Player.velocityX = -playerAccel;
        } else if key == ebiten.KeyArrowRight {
            game.Player.velocityX = playerAccel;
        // FIXME: make ebiten understand key mapping
        } else if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return fmt.Errorf("quit")
        } else if key == ebiten.KeySpace && game.Player.bulletCounter == 0{
            game.Bullets = append(game.Bullets, game.Player.MakeBullet())
            game.Player.bulletCounter = 5
        }
    }

    return nil
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

    bulletImage, err := loadPng("images/player/bullet.png")

    return &Player{x: x, y: y, pic: ebiten.NewImageFromImage(playerImage), bullet: ebiten.NewImageFromImage(bulletImage)}, nil
}

type Game struct {
    Player *Player
    Background *Background
    Bullets []*Bullet
}

func (game *Game) Update() error {

    game.Background.Update()

    err := game.Player.HandleKeys(game)
    if err != nil {
        return err
    }

    game.Player.Move()

    for i := 0; i < 2; i++ {
        var outBullets []*Bullet
        for _, bullet := range game.Bullets {
            bullet.Move()
            if bullet.IsAlive() {
                outBullets = append(outBullets, bullet)
            }
        }
        game.Bullets = outBullets
    }

    return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
    game.Background.Draw(screen)
    // ebitenutil.DebugPrint(screen, "debugging")
    game.Player.Draw(screen)

    for _, bullet := range game.Bullets {
        bullet.Draw(screen)
    }
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(1024, 768)
    ebiten.SetWindowTitle("Hello!")

    log.Printf("Loading player")

    player, err := MakePlayer(320, 400)
    if err != nil {
        log.Printf("Failed to make player: %v", err)
        return
    }

    background, err := MakeBackground()
    if err != nil {
        log.Printf("Failed to make background: %v", err)
        return
    }

    log.Printf("Running")

    err = ebiten.RunGame(&Game{
        Background: background,
        Player: player,
    })
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }
}

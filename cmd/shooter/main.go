package main

import (
    "log"
    "fmt"
    "math/rand"
    "math"

    "image/color"
    _ "image/png"

    gameImages "github.com/kazzmir/webgl-shooter/images"
    fontLib "github.com/kazzmir/webgl-shooter/font"

    "github.com/hajimehoshi/ebiten/v2"
    _ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

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
    Image *ebiten.Image
}

type Background struct {
    // Star *ebiten.Image
    // Star2 *ebiten.Image
    Stars []*StarPosition
}

func randomFloat(min float64, max float64) float64 {
    return min + rand.Float64() * (max - min)
}

func MakeBackground() (*Background, error) {
    starImage, err := gameImages.LoadImage(gameImages.ImageStar1)
    if err != nil {
        return nil, err
    }

    starImage2, err := gameImages.LoadImage(gameImages.ImageStar2)
    if err != nil {
        return nil, err
    }

    planet1, err := gameImages.LoadImage(gameImages.ImagePlanet)
    if err != nil {
        return nil, err
    }

    images := []*ebiten.Image{
        ebiten.NewImageFromImage(starImage),
        ebiten.NewImageFromImage(starImage2),
        ebiten.NewImageFromImage(planet1),
    }

    stars := make([]*StarPosition, 0)
    for i := 0; i < 50; i++ {
        x := randomFloat(0, float64(ScreenWidth))
        y := randomFloat(0, float64(ScreenHeight))
        dx := 0.0
        dy := randomFloat(0.6, 1.1)

        image := images[rand.Intn(len(images))]

        stars = append(stars, &StarPosition{x: x, y: y, dx: dx, dy: dy, Image: image})
    }

    return &Background{
        // Star: ebiten.NewImageFromImage(starImage),
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
    screen.Fill(color.RGBA{0x1b, 0x22, 0x24, 0xff})

    for _, star := range background.Stars {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(star.x, star.y)
        screen.DrawImage(star.Image, options)
    }
}

type ShaderManager struct {
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
}

func MakeShaderManager() (*ShaderManager, error) {
    redShader, err := LoadRedShader()
    if err != nil {
        return nil, err
    }

    shadowShader, err := LoadShadowShader()
    if err != nil {
        return nil, err
    }

    return &ShaderManager{
        RedShader: redShader,
        ShadowShader: shadowShader,
    }, nil
}

type Enemy interface {
    Move()
    Draw(screen *ebiten.Image, shaders *ShaderManager)
}

type NormalEnemy struct {
    x, y float64
    velocityX, velocityY float64
    Life float64
    pic *ebiten.Image
    Flip bool
}

func (enemy *NormalEnemy) Move() {
    enemy.x += enemy.velocityX
    enemy.y += enemy.velocityY

    if enemy.y > ScreenHeight + 50 {
        enemy.y = -100
    }
}

func (enemy *NormalEnemy) Draw(screen *ebiten.Image, shaders *ShaderManager) {

    // draw shadow
    shaderOptions := &ebiten.DrawRectShaderOptions{}
    if enemy.Flip {
        shaderOptions.GeoM.Rotate(math.Pi)
    }
    shaderOptions.GeoM.Translate(enemy.x, enemy.y + 10)
    shaderOptions.Blend = AlphaBlender
    shaderOptions.Images[0] = enemy.pic
    bounds := enemy.pic.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.ShadowShader, shaderOptions)

    options := &ebiten.DrawImageOptions{}
    // flip 180 degrees
    if enemy.Flip {
        options.GeoM.Rotate(math.Pi)
    }
    options.GeoM.Translate(enemy.x, enemy.y)
    screen.DrawImage(enemy.pic, options)
}

func MakeEnemy1(x, y float64) (Enemy, error) {
    enemyImage, err := gameImages.LoadImage(gameImages.ImageEnemy1)
    if err != nil {
        return nil, err
    }

    return &NormalEnemy{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 2,
        Life: 10,
        pic: ebiten.NewImageFromImage(enemyImage),
        Flip: true,
    }, nil
}

func MakeEnemy2(x, y float64) (Enemy, error) {
    enemyImage, err := gameImages.LoadImage(gameImages.ImageEnemy2)
    if err != nil {
        return nil, err
    }

    return &NormalEnemy{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 2,
        Life: 10,
        pic: ebiten.NewImageFromImage(enemyImage),
        Flip: false,
    }, nil
}

type Player struct {
    x, y float64
    Jump int
    velocityX, velocityY float64
    bulletCounter int
    pic *ebiten.Image
    bullet *ebiten.Image
    Score int
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
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

var AlphaBlender ebiten.Blend = ebiten.Blend{
    BlendFactorSourceRGB:        ebiten.BlendFactorSourceAlpha,
    BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
    BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
    BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
    BlendOperationRGB:           ebiten.BlendOperationAdd,
    BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

func (player *Player) Draw(screen *ebiten.Image, shaders *ShaderManager, font *text.GoTextFaceSource) {
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Score: %v", player.Score), &text.GoTextFace{Source: font, Size: 15}, op)

    options := &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(player.x + player.velocityX * 3, player.y + 10)
    options.Blend = AlphaBlender
    options.Images[0] = player.pic
    bounds := player.pic.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.ShadowShader, options)

    /*
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(player.x, player.y)
    screen.DrawImage(player.pic, options)
    */

    if player.Jump > 0 {
        options := &ebiten.DrawRectShaderOptions{}
        options.GeoM.Translate(player.x, player.y)
        options.Blend = AlphaBlender
        options.Images[0] = player.pic
        options.Uniforms = make(map[string]interface{})
        var radians float32 = math.Pi * float32(player.Jump) * 360 / JumpDuration / 180.0
        // radians = math.Pi * 90 / 180
        // log.Printf("Red: %v", radians)
        options.Uniforms["Red"] = radians
        bounds := player.pic.Bounds()
        screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, options)
    } else {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(player.x, player.y)
        screen.DrawImage(player.pic, options)
    }
}

func (player *Player) HandleKeys(game *Game) error {
    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)
    playerAccel := 3.8
    if player.Jump > 0 {
        playerAccel = 5
    }
    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            player.velocityY = -playerAccel;
        } else if key == ebiten.KeyArrowDown {
            player.velocityY = playerAccel;
        } else if key == ebiten.KeyArrowLeft {
            player.velocityX = -playerAccel;
        } else if key == ebiten.KeyArrowRight {
            player.velocityX = playerAccel;
        } else if key == ebiten.KeyShift && player.Jump <= -50 {
            player.Jump = JumpDuration
        // FIXME: make ebiten understand key mapping
        } else if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        } else if key == ebiten.KeySpace && game.Player.bulletCounter == 0{
            game.Bullets = append(game.Bullets, game.Player.MakeBullet())
            player.bulletCounter = 5
        }
    }

    if player.Jump > -50 {
        player.Jump -= 1
    }

    return nil
}

/*
func loadPng(path string) (image.Image, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    defer file.Close()

    img, _, err := image.Decode(file)
    return img, err
}
*/

const JumpDuration = 50

func MakePlayer(x, y float64) (*Player, error) {

    playerImage, err := gameImages.LoadImage(gameImages.ImagePlayer)

    if err != nil {
        return nil, err
    }

    bulletImage, err := gameImages.LoadImage(gameImages.ImageBullet)
    if err != nil {
        return nil, err
    }

    return &Player{
        x: x,
        y: y,
        pic: ebiten.NewImageFromImage(playerImage),
        bullet: ebiten.NewImageFromImage(bulletImage),
        Jump: -50,
        Score: 0,
    }, nil
}

type Game struct {
    Player *Player
    Background *Background
    Bullets []*Bullet
    Font *text.GoTextFaceSource
    Enemies []Enemy
    ShaderManager *ShaderManager
}

func (game *Game) MakeEnemy() error {
    var enemy Enemy
    var err error

    switch rand.Intn(2) {
        case 0:
            enemy, err = MakeEnemy1(randomFloat(50, ScreenWidth - 50), randomFloat(-500, -50))
        case 1:
            enemy, err = MakeEnemy2(randomFloat(50, ScreenWidth - 50), randomFloat(-500, -50))
    }

    if err != nil {
        return err
    }

    game.Enemies = append(game.Enemies, enemy)

    return nil
}

func (game *Game) Update() error {

    game.Background.Update()

    err := game.Player.HandleKeys(game)
    if err != nil {
        return err
    }

    game.Player.Move()

    for _, enemy := range game.Enemies {
        enemy.Move()
    }

    for i := 0; i < 3; i++ {
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

    for _, enemy := range game.Enemies {
        enemy.Draw(screen, game.ShaderManager)
    }

    // ebitenutil.DebugPrint(screen, "debugging")
    game.Player.Draw(screen, game.ShaderManager, game.Font)

    for _, bullet := range game.Bullets {
        bullet.Draw(screen)
    }

    /*
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, "Hello, World!", &text.GoTextFace{Source: game.Font, Size: 15}, op)
    */
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("Shooter")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

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

    font, err := fontLib.LoadFont()
    if err != nil {
        log.Printf("Failed to load font: %v", err)
        return
    }

    shaderManager, err := MakeShaderManager()

    game := Game{
        Background: background,
        Player: player,
        Font: font,
        ShaderManager: shaderManager,
    }

    for i := 0; i < 5; i++ {
        err = game.MakeEnemy()
        if err != nil {
            log.Printf("Failed to make enemy: %v", err)
            return
        }
    }

    log.Printf("Running")
    err = ebiten.RunGame(&game)
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }

    log.Printf("Bye!")
}

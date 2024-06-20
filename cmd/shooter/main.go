package main

import (
    "log"
    "fmt"
    "io"
    "time"
    "os"
    "bytes"
    "errors"
    "math/rand"
    "math"
    "strconv"
    "sync"
    "sync/atomic"
    "context"
    "runtime/pprof"

    "image"
    "image/color"
    _ "image/png"

    gameImages "github.com/kazzmir/webgl-shooter/images"
    fontLib "github.com/kazzmir/webgl-shooter/font"
    audioFiles "github.com/kazzmir/webgl-shooter/audio"
    blurLib "github.com/kazzmir/webgl-shooter/lib/blur"

    "github.com/hajimehoshi/ebiten/v2"
    _ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/audio"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const debugForceBoss = false

const ScreenWidth = 1024
const ScreenHeight = 768

const EnergyIncreasePerFrame = 0.4

func onScreen(x float64, y float64, margin float64) bool {
    return x > -margin && x < ScreenWidth + margin && y > -margin && y < ScreenHeight + margin
}

func toFloatArray(color color.Color) []float32 {
    r, g, b, a := color.RGBA()
    var max float32 = 65535.0
    return []float32{float32(r) / max, float32(g) / max, float32(b) / max, float32(a) / max}
}

func drawCenteredImage(screen *ebiten.Image, pic *ebiten.Image, x float64, y float64){
    x1 := x - float64(pic.Bounds().Dx()) / 2
    y1 := y - float64(pic.Bounds().Dy()) / 2
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(x1, y1)
    screen.DrawImage(pic, options)
}

// generates a bunch of colors between start and end, interpolating linerally
func linearGradient(start color.Color, end color.Color, steps uint32) chan color.RGBA {
    out := make(chan color.RGBA)

    go func(){
        var max int32 = 255
        var i int32

        r1, g1, b1, a1 := start.RGBA()
        r2, g2, b2, a2 := end.RGBA()

        for i = 0; i < int32(steps); i++ {
            r := uint8((int32(r1) + (int32(r2) - int32(r1)) * i / int32(steps)) / max)
            g := uint8((int32(g1) + (int32(g2) - int32(g1)) * i / int32(steps)) / max)
            b := uint8((int32(b1) + (int32(b2) - int32(b1)) * i / int32(steps)) / max)
            a := uint8((int32(a1) + (int32(a2) - int32(a1)) * i / int32(steps)) / max)

            out <- color.RGBA{R: r, G: g, B: b, A: a}
        }
        close(out)
    }()

    return out
}

func createLinearRectangle(width uint32, height uint32, start color.Color, end color.Color) image.Image {
    out := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

    gradient := linearGradient(start, end, height)

    for y := 0; y < int(height); y++ {
        current := premultiplyAlpha(<-gradient)

        for x := 0; x < out.Bounds().Dx(); x++ {
            out.Set(x, out.Bounds().Dy() - y - 1, current)
        }
    }

    return out
}

type Bullet struct {
    x, y float64
    Strength float64
    velocityX, velocityY float64
    pic *ebiten.Image
    animation *Animation
    health int
}

func (bullet *Bullet) Damage(amount int) {
    bullet.health -= amount
}

func (bullet *Bullet) Draw(screen *ebiten.Image) {

    if bullet.animation != nil {
        bullet.animation.Draw(screen, bullet.x, bullet.y)
    } else if bullet.pic != nil {
        drawCenteredImage(screen, bullet.pic, bullet.x, bullet.y)
    }
}

func (bullet *Bullet) Move(){
    bullet.x += bullet.velocityX
    bullet.y += bullet.velocityY

    if bullet.animation != nil {
        bullet.animation.Update()
    }
}

func (bullet *Bullet) IsAlive() bool {
    return bullet.health > 0 && onScreen(bullet.x, bullet.y, 10)
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
    EdgeShader *ebiten.Shader
    ExplosionShader *ebiten.Shader
    AlphaCircleShader *ebiten.Shader
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

    edgeShader, err := LoadEdgeShader()
    if err != nil {
        return nil, err
    }

    explosionShader, err := LoadExplosionShader()
    if err != nil {
        return nil, err
    }

    alphaCircleShader, err := LoadAlphaCircleShader()
    if err != nil {
        return nil, err
    }

    return &ShaderManager{
        RedShader: redShader,
        ShadowShader: shadowShader,
        EdgeShader: edgeShader,
        ExplosionShader: explosionShader,
        AlphaCircleShader: alphaCircleShader,
    }, nil
}

const BombDelay = 60

type Player struct {
    x, y float64
    Jump int
    velocityX, velocityY float64
    rawImage image.Image
    pic *ebiten.Image
    Guns []Gun
    GunEnergy float64
    MaxEnergy float64
    Health float64
    MaxHealth float64
    Score uint64
    Kills uint64
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
    Counter int
    SoundShoot chan bool
    Bombs int
    BombCounter int

    PowerupEnergy int
}

func (player *Player) Damage(amount float64) {
    player.Health -= amount
    if player.Health < 0 {
        player.Health = 0
    }
}

func (player *Player) IsAlive() bool {
    return player.Health > 0
}

func (player *Player) Bounds() image.Rectangle {
    bounds := player.rawImage.Bounds()

    x1 := player.x - float64(bounds.Dx()) / 2
    y1 := player.y - float64(bounds.Dy()) / 2
    x2 := x1 + float64(bounds.Dx())
    y2 := y1 + float64(bounds.Dy())

    return image.Rect(int(x1), int(y1), int(x2), int(y2))
}

func (player *Player) Collide(x float64, y float64) bool {
    bounds := player.Bounds()
    if int(x) >= bounds.Min.X && int(x) <= bounds.Max.X && int(y) >= bounds.Min.Y && int(y) <= bounds.Max.Y {
        cx := int(x) - bounds.Min.X
        cy := int(y) - bounds.Min.Y
        c := player.rawImage.At(cx, cy)
        _, _, _, a := c.RGBA()
        if a > 200 * 255 {
            return true
        }
    }

    return false
}

func sameType(a interface{}, b interface{}) bool {
    return fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b)
}

func haveGun(guns []Gun, gun Gun) bool {
    for _, g := range guns {
        if sameType(g, gun){
            return true
        }
    }

    return false
}

func (player *Player) EnableNextGun(){
    guns := []Gun{&DualBasicGun{enabled: true}, &BeamGun{enabled: true}, &MissleGun{enabled: true}}
    for _, gun := range guns {
        if !haveGun(player.Guns, gun) {
            player.Guns = append(player.Guns, gun)
            return
        }
    }
}

func (player *Player) Move() {
    player.Counter += 1

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

    player.GunEnergy += EnergyIncreasePerFrame
    if player.GunEnergy > player.MaxEnergy {
        player.GunEnergy = player.MaxEnergy
    }

    for _, gun := range player.Guns {
        gun.Update()
    }

    if player.BombCounter > 0 {
        player.BombCounter -= 1
    }

    if player.PowerupEnergy > 0 {
        player.PowerupEnergy -= 1
    }
}

func (player *Player) Shoot(imageManager *ImageManager, soundManager *SoundManager) []*Bullet {

    var bullets []*Bullet

    for _, gun := range player.Guns {
        if gun.IsEnabled() && (player.PowerupEnergy > 0 || gun.EnergyUsed() <= player.GunEnergy) {
            more, err := gun.Shoot(imageManager, player.x, player.y - float64(player.pic.Bounds().Dy()) / 2)
            if err != nil {
                log.Printf("Could not create bullets: %v", err)
            } else {
                if more != nil {
                    if player.PowerupEnergy == 0 {
                        player.GunEnergy -= gun.EnergyUsed()
                    }
                    bullets = append(bullets, more...)

                    select {
                        case <-player.SoundShoot:
                            // soundManager.Play(audioFiles.AudioShoot1)
                            gun.DoSound(soundManager)
                            go func(){
                                time.Sleep(10 * time.Millisecond)
                                player.SoundShoot <- true
                            }()
                        default:
                    }
                }
            }
        }
    }

    return bullets
}

var AlphaBlender ebiten.Blend = ebiten.Blend{
    BlendFactorSourceRGB:        ebiten.BlendFactorSourceAlpha,
    BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
    BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
    BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
    BlendOperationRGB:           ebiten.BlendOperationAdd,
    BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

func (player *Player) Draw(screen *ebiten.Image, shaders *ShaderManager, imageManager *ImageManager, font *text.GoTextFaceSource) {
    face := &text.GoTextFace{Source: font, Size: 15} 

    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Score: %v", player.Score), face, op)

    op.GeoM.Translate(1, 20)
    text.Draw(screen, fmt.Sprintf("Kills: %v", player.Kills), face, op)

    op.GeoM.Translate(1, 40)
    if player.PowerupEnergy > 0 {
        text.Draw(screen, fmt.Sprintf("Energy: MAX"), face, op)
    } else {
        text.Draw(screen, fmt.Sprintf("Energy: %.2f", player.GunEnergy), face, op)
    }

    playerX := player.x - float64(player.pic.Bounds().Dx()) / 2
    playerY := player.y - float64(player.pic.Bounds().Dy()) / 2

    options := &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(playerX + player.velocityX * 3, playerY + 10)
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
        options.GeoM.Translate(playerX, playerY)
        options.Blend = AlphaBlender
        options.Images[0] = player.pic
        options.Uniforms = make(map[string]interface{})
        var radians float64 = math.Pi * float64(player.Jump) * 360 / JumpDuration / 180.0
        // radians = math.Pi * 90 / 180
        // log.Printf("Red: %v", radians)
        // red := vec4(abs(sin(Red) / 3), 0, 0, 0)
        options.Uniforms["Red"] = toFloatArray(color.RGBA{R: uint8(math.Abs(math.Sin(radians) / 3) * 255), G: 0, B: 0, A: 0})
        bounds := player.pic.Bounds()
        screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, options)
    } else {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(playerX, playerY)
        screen.DrawImage(player.pic, options)
    }

    if player.PowerupEnergy > 0 {
        options = &ebiten.DrawRectShaderOptions{}
        options.GeoM.Translate(playerX, playerY)
        options.Blend = AlphaBlender
        options.Images[0] = player.pic
        options.Uniforms = make(map[string]interface{})
        // options.Uniforms["Color"] = []float32{0, 0, float32((math.Sin(float64(player.Counter) * 7 * math.Pi / 180.0) + 1) / 2), 1}
        alpha := float32((math.Sin(float64(player.PowerupEnergy) * 7 * math.Pi / 180.0) + 1) / 2)
        r, g, b, _ := PowerupColor.RGBA()
        useColor := color.RGBA{R: uint8(r / 255), G: uint8(g / 255), B: uint8(b / 255), A: uint8(255.0 * alpha)}
        options.Uniforms["Color"] = toFloatArray(useColor)
        // options.Uniforms["Color"] = []float32{0, 0, 1, 1}
        screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.EdgeShader, options)
    }

    gunFace := &text.GoTextFace{Source: font, Size: 10}

    var iconX float64 = 150
    var iconY float64 = 3
    for i, gun := range player.Guns {
        gun.DrawIcon(screen, imageManager, iconX, iconY)

        op := &text.DrawOptions{}
        op.GeoM.Translate(iconX + 2, iconY + 22)
        var color_ color.RGBA = color.RGBA{0xff, 0xff, 0xff, 0xff}
        if !gun.IsEnabled() {
            color_ = color.RGBA{0xff, 0, 0, 0xff}
        }
        op.ColorScale.ScaleWithColor(color_)
        text.Draw(screen, strconv.Itoa(i+1), gunFace, op)

        iconX += 30
    }

    ShowBombsHud(screen, imageManager, iconX, iconY, player.Bombs)

    energy, _, err := imageManager.LoadImage(gameImages.ImageEnergyBar)
    if err != nil {
        log.Printf("Could not load energy image: %v", err)
    } else {
        if player.PowerupEnergy > 0 {
            vector.DrawFilledRect(screen, 5, 100, float32(energy.Bounds().Dx()), float32(energy.Bounds().Dy()), PowerupColor, true)
        } else {
            options := &ebiten.DrawImageOptions{}
            useHeight := int(player.GunEnergy / player.MaxEnergy * float64(energy.Bounds().Dy()))

            options.GeoM.Translate(5, 100 + float64(energy.Bounds().Dy()) - float64(useHeight))

            vector.StrokeRect(screen, 5, 100, float32(energy.Bounds().Dx()), float32(energy.Bounds().Dy()), 1, premultiplyAlpha(color.RGBA{R: 0xaa, G: 0xe9, B: 0xfb, A: 200}), true)

            sub := energy.SubImage(image.Rect(0, energy.Bounds().Dy() - int(useHeight), energy.Bounds().Dx(), energy.Bounds().Dy())).(*ebiten.Image)
            screen.DrawImage(sub, options)
        }
    }

    health, _, err := imageManager.LoadImage(gameImages.ImageHealthBar)
    if err != nil {
        log.Printf("Could not load health image: %v", err)
    } else {
        options := &ebiten.DrawImageOptions{}
        useHeight := int(player.Health / player.MaxHealth * float64(health.Bounds().Dy()))

        yVal := 400.0

        options.GeoM.Translate(5, yVal + float64(health.Bounds().Dy()) - float64(useHeight))

        vector.StrokeRect(screen, 5, float32(yVal), float32(health.Bounds().Dx()), float32(health.Bounds().Dy()), 1, premultiplyAlpha(color.RGBA{R: 0xaa, G: 0xe9, B: 0xfb, A: 200}), true)

        sub := health.SubImage(image.Rect(0, health.Bounds().Dy() - int(useHeight), health.Bounds().Dx(), health.Bounds().Dy())).(*ebiten.Image)
        screen.DrawImage(sub, options)
    }
}

func enableGun(guns []Gun, index int) {
    if index < len(guns) {
        gun := guns[index]
        gun.SetEnabled(!gun.IsEnabled())
    }
}

func (player *Player) HandleKeys(game *Game, run *Run) error {
    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)

    maxVelocity := 3.8

    playerAccel := 0.9
    if player.Jump > 0 {
        playerAccel = 3
        maxVelocity = 5.5
    }

    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            player.velocityY -= playerAccel
        } else if key == ebiten.KeyArrowDown {
            player.velocityY += playerAccel
        } else if key == ebiten.KeyArrowLeft {
            player.velocityX -= playerAccel
        } else if key == ebiten.KeyArrowRight {
            player.velocityX += playerAccel
        } else if key == ebiten.KeyShift && player.Jump <= -50 {
            player.Jump = JumpDuration
        } else if key == ebiten.KeyB && player.BombCounter == 0 && player.Bombs > 0 {
            game.Bombs = append(game.Bombs, MakeBomb(player.x, player.y - 20, 0, -1.8))
            player.Bombs -= 1
            player.BombCounter = BombDelay
        } else if key == ebiten.KeySpace {
            game.Bullets = append(game.Bullets, game.Player.Shoot(game.ImageManager, game.SoundManager)...)
        }
    }

    player.velocityX = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityX))
    player.velocityY = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityY))

    moreKeys := make([]ebiten.Key, 0)
    moreKeys = inpututil.AppendJustPressedKeys(moreKeys)
    for _, key := range moreKeys {
        // FIXME: make ebiten understand key mapping
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            // return ebiten.Termination
            run.Mode = RunMenu
        } else if key == ebiten.KeyDigit1 {
            enableGun(player.Guns, 0)
        } else if key == ebiten.KeyDigit2 {
            enableGun(player.Guns, 1)
        } else if key == ebiten.KeyDigit3 {
            enableGun(player.Guns, 2)
        } else if key == ebiten.KeyDigit4 {
            enableGun(player.Guns, 3)
        }
    }

    if player.Jump > -50 {
        player.Jump -= 1
    }

    return nil
}

const JumpDuration = 50

func MakePlayer(x, y float64) (*Player, error) {

    playerImage, err := gameImages.LoadImage(gameImages.ImagePlayer)

    if err != nil {
        return nil, err
    }

    soundChan := make(chan bool, 2)
    soundChan <- true

    return &Player{
        x: x,
        y: y,
        rawImage: playerImage,
        pic: ebiten.NewImageFromImage(playerImage),
        // Gun: &BasicGun{},
        // Gun: &DualBasicGun{},
        GunEnergy: 100.0,
        MaxEnergy: 100.0,
        Health: 100.0,
        MaxHealth: 100.0,
        Bombs: 0,
        Guns: []Gun{
            &BasicGun{enabled: true},
            // &DualBasicGun{enabled: false},
            // &BeamGun{enabled: false},
            // &MissleGun{enabled: false},
        },
        // Gun: &BeamGun{},
        Jump: -50,
        Score: 0,
        SoundShoot: soundChan,
    }, nil
}

type ImagePair struct {
    Image *ebiten.Image
    Raw image.Image
}

type ImageManager struct {
    Images map[gameImages.Image]ImagePair
}

func MakeImageManager() *ImageManager {
    return &ImageManager{
        Images: make(map[gameImages.Image]ImagePair),
    }
}

func (manager *ImageManager) CreateEnergyImage() image.Image {
    return createLinearRectangle(15, 250, color.RGBA{R: 0, G: 0, B: 5, A: 210}, color.RGBA{R: 0, G: 0, B: 255, A: 210})
}

func (manager *ImageManager) CreateHealthImage() image.Image {
    // #f00f26
    // #e9f366
    return createLinearRectangle(15, 250, color.RGBA{R: 0xf0, G: 0x0f, B: 0x26, A: 210}, color.RGBA{R: 0xe9, G: 0xf3, B: 0x66, A: 210})
}

func (manager *ImageManager) BlurImage(name gameImages.Image, factor float64, blur int, blurColor color.Color) (*ebiten.Image, error) {

    blurName := name + "-blur"

    if image, ok := manager.Images[blurName]; ok {
        return image.Image, nil
    }

    _, raw, err := manager.LoadImage(name)
    if err != nil {
        return nil, err
    }

    blurred := blurLib.MakeBlur(raw, factor, blur, blurColor)
    converted := ebiten.NewImageFromImage(blurred)
    manager.Images[blurName] = ImagePair{
        Image: converted,
        Raw: blurred,
    }

    return converted, nil
}

func (manager *ImageManager) LoadImage(name gameImages.Image) (*ebiten.Image, image.Image, error) {
    if image, ok := manager.Images[name]; ok {
        return image.Image, image.Raw, nil
    }

    if name == gameImages.ImageEnergyBar {
        raw := manager.CreateEnergyImage()
        converted := ebiten.NewImageFromImage(raw)
        manager.Images[name] = ImagePair{
            Image: converted,
            Raw: raw,
        }
        return converted, raw, nil
    }

    if name == gameImages.ImageHealthBar {
        raw := manager.CreateHealthImage()
        converted := ebiten.NewImageFromImage(raw)
        manager.Images[name] = ImagePair{
            Image: converted,
            Raw: raw,
        }
        return converted, raw, nil
    }

    loaded, err := gameImages.LoadImage(name)
    if err != nil {
        return nil, nil, err
    }

    converted := ebiten.NewImageFromImage(loaded)

    manager.Images[name] = ImagePair{
        Image: converted,
        Raw: loaded,
    }

    return converted, loaded, nil
}

func (manager *ImageManager) LoadAnimation(name gameImages.Image) (*Animation, error) {
    loaded, _, err := manager.LoadImage(name)
    if err != nil {
        return nil, err
    }

    switch name {
        case gameImages.ImageExplosion2: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageExplosion3: return NewAnimation(loaded, 4, 5, 0.7, false), nil
        case gameImages.ImageHit: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageHit2: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageBeam1:
            return NewAnimationCoordinates(loaded, 2, 3, 0.13, []SheetCoordinate{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}, {0, 1}, {2, 0}, {1, 0}}, true), nil
        case gameImages.ImageWave1:
            return NewAnimationCoordinates(loaded, 1, 3, 0.10, []SheetCoordinate{{0, 0}, {1, 0}, {2, 0}, {1, 0}}, true), nil
        case gameImages.ImageRotate1: return NewAnimation(loaded, 2, 2, 0.14, true), nil
        case gameImages.ImageFire1: return NewAnimation(loaded, 5, 6, 0.7, true), nil
    }

    return nil, fmt.Errorf("No such animation %v", name)
}

type SoundHandler struct {
    Make func() *audio.Player
    MakeLoop func() (*audio.Player, error)
    // Players chan *audio.Player
}

type SoundManager struct {
    Sounds map[audioFiles.AudioName]*SoundHandler
    Context *audio.Context
    SampleRate int
    Quit context.Context
    Volume float64
}

func (manager *SoundManager) SetVolume(volume float64){
    manager.Volume = volume
}

func (manager *SoundManager) GetVolume() float64 {
    return manager.Volume
}

func MakeSoundManager(quit context.Context, audioContext *audio.Context, volume float64) (*SoundManager, error) {
    manager := SoundManager{
        Sounds: make(map[audioFiles.AudioName]*SoundHandler),
        SampleRate: 48000,
        Context: audioContext,
        Quit: quit,
        Volume: volume,
    }

    return &manager, manager.LoadAll()
}

func MakeSoundHandler(name audioFiles.AudioName, context *audio.Context, sampleRate int) (*SoundHandler, error) {
    var data []byte

    var create sync.Once

    load := func(){
        log.Printf("Creating sound %v", name)
        stream, err := audioFiles.LoadSound(name, sampleRate)
        if err != nil {
            log.Printf("Error loading sound %v: %v", name, err)
            return
        }

        data, err = io.ReadAll(stream)
        if err != nil {
            log.Printf("Error loading sound %v: %v", name, err)
            return
        }

        log.Printf("  loaded %v", name)
    }

    return &SoundHandler{
        Make: func() *audio.Player {
            create.Do(load)
            return context.NewPlayerFromBytes(data)
        },
        MakeLoop: func() (*audio.Player, error) {
            create.Do(load)
            return context.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data)) + 1000))
        },
    }, nil
}

func (manager *SoundManager) LoadAll() error {

    sounds := audioFiles.AllSounds
    for _, sound := range sounds {
        handler, err := MakeSoundHandler(sound, manager.Context, manager.SampleRate)
        if err != nil {
            return fmt.Errorf("Error loading %v: %v", sound, err)
        }
        manager.Sounds[sound] = handler

        go handler.Make()
    }

    return nil
}

func (manager *SoundManager) Play(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        player := handler.Make()
        player.SetVolume(manager.GetVolume() / 100.0)
        player.Play()
    }
}

func (manager *SoundManager) PlayLoop(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        go func(){
            player, err := handler.MakeLoop()
            if err != nil {
                log.Printf("Failed to play audio loop %v: %v", name, err)
            } else {
                player.SetVolume(manager.GetVolume() / 100.0)
                go func(){
                    for {
                        select {
                            case <-manager.Quit.Done():
                                player.Close()
                                return
                            case <-time.After(100 * time.Millisecond):
                                if !player.IsPlaying() {
                                    return
                                }

                                player.SetVolume(manager.GetVolume() / 100.0)
                        }
                    }
                }()

                player.Play()
            }
        }()
    }
}

const GameFadeIn = 20
const GameFadeOut = 40
const GameWhiteFlash = 50

type GameCounter struct {
    Limit int
    Counter int
}

func (counter *GameCounter) Do(f func()) {
    if counter.Counter == 0 {
        f()
        counter.Counter = counter.Limit
    }
}

func (counter *GameCounter) Update() {
    if counter.Counter > 0 {
        counter.Counter -= 1
    }
}

type Game struct {
    Counters map[string]*GameCounter
    Player *Player
    Background *Background
    Bullets []*Bullet
    EnemyBullets []*Bullet
    Font *text.GoTextFaceSource
    Asteroids []*Asteroid
    Enemies []Enemy
    Powerups []Powerup
    Explosions []Explosion
    Bombs []*Bomb
    ShaderManager *ShaderManager
    ImageManager *ImageManager
    SoundManager *SoundManager
    FadeIn int
    FadeOut int
    WhiteFlash int

    ShakeTime uint64

    PlayerDied sync.Once

    BossMode bool
    // runs one time when the boss should appear
    DoBoss sync.Once
    // runs one time when the level ends
    DoEnd sync.Once
    End atomic.Bool

    MusicPlayer sync.Once

    Quit context.Context
    Cancel context.CancelFunc

    // number of ticks the game has run
    Counter uint64
}

func (game *Game) GetCounter(name string, limit int) *GameCounter {
    use, ok := game.Counters[name]
    if ok {
        return use
    }

    counter := GameCounter{
        Counter: 0,
        Limit: limit,
    }

    game.Counters[name] = &counter

    return game.Counters[name]
}

func (game *Game) Close() {
    game.Cancel()
}

func (game *Game) MakeEnemy(x float64, y float64, kind int, move Movement) error {
    var enemy Enemy
    var err error

    switch kind {
        case 0:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy1)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy1(x, y, raw, pic, move)
        case 1:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy2)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 2:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy3)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 3:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy4)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 4:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy5)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)

    }

    if err != nil {
        return err
    }

    game.Enemies = append(game.Enemies, enemy)

    return nil
}

func (game *Game) MakeEnemies(count int) error {

    for i := 0; i < count; i++ {
        var generator chan Coordinate
        switch rand.Intn(5) {
            case 0: generator = MakeGroupGeneratorX()
            case 1: generator = MakeGroupGeneratorVertical(rand.Intn(3) + 3)
            case 2: generator = MakeGroupGeneratorCircle(100, 6)
            case 3: generator = MakeGroupGenerator1x2()
            case 4: generator = MakeGroupGenerator2x2()
        }

        x := randomFloat(50, ScreenWidth - 50)
        y := float64(-200)
        kind := rand.Intn(5)

        move := makeMovement()

        for coord := range generator {
            err := game.MakeEnemy(x + coord.x, y + coord.y, kind, move.Copy())
            if err != nil {
                return err
            }
        }
    }

    return nil
}

var LevelEnd error = errors.New("end of level")

func (game *Game) UpdateCounters() {
    for _, counter := range game.Counters {
        counter.Update()
    }
}

func (game *Game) Shake() {
    game.ShakeTime = 10
}

func (game *Game) BigShake() {
    game.ShakeTime = 20
}

func (game *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
    if game.ShakeTime > 0 {
        geoM.Translate(randomFloat(-4, 4), randomFloat(-4, 4))
    }

    screen.DrawImage(offscreen, &ebiten.DrawImageOptions{
        GeoM: geoM,
    })
}

func (game *Game) Update(run *Run) error {

    game.UpdateCounters()

    game.Counter += 1
    if game.ShakeTime > 0 {
        game.ShakeTime -= 1
    }

    makeAnimatedExplosion := func(x float64, y float64, name gameImages.Image) {
        animation, err := game.ImageManager.LoadAnimation(name)
        if err == nil {
            game.Explosions = append(game.Explosions, MakeAnimatedExplosion(x, y, animation))
        } else {
            log.Printf("Could not load explosion sheet %v: %v", name, err)
        }
    }

    // this could be enemy.MakeExplosion() or something to let each enemy create its own explosion type
    explodeEnemy := func(enemy Enemy){
        x, y := enemy.Coords()
        makeAnimatedExplosion(x, y, gameImages.ImageExplosion2)
    }

    explodeAsteroid := func(asteroid *Asteroid){
        makeAnimatedExplosion(asteroid.x, asteroid.y, gameImages.ImageExplosion3)
    }

    playerDied := func(){
        game.SoundManager.Play(audioFiles.AudioExplosion3)
        game.End.Store(true)

        makeAnimatedExplosion(game.Player.x, game.Player.y, gameImages.ImageExplosion2)
    }

    if game.End.Load() {
        game.DoEnd.Do(func(){
            game.FadeOut = GameFadeOut * 3
        })
    }

    if game.FadeOut > 0 {
        game.FadeOut -= 1

        if game.FadeOut == 0 {
            return LevelEnd
        }
    }

    if game.FadeIn < GameFadeIn {
        game.FadeIn += 1
    }

    if game.WhiteFlash > 0 {
        game.WhiteFlash -= 1
    }

    game.MusicPlayer.Do(func(){
        game.SoundManager.PlayLoop(audioFiles.AudioStellarPulseSong)
    })

    game.Background.Update()

    if game.Player.IsAlive() {
        err := game.Player.HandleKeys(game, run)
        if err != nil {
            return err
        }

        game.Player.Move()
    }

    for _, asteroid := range game.Asteroids {
        asteroid.Move()
        if asteroid.Collide(game.Player, game.ImageManager) {
            game.Player.Damage(2)
            asteroid.Damage(2)

            if ! game.Player.IsAlive() {
                game.PlayerDied.Do(playerDied)
            }

            if !asteroid.IsAlive() {
                game.Shake()

                game.SoundManager.Play(audioFiles.AudioExplosion3)
                explodeAsteroid(asteroid)
            }
        }
    }

    var powerupOut []Powerup
    for _, powerup := range game.Powerups {
        powerup.Move()
        if powerup.Collide(game.Player, game.ImageManager) {
            powerup.Activate(game.Player, game.SoundManager)
        }

        if powerup.IsAlive() {
            powerupOut = append(powerupOut, powerup)
        }
    }
    game.Powerups = powerupOut

    for _, enemy := range game.Enemies {
        bullets := enemy.Move(game.Player, game.ImageManager)
        game.EnemyBullets = append(game.EnemyBullets, bullets...)

        if game.Player.IsAlive() {
            collideX, collideY, isCollide := enemy.CollidePlayer(game.Player)

            if isCollide {
                game.GetCounter("player hit enemy", 30).Do(func(){
                    game.SoundManager.Play(audioFiles.AudioHit1)
                })

                makeAnimatedExplosion(collideX, collideY, gameImages.ImageHit2)

                enemy.Damage(2)
                game.Player.Damage(2)
                if ! game.Player.IsAlive() {
                    game.PlayerDied.Do(playerDied)
                }

                if ! enemy.IsAlive() {
                    game.Player.Score += 1
                    game.Player.Kills += 1
                    game.SoundManager.Play(audioFiles.AudioExplosion3)

                    explodeEnemy(enemy)
                }
            }
        }
    }

    explosionOut := make([]Explosion, 0)
    for _, explosion := range game.Explosions {
        explosion.Move()
        if explosion.IsAlive() {
            explosionOut = append(explosionOut, explosion)
        }
    }
    game.Explosions = explosionOut

    // run bullet physics at 3x
    for i := 0; i < 3; i++ {
        var outBullets []*Bullet
        for _, bullet := range game.Bullets {
            bullet.Move()

            for _, asteroid := range game.Asteroids {
                if asteroid.IsAlive() && asteroid.Collision(bullet.x, bullet.y, game.ImageManager) {
                    asteroid.Damage(bullet.Strength)
                    game.Player.Score += 1
                    bullet.Damage(1)

                    game.SoundManager.Play(audioFiles.AudioHit1)

                    animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit)
                    if err != nil {
                        log.Printf("Could not load hit animation: %v", err)
                    } else {
                        game.Explosions = append(game.Explosions, MakeAnimatedExplosion(bullet.x, bullet.y, animation))
                    }

                    if ! asteroid.IsAlive() {
                        game.Shake()
                        game.SoundManager.Play(audioFiles.AudioExplosion3)
                        animation, err := game.ImageManager.LoadAnimation(gameImages.ImageExplosion3)
                        if err == nil {
                            game.Explosions = append(game.Explosions, MakeAnimatedExplosion(asteroid.x, asteroid.y, animation))
                        }
                        break
                    }
                }
            }

            if bullet.IsAlive() {
                for _, enemy := range game.Enemies {
                    if enemy.IsAlive() && enemy.Collision(bullet.x, bullet.y) {
                        game.Player.Score += 1
                        bullet.Damage(1)
                        enemy.Damage(bullet.Strength)
                        if ! enemy.IsAlive() {
                            game.Shake()
                            game.Player.Kills += 1
                            game.SoundManager.Play(audioFiles.AudioExplosion3)

                            // create a powerup every X kills
                            if game.Player.Kills % 20 == 0 {
                                game.Powerups = append(game.Powerups, MakeRandomPowerup(randomFloat(10, ScreenWidth-10), -20))
                            }

                            explodeEnemy(enemy)

                            // create a powerup where the enemy died every once in a while
                            if rand.Intn(20) == 0 {
                                x, y := enemy.Coords()
                                game.Powerups = append(game.Powerups, MakeRandomPowerup(x, y))
                            }
                        }

                        game.SoundManager.Play(audioFiles.AudioHit1)

                        animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit)
                        if err != nil {
                            log.Printf("Could not load hit animation: %v", err)
                        } else {
                            game.Explosions = append(game.Explosions, MakeAnimatedExplosion(bullet.x, bullet.y, animation))
                        }
                        break
                    }
                }
            }

            if bullet.IsAlive() {
                outBullets = append(outBullets, bullet)
            }
        }
        game.Bullets = outBullets

        var outEnemyBullets []*Bullet
        for _, bullet := range game.EnemyBullets {
            bullet.Move()

            if game.Player.IsAlive() && game.Player.Collide(bullet.x, bullet.y) {
                game.SoundManager.Play(audioFiles.AudioHit2)

                game.Player.Damage(bullet.Strength)
                if ! game.Player.IsAlive() {
                    game.PlayerDied.Do(playerDied)
                }

                animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit2)
                if err == nil {
                    game.Explosions = append(game.Explosions, MakeAnimatedExplosion(bullet.x, bullet.y, animation))
                } else {
                    log.Printf("Could not load explosion sheet: %v", err)
                }

                bullet.Damage(1)
            }

            if bullet.IsAlive() {
                outEnemyBullets = append(outEnemyBullets, bullet)
            }
        }
        game.EnemyBullets = outEnemyBullets
    }

    bombExplode := func(bomb *Bomb){
        game.WhiteFlash = GameWhiteFlash
        game.BigShake()
        game.SoundManager.Play(audioFiles.AudioExplosion3)

        var bombDamage float64 = 50

        for _, enemy := range game.Enemies {
            x, y := enemy.Coords()
            if enemy.IsAlive() && bomb.Touch(x, y) {
                enemy.Damage(bombDamage)
                if ! enemy.IsAlive() {
                    explodeEnemy(enemy)
                }
            }
        }

        for _, asteroid := range game.Asteroids {
            if asteroid.IsAlive() && bomb.Touch(asteroid.x, asteroid.y) {
                asteroid.Damage(bombDamage)
                if ! asteroid.IsAlive() {
                    game.Shake()
                    explodeAsteroid(asteroid)
                }
            }
        }

    }
    bombOut := make([]*Bomb, 0)
    for _, bomb := range game.Bombs {
        bomb.Update(bombExplode)
        if bomb.IsAlive() {
            bombOut = append(bombOut, bomb)
        }
    }
    game.Bombs = bombOut

    enemyOut := make([]Enemy, 0)
    for _, enemy := range game.Enemies {
        if enemy.IsAlive() {
            enemyOut = append(enemyOut, enemy)
        }
    }
    game.Enemies = enemyOut

    asteroidOut := make([]*Asteroid, 0)
    for _, asteroid := range game.Asteroids {
        if asteroid.IsAlive() {
            asteroidOut = append(asteroidOut, asteroid)
        }
    }
    game.Asteroids = asteroidOut

    if rand.Intn(6000) == 0 {
        game.Powerups = append(game.Powerups, MakeRandomPowerup(randomFloat(10, ScreenWidth-10), -20))
    }

    if !game.BossMode && !game.End.Load(){
        if len(game.Enemies) == 0 || (len(game.Enemies) < 10 && rand.Intn(100) == 0) {
            game.MakeEnemies(1)
        }

        if len(game.Asteroids) < 15 && rand.Intn(200) == 0 {
            game.Asteroids = append(game.Asteroids, MakeAsteroid(randomFloat(-50, ScreenWidth + 50), -50))
        }

        // create the boss after 2 minutes
        const bossTime = 60 * 120
        // const bossTime = 60 * 1
        if debugForceBoss || (game.Counter > bossTime && rand.Intn(1000) == 0) {
            game.BossMode = true
            game.DoBoss.Do(func(){
                log.Printf("Created boss!")
                boss1Pic, rawImage, err := game.ImageManager.LoadImage(gameImages.ImageBoss1)
                if err != nil {
                    log.Printf("Unable to load boss: %v", err)
                } else {
                    boss, err := MakeBoss1(ScreenWidth / 2, -150, rawImage, boss1Pic)
                    if err != nil {
                        log.Printf("Unable to make boss: %v", err)
                    } else {
                        game.Enemies = append(game.Enemies, boss)

                        go func(){
                            for {
                                select {
                                    case <-game.Quit.Done():
                                        return
                                    case <-boss.Dead():
                                        game.End.Store(true)
                                        return
                                }
                            }
                        }()

                    }
                }
            })
        }

    }

    return nil
}

// draw a big orange circle that fades out towards the edge of the circle
func (game *Game) TestAlphaCircle(screen *ebiten.Image){
    {
        options := &ebiten.DrawRectShaderOptions{}
        cx := 300
        cy := 300
        radius := 100
        options.GeoM.Translate(float64(cx - radius), float64(cy - radius))
        // options.Blend = AlphaBlender
        // options.Images[0] = player.pic
        options.Uniforms = make(map[string]interface{})
        // radians = math.Pi * 90 / 180
        // log.Printf("Red: %v", radians)
        // red := vec4(abs(sin(Red) / 3), 0, 0, 0)
        options.Uniforms["Center"] = []float32{float32(cx), float32(cy)}
        options.Uniforms["Radius"] = float32(radius)
        options.Uniforms["CenterAlpha"] = float32(0.9)
        options.Uniforms["EdgeAlpha"] = float32(0.2)
        options.Uniforms["Color"] = []float32{1, 0.5, 0}

        screen.DrawRectShader(radius * 2, radius * 2, game.ShaderManager.AlphaCircleShader, options)
    }
}

func (game *Game) Draw(screen *ebiten.Image) {
    game.Background.Draw(screen)

    for _, enemy := range game.Enemies {
        enemy.Draw(screen, game.ShaderManager)
    }

    for _, powerup := range game.Powerups {
        powerup.Draw(screen, game.ImageManager, game.ShaderManager)
    }

    for _, explosion := range game.Explosions {
        explosion.Draw(screen, game.ShaderManager)
    }

    for _, asteroid := range game.Asteroids {
        asteroid.Draw(screen, game.ImageManager, game.ShaderManager)
    }

    // ebitenutil.DebugPrint(screen, "debugging")
    if game.Player.IsAlive() {
        game.Player.Draw(screen, game.ShaderManager, game.ImageManager, game.Font)
    }

    for _, bullet := range game.Bullets {
        bullet.Draw(screen)
    }

    for _, bullet := range game.EnemyBullets {
        bullet.Draw(screen)
    }

    for _, bomb := range game.Bombs {
        bomb.Draw(screen, game.ImageManager, game.ShaderManager)
    }

    if game.WhiteFlash > 0 {
        flash := premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: uint8(game.WhiteFlash * 255 / GameWhiteFlash)})
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, &flash, true)
    }

    if game.FadeIn < GameFadeIn {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, &color.RGBA{R: 0, G: 0, B: 0, A: uint8(255 - game.FadeIn * 255 / GameFadeIn)}, true)
    }

    if game.FadeOut > 0 && game.FadeOut <= GameFadeOut {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, &color.RGBA{R: 0, G: 0, B: 0, A: uint8(255 - game.FadeOut * 255 / GameFadeOut)}, true)
    }

    // vector.StrokeRect(screen, 0, 0, 100, 100, 3, &color.RGBA{R: 255, G: 0, B: 0, A: 128}, true)
    // vector.DrawFilledRect(screen, 0, 0, 100, 100, &color.RGBA{R: 255, G: 0, B: 0, A: 64}, true)
}

func (game *Game) PreloadAssets() error {
    // preload assets
    _, err := game.ImageManager.LoadAnimation(gameImages.ImageExplosion2)
    if err != nil {
        return err
    }

    return nil
}

func premultiplyAlpha(value color.RGBA) color.RGBA {
    a := float32(value.A) / 255.0

    return color.RGBA{
        R: uint8(float32(value.R) * a),
        G: uint8(float32(value.G) * a),
        B: uint8(float32(value.B) * a),
        A: value.A,
    }
}

type RunMode int
const (
    RunGame RunMode = iota
    RunMenu RunMode = iota
)

type Run struct {
    Game *Game
    Menu *Menu
    Mode RunMode
    Quit context.Context
    Cancel context.CancelFunc
    Volume float64
}

func (run *Run) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
    if run.Game != nil && run.Mode == RunGame {
        run.Game.DrawFinalScreen(screen, offscreen, geoM)
    } else {
        screen.DrawImage(offscreen, &ebiten.DrawImageOptions{
            GeoM: geoM,
        })
    }
}

func (run *Run) GetVolume() float64 {
    return run.Volume
}

func (run *Run) updateVolume(){
    if run.Game != nil {
        run.Game.SoundManager.SetVolume(run.Volume)
    }
}

func (run *Run) SetVolume(volume float64){
    run.Volume = volume
    run.updateVolume()
}

func (run *Run) IncreaseVolume() {
    run.Volume += 10
    if run.Volume > 100 {
        run.Volume = 100
    }
    run.updateVolume()
}

func (run *Run) DecreaseVolume() {
    run.Volume -= 10
    if run.Volume < 0 {
        run.Volume = 0
    }
    run.updateVolume()
}

func (run *Run) Update() error {
    switch run.Mode {
        case RunGame:
            err := run.Game.Update(run)
            if errors.Is(err, LevelEnd) {
                run.Game.Close()
                run.Game = nil
                run.Mode = RunMenu
                return nil
            } else {
                return err
            }
        case RunMenu: return run.Menu.Update(run)
    }

    return fmt.Errorf("Unknown mode %v", run.Mode)
}

func (run *Run) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (run *Run) Draw(screen *ebiten.Image) {
    if run.Game != nil {
        run.Game.Draw(screen)
    }

    if run.Mode == RunMenu {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 92}, true)
        run.Menu.Draw(screen)
    }

    /*
    switch run.Mode {
        case RunGame: run.Game.Draw(screen)
        case RunMenu: run.Menu.Draw(screen)
    }
    */
}

func MakeGame(audioContext *audio.Context, run *Run) (*Game, error) {
    player, err := MakePlayer(ScreenWidth / 2, ScreenHeight - 100)
    if err != nil {
        return nil, err
    }

    background, err := MakeBackground()
    if err != nil {
        return nil, err
    }

    font, err := fontLib.LoadFont()
    if err != nil {
        return nil, err
    }

    shaderManager, err := MakeShaderManager()
    if err != nil {
        return nil, err
    }

    quitContext, cancel := context.WithCancel(context.Background())

    soundManager, err := MakeSoundManager(quitContext, audioContext, run.GetVolume())
    if err != nil {
        cancel()
        return nil, err
    }

    game := Game{
        Counters: make(map[string]*GameCounter),
        Background: background,
        Player: player,
        Font: font,
        ShaderManager: shaderManager,
        ImageManager: MakeImageManager(),
        SoundManager: soundManager,
        FadeIn: 0,
        BossMode: false,
        Quit: quitContext,
        Cancel: cancel,
    }

    err = game.MakeEnemies(2)
    if err != nil {
        cancel()
        return nil, err
    }

    // for debugging
    game.Powerups = append(game.Powerups, MakePowerupWeapon(randomFloat(10, ScreenWidth-10), -20))

    err = game.PreloadAssets()
    if err != nil {
        cancel()
        return nil, err
    }

    return &game, nil
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    profile := false

    if profile {
        cpuProfile, err := os.Create("profile.cpu")
        if err != nil {
            log.Printf("Unable to create profile.cpu: %v", err)
        } else {
            defer cpuProfile.Close()
            pprof.StartCPUProfile(cpuProfile)
            defer pprof.StopCPUProfile()
        }
    }

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("Shooter")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    log.Printf("Loading objects")

    audioContext := audio.NewContext(48000)

    var initialVolume float64 = 80

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    menu, err := createMenu(quit, audioContext, initialVolume)
    if err != nil {
        log.Printf("Unable to create menu: %v", err)
        return
    }

    /*
    game, err := MakeGame()
    if err != nil {
        log.Printf("Unable to create game: %v", err)
        return
    }
    */

    run := Run{
        Mode: RunMenu,
        Game: nil,
        Quit: quit,
        Cancel: cancel,
        Menu: menu,
        Volume: initialVolume,
    }

    log.Printf("Running")
    err = ebiten.RunGame(&run)
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }

    log.Printf("Bye!")

    if profile {
        memProfile, err := os.Create("profile.mem")
        if err != nil {
            log.Printf("Unable to create profile.mem: %v", err)
        } else {
            defer memProfile.Close()
            pprof.WriteHeapProfile(memProfile)
        }
    }
}

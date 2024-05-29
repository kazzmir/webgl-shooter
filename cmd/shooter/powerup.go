package main

import (
    audioFiles "github.com/kazzmir/webgl-shooter/audio"
    gameImages "github.com/kazzmir/webgl-shooter/images"

    "math"
    "math/rand"
    "log"
    "image"
    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
)

type Collidable interface {
    Bounds() image.Rectangle
    Collide(x float64, y float64) bool
}

func isColliding(from image.Rectangle, collidable Collidable) bool {
    playerBounds := collidable.Bounds()

    overlap := from.Intersect(playerBounds)
    if overlap.Empty() {
        return false
    }

    samplePoints := int(math.Sqrt(float64(overlap.Dx() * overlap.Dy())))
    if samplePoints < 3 {
        samplePoints = 3
    }

    for i := 0; i < samplePoints; i++ {
        x := randomFloat(float64(overlap.Min.X), float64(overlap.Max.X))
        y := randomFloat(float64(overlap.Min.Y), float64(overlap.Max.Y))

        if collidable.Collide(x, y) {
            return true
        }

    }

    return false
}

type Powerup interface {
    Move()
    Collide(player *Player, imageManager *ImageManager) bool
    Activate(player *Player, soundManager *SoundManager)
    IsAlive() bool
    Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager)
}

type PowerupEnergy struct {
    x, y float64
    velocityX float64
    velocityY float64
    activated bool
    angle uint64
}

func MakePowerupEnergy(x float64, y float64) Powerup {
    return &PowerupEnergy{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 1.5,
        activated: false,
        angle: 0,
    }
}

func (powerup *PowerupEnergy) Move() {
    powerup.x += powerup.velocityX
    powerup.y += powerup.velocityY
    powerup.angle += 1
}

func (powerup *PowerupEnergy) IsAlive() bool {
    return !powerup.activated && powerup.y < ScreenHeight + 20
}

func (powerup *PowerupEnergy) Activate(player *Player, soundManager *SoundManager){
    if !powerup.activated {
        player.PowerupEnergy = 60 * 10
        powerup.activated = true
        soundManager.Play(audioFiles.AudioEnergy)
    }
}

var PowerupColor color.Color = color.RGBA{R: 0x7e, G: 0x29, B: 0xd6, A: 0xff}

func (powerup *PowerupEnergy) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager){

    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup2)
    if err != nil {
        log.Printf("Could not load powerup image: %v", err)
        return
    }

    width := float64(pic.Bounds().Dx())
    height := float64(pic.Bounds().Dy())

    // x1 := powerup.x - width / 2
    // y1 := powerup.y - height / 2
    options := &ebiten.DrawImageOptions{}

    // translate such that center is at origin
    options.GeoM.Translate(-width/2, -height/2)
    // rotate
    options.GeoM.Rotate(float64(powerup.angle) * math.Pi / 180.0)

    options.GeoM.Translate(powerup.x, powerup.y)
    screen.DrawImage(pic, options)

    shaderOptions := &ebiten.DrawRectShaderOptions{}
    shaderOptions.GeoM.Translate(-width/2, -height/2)
    shaderOptions.GeoM.Rotate(float64(powerup.angle) * math.Pi / 180.0)
    shaderOptions.GeoM.Translate(powerup.x, powerup.y)
    shaderOptions.Blend = AlphaBlender
    shaderOptions.Images[0] = pic
    shaderOptions.Uniforms = make(map[string]interface{})
    // options.Uniforms["Color"] = []float32{0, 0, float32((math.Sin(float64(player.Counter) * 7 * math.Pi / 180.0) + 1) / 2), 1}

    alpha := float32((math.Sin(float64(powerup.angle) * 7 * math.Pi / 180.0) + 1) / 2)

    r, g, b, _ := PowerupColor.RGBA()
    useColor := color.RGBA{R: uint8(r / 255), G: uint8(g / 255), B: uint8(b / 255), A: uint8(255.0 * alpha)}
    // useColor.a = float32((math.Sin(float64(powerup.angle) * 7 * math.Pi / 180.0) + 1) / 2)
    shaderOptions.Uniforms["Color"] = toFloatArray(useColor)
    // v := []float32{float32(0x7e) / 255.0, float32(0x29) / 255.0, float32(0xd6) / 255.0, 0xff / 255.0}
    // options.Uniforms["Color"] = []float32{0, 0, 1, 1}
    screen.DrawRectShader(pic.Bounds().Dx(), pic.Bounds().Dy(), shaders.EdgeShader, shaderOptions)
}

func (powerup *PowerupEnergy) Collide(player *Player, imageManager *ImageManager) bool {
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup1)
    if err != nil {
        return false
    }

    translate := image.Point{
        X: int(powerup.x - float64(pic.Bounds().Dx())/2),
        Y: int(powerup.y - float64(pic.Bounds().Dy())/2),
    }
    bounds := pic.Bounds().Add(translate)
    return isColliding(bounds, player)
}

type PowerupHealth struct {
    x, y float64
    velocityX float64
    velocityY float64
    activated bool
    counter uint64
}

func (powerup *PowerupHealth) Move() {
    powerup.x += powerup.velocityX
    powerup.y += powerup.velocityY
    powerup.counter += 1
}

func (powerup *PowerupHealth) IsAlive() bool {
    return !powerup.activated && powerup.y < ScreenHeight + 20
}

func (powerup *PowerupHealth) Collide(player *Player, imageManager *ImageManager) bool {
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup3)
    if err != nil {
        return false
    }

    translate := image.Point{
        X: int(powerup.x - float64(pic.Bounds().Dx())/2),
        Y: int(powerup.y - float64(pic.Bounds().Dy())/2),
    }
    bounds := pic.Bounds().Add(translate)
    return isColliding(bounds, player)
}

func (powerup *PowerupHealth) Activate(player *Player, soundManager *SoundManager){
    if !powerup.activated {
        player.Health = math.Min(player.MaxHealth, player.Health + 20)
        powerup.activated = true
        soundManager.Play(audioFiles.AudioHealth)
    }
}

func (powerup *PowerupHealth) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager){

    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup3)
    if err != nil {
        log.Printf("Could not load powerup image: %v", err)
        return
    }

    width := float64(pic.Bounds().Dx())
    height := float64(pic.Bounds().Dy())

    // x1 := powerup.x - width / 2
    // y1 := powerup.y - height / 2
    options := &ebiten.DrawImageOptions{}

    // translate such that center is at origin
    options.GeoM.Translate(-width/2, -height/2)
    options.GeoM.Translate(powerup.x, powerup.y)
    screen.DrawImage(pic, options)

    shaderOptions := &ebiten.DrawRectShaderOptions{}
    shaderOptions.GeoM.Translate(-width/2, -height/2)
    shaderOptions.GeoM.Translate(powerup.x, powerup.y)
    shaderOptions.Uniforms = make(map[string]interface{})
    v := uint8(math.Abs(math.Sin(float64(powerup.counter) * 4 * math.Pi / 180.0) / 3) * 255)
    shaderOptions.Uniforms["Red"] = toFloatArray(color.RGBA{R: v, G: v, B: v, A: 0})
    shaderOptions.Blend = AlphaBlender
    shaderOptions.Images[0] = pic
    bounds := pic.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, shaderOptions)
}

func MakePowerupHealth(x float64, y float64) Powerup {
    return &PowerupHealth{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 1.5,
        activated: false,
    }
}

func MakeRandomPowerup(x float64, y float64) Powerup {
    switch rand.Intn(2) {
        case 0: return MakePowerupEnergy(x, y)
        case 1: return MakePowerupHealth(x, y)
    }

    return nil
}

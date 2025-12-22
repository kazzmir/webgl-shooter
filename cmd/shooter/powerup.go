package main

import (
    audioFiles "github.com/kazzmir/webgl-shooter/audio"
    gameImages "github.com/kazzmir/webgl-shooter/images"

    "math"
    "math/rand/v2"
    "log"
    "image"
    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
)

func drawCenter(screen *ebiten.Image, img *ebiten.Image, x float64, y float64, extra ebiten.GeoM) {
    width := float64(img.Bounds().Dx())
    height := float64(img.Bounds().Dy())

    options := &ebiten.DrawImageOptions{}

    options.GeoM.Concat(extra)
    // translate such that center is at origin
    options.GeoM.Translate(-width/2, -height/2)
    options.GeoM.Translate(x, y)
    screen.DrawImage(img, options)
}

func drawGlow(screen *ebiten.Image, img *ebiten.Image, shaders *ShaderManager, x float64, y float64, counter uint64, extra ebiten.GeoM) {

    width := float64(img.Bounds().Dx())
    height := float64(img.Bounds().Dy())

    // x1 := powerup.x - width / 2
    // y1 := powerup.y - height / 2
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Concat(extra)

    // translate such that center is at origin
    options.GeoM.Translate(-width/2, -height/2)
    options.GeoM.Translate(x, y)
    screen.DrawImage(img, options)

    shaderOptions := &ebiten.DrawRectShaderOptions{}
    shaderOptions.GeoM.Concat(extra)
    shaderOptions.GeoM.Translate(-width/2, -height/2)
    shaderOptions.GeoM.Translate(x, y)
    shaderOptions.Uniforms = make(map[string]interface{})
    v := uint8(math.Abs(math.Sin(float64(counter) * 4 * math.Pi / 180.0) / 3) * 255)
    shaderOptions.Uniforms["Red"] = toFloatArray(color.RGBA{R: v, G: v, B: v, A: 0})
    shaderOptions.Blend = AlphaBlender
    shaderOptions.Images[0] = img
    bounds := img.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, shaderOptions)
}

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

        // FIXME: this needs to take into account the 'from' image to see if there is a non-alpha pixel at x,y
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
    Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM)
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

func (powerup *PowerupEnergy) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM){

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

    options.GeoM.Concat(extra)

    // translate such that center is at origin
    options.GeoM.Translate(-width/2, -height/2)
    // rotate
    options.GeoM.Rotate(float64(powerup.angle) * math.Pi / 180.0)

    options.GeoM.Translate(powerup.x, powerup.y)
    screen.DrawImage(pic, options)

    shaderOptions := &ebiten.DrawRectShaderOptions{}
    shaderOptions.GeoM.Concat(extra)
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

func (powerup *PowerupHealth) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM){

    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup3)
    if err != nil {
        log.Printf("Could not load powerup image: %v", err)
        return
    }

    drawGlow(screen, pic, shaders, powerup.x, powerup.y, powerup.counter, extra)
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

type PowerupWeapon struct {
    x, y float64
    velocityX float64
    velocityY float64
    activated bool
    counter uint64
}

func (powerup *PowerupWeapon) Move() {
    powerup.x += powerup.velocityX
    powerup.y += powerup.velocityY
    powerup.counter += 1
}

func (powerup *PowerupWeapon) IsAlive() bool {
    return !powerup.activated && powerup.y < ScreenHeight + 20
}

func (powerup *PowerupWeapon) Collide(player *Player, imageManager *ImageManager) bool {
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup4)
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

func (powerup *PowerupWeapon) Activate(player *Player, soundManager *SoundManager){
    if !powerup.activated {
        player.EnableNextGun()
        powerup.activated = true
        // FIXME: find a new sound
        soundManager.Play(audioFiles.AudioHealth)
    }
}

func (powerup *PowerupWeapon) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM){

    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup4)
    if err != nil {
        log.Printf("Could not load powerup image: %v", err)
        return
    }

    blurred, err := imageManager.BlurImage(gameImages.ImagePowerup4, 1.2, 2, color.RGBA{R: 255, G: 255, B: 0, A: 255})
    if err != nil {
        log.Printf("Unable to create blur: %v", err)
        return
    }

    drawCenter(screen, blurred, powerup.x, powerup.y, extra)
    drawGlow(screen, pic, shaders, powerup.x, powerup.y, powerup.counter, extra)
}

type PowerupBomb struct {
    x, y float64
    velocityX float64
    velocityY float64
    activated bool
    counter uint64
}

func (powerup *PowerupBomb) IsAlive() bool {
    return !powerup.activated && powerup.y < ScreenHeight + 20
}

func (powerup *PowerupBomb) Move() {
    powerup.x += powerup.velocityX
    powerup.y += powerup.velocityY
    powerup.counter += 1
}

func (powerup *PowerupBomb) Activate(player *Player, soundManager *SoundManager){
    if !powerup.activated {
        player.IncreaseBombs()
        powerup.activated = true
        // FIXME: find a new sound
        soundManager.Play(audioFiles.AudioHealth)
    }
}

func (powerup *PowerupBomb) Collide(player *Player, imageManager *ImageManager) bool {
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerupBomb)
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

func (powerup *PowerupBomb) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM){
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerupBomb)
    if err != nil {
        return
    }
    drawGlow(screen, pic, shaders, powerup.x, powerup.y, powerup.counter, extra)
}

type PowerupEnergyIncrease struct {
    x, y float64
    velocityX, velocityY float64
    activated bool
    counter uint64
    increase uint64
}

func (powerup *PowerupEnergyIncrease) IsAlive() bool {
    return !powerup.activated && powerup.y < ScreenHeight + 20
}

func (powerup *PowerupEnergyIncrease) Move() {
    powerup.x += powerup.velocityX
    powerup.y += powerup.velocityY
    powerup.counter += 1
}

func (powerup *PowerupEnergyIncrease) Activate(player *Player, soundManager *SoundManager){
    if !powerup.activated {
        player.IncreaseMaxEnergy(float64(powerup.increase))
        powerup.activated = true
        soundManager.Play(audioFiles.AudioHealth)
    }
}

func (powerup *PowerupEnergyIncrease) Collide(player *Player, imageManager *ImageManager) bool {
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup5)
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

func (powerup *PowerupEnergyIncrease) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager, extra ebiten.GeoM){
    pic, _, err := imageManager.LoadImage(gameImages.ImagePowerup5)
    if err != nil {
        return
    }
    drawGlow(screen, pic, shaders, powerup.x, powerup.y, powerup.counter, extra)
}

func MakePowerupEnergyIncrease(x float64, y float64) Powerup {
    return &PowerupEnergyIncrease{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 1.5,
        activated: false,
        counter: 0,
        increase: 25,
    }
}

func MakePowerupWeapon(x float64, y float64) Powerup {
    return &PowerupWeapon{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 1.8,
        activated: false,
    }
}

func MakePowerupBomb(x float64, y float64) Powerup {
    return &PowerupBomb{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 1.8,
        activated: false,
    }
}

func MakeRandomPowerup(x float64, y float64) Powerup {
    switch rand.N(5) {
        case 0: return MakePowerupEnergy(x, y)
        case 1: return MakePowerupHealth(x, y)
        case 2: return MakePowerupWeapon(x, y)
        case 3: return MakePowerupBomb(x, y)
        case 4: return MakePowerupEnergyIncrease(x, y)
    }

    return nil
}

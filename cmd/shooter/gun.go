package main

import (
    gameImages "github.com/kazzmir/webgl-shooter/images"
    audioFiles "github.com/kazzmir/webgl-shooter/audio"

    _ "log"

    "strconv"
    "math"

    "image"
    "image/color"
    "image/draw"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Gun interface {
    Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error)
    Rate() float64
    DoSound(soundManager *SoundManager)
    DrawIcon(screen *ebiten.Image, imageManager *ImageManager, x float64, y float64, textFace *text.GoTextFace)
    IsEnabled() bool
    SetEnabled(bool)
    IncreaseExperience(float64)
    Update()
    EnergyUsed() float64

    // a value from 0.0 to 1.0 indicating how close the gun is to leveling up
    LevelPercent() float64
}

type BasicGun struct {
    enabled bool
    level int
    experience float64

    // for tracking fire rate
    counter int
}

func experienceForLevel(level int) float64 {
    return 100 * math.Pow(1.6, float64(level))
}

func (basic *BasicGun) EnergyUsed() float64 {
    if basic.level <= 2 {
        return 1
    } else if basic.level <= 5 {
        return 2
    }

    return 3
}

func (basic *BasicGun) IncreaseExperience(amount float64) {
    basic.experience += amount
    // log.Printf("BasicGun gained %f experience, total %f", amount, basic.experience)
    if basic.experience >= experienceForLevel(basic.level) {
        basic.experience -= experienceForLevel(basic.level)
        basic.level += 1
    }
}

func (basic *BasicGun) LevelPercent() float64 {
    required := experienceForLevel(basic.level)
    if required == 0 {
        return 0.0
    }
    return basic.experience / required
}

func (basic *BasicGun) Update() {
    if basic.counter > 0 {
        basic.counter -= 1
    }
}

func (basic *BasicGun) IsEnabled() bool {
    return basic.enabled
}

func (basic *BasicGun) SetEnabled(enabled bool) {
    basic.enabled = enabled
}

func (basic *BasicGun) Rate() float64 {
    return 10 + float64(basic.level)
}

func drawGunBox(screen *ebiten.Image, x float64, y float64, color_ color.Color, icon *ebiten.Image) {
    size := 20
    vector.StrokeRect(screen, float32(x), float32(y), float32(size), float32(size), 2, color_, true)

    padding := 4

    if icon != nil {
        bounds := icon.Bounds()

        scaleX := float64(size-padding) / float64(bounds.Dx())
        scaleY := float64(size-padding) / float64(bounds.Dy())

        where := screen.SubImage(image.Rect(int(x), int(y), int(x)+size, int(y)+size)).(*ebiten.Image)

        var options ebiten.DrawImageOptions
        options.GeoM.Scale(scaleX, scaleY)
        options.GeoM.Translate(x+float64(padding)/2, y+float64(padding)/2)

        where.DrawImage(icon, &options)
    }
}

func iconColor(enabled bool) color.Color {
    if enabled {
        return color.White
    } else {
        return color.RGBA{R: 255, G: 0, B: 0, A: 255}
    }
}

func (basic *BasicGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, x float64, y float64, textFace *text.GoTextFace) {
    pic, _, err := imageManager.LoadImage(gameImages.ImageBullet)
    if err != nil {
        pic = nil
    }

    drawGunBox(screen, x, y, iconColor(basic.enabled), pic)

    levelGaugeX := x + 20 + 5
    gaugeWidth := float32(10)
    gaugeHeight := float32(20)

    vector.StrokeRect(screen, float32(levelGaugeX), float32(y), gaugeWidth, gaugeHeight, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
    vector.FillRect(screen, float32(levelGaugeX)+1, float32(y)+gaugeHeight-float32(basic.LevelPercent()*float64(gaugeHeight-2)), gaugeWidth-2, float32(basic.LevelPercent()*float64(gaugeHeight-2)), color.RGBA{R: 0, G: 255, B: 0, A: 255}, false)

    op := &text.DrawOptions{}
    op.GeoM.Translate(levelGaugeX, y + float64(gaugeHeight) + 1)
    var color_ color.RGBA = color.RGBA{0xff, 0xff, 0xff, 0xff}
    op.ColorScale.ScaleWithColor(color_)
    text.Draw(screen, strconv.Itoa(basic.level + 1), textFace, op)
}

func (basic *BasicGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (basic *BasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if basic.enabled && basic.counter == 0 {
        pic, _, err := imageManager.LoadImage(gameImages.ImageBullet)
        if err != nil {
            return nil, err
        }

        basic.counter = int(60.0 / basic.Rate())

        if basic.level <= 2 {
            velocityY := -2.5

            bullet := Bullet{
                x: x,
                y: y,
                Strength: 1,
                health: 1,
                velocityX: 0,
                velocityY: velocityY,
                pic: pic,
                Gun: basic,
            }

            return []*Bullet{&bullet}, nil
        } else if basic.level <= 5 {
            velocityY := -2.5

            makeBullet := func(offsetX float64) *Bullet {
                return &Bullet{
                    x: x + offsetX,
                    y: y,
                    Strength: 1.1,
                    health: 1,
                    velocityX: 0,
                    velocityY: velocityY,
                    pic: pic,
                    Gun: basic,
                }
            }

            return []*Bullet{makeBullet(-6), makeBullet(6)}, nil
        } else {
            velocityY := -2.5

            makeBullet := func(offsetX float64, offsetY float64) *Bullet {
                return &Bullet{
                    x: x + offsetX,
                    y: y + offsetY,
                    Strength: 1.1,
                    health: 1,
                    velocityX: 0,
                    velocityY: velocityY,
                    pic: pic,
                    Gun: basic,
                }
            }

            return []*Bullet{makeBullet(-10, 3), makeBullet(10, 3), makeBullet(0, 0)}, nil
        }
    } else {
        return nil, nil
    }
}

type DualBasicGun struct {
    enabled bool
    counter int
    icon *ebiten.Image
    level int
    experience float64
}

func (dual *DualBasicGun) LevelPercent() float64 {
    required := experienceForLevel(dual.level)
    if required == 0 {
        return 0.0
    }
    return dual.experience / required
}

func (dual *DualBasicGun) IncreaseExperience(experience float64) {
}

func (dual *DualBasicGun) Update() {
    if dual.counter > 0 {
        dual.counter -= 1
    }
}

func (dual *DualBasicGun) EnergyUsed() float64 {
    return 2
}

func (dual *DualBasicGun) IsEnabled() bool {
    return dual.enabled
}

func (dual *DualBasicGun) SetEnabled(enabled bool) {
    dual.enabled = enabled
}

func (dual *DualBasicGun) Rate() float64 {
    return 7
}

func (dual *DualBasicGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, x float64, y float64, textFace *text.GoTextFace) {
    if dual.icon == nil {
        _, bullet, err := imageManager.LoadImage(gameImages.ImageBullet)
        if err == nil {
            icon := image.NewRGBA(image.Rect(0, 0, bullet.Bounds().Dx() * 2 + 5, bullet.Bounds().Dy()))
            /*
            for x := 0; x < icon.Bounds().Dx(); x++ {
                icon.Set(x, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
                icon.Set(x, icon.Bounds().Dy()-1, color.RGBA{R: 0, G: 255, B: 0, A: 255})
            }

            for y := 0; y < icon.Bounds().Dy(); y++ {
                icon.Set(0, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
                icon.Set(icon.Bounds().Dx()-1, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
            }
            */

            // draw.Draw(icon, icon.Bounds(), bullet, image.Point{X: 0, Y: 0}, draw.Src)
            draw.Draw(icon, icon.Bounds(), bullet, image.Point{X: 0, Y: 0}, draw.Src)
            draw.Draw(icon, icon.Bounds().Add(image.Point{X: bullet.Bounds().Dx() + 2, Y: 0}), bullet, image.Point{X: 0, Y: 0}, draw.Src)
            dual.icon = ebiten.NewImageFromImage(icon)
        }
    }

    /*
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(100, 100)
    screen.DrawImage(dual.icon, &options)
    */

    drawGunBox(screen, x, y, iconColor(dual.enabled), dual.icon)
}

func (dual *DualBasicGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (dual *DualBasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if dual.enabled && dual.counter == 0 {
        dual.counter = int(60.0 / dual.Rate())
        velocityY := -2.5

        pic, _, err := imageManager.LoadImage(gameImages.ImageBullet)
        if err != nil {
            return nil, err
        }

        bullet1 := Bullet{
            x: x - 10,
            y: y,
            Strength: 1,
            health: 1,
            velocityX: 0,
            velocityY: velocityY,
            pic: pic,
            Gun: dual,
        }

        bullet2 := bullet1
        bullet2.x += 20

        return []*Bullet{&bullet1, &bullet2}, nil
    } else {
        return nil, nil
    }
}

type BeamGun struct {
    enabled bool
    counter int
    level int
    experience float64
}

func (beam *BeamGun) LevelPercent() float64 {
    required := experienceForLevel(beam.level)
    if required == 0 {
        return 0.0
    }
    return beam.experience / required
}

func (beam *BeamGun) Update() {
    if beam.counter > 0 {
        beam.counter -= 1
    }
}

func (beam *BeamGun) IncreaseExperience(experience float64) {
    // TODO
}

func (beam *BeamGun) EnergyUsed() float64 {
    return 2.5
}

func (beam *BeamGun) IsEnabled() bool {
    return beam.enabled
}

func (beam *BeamGun) SetEnabled(enabled bool) {
    beam.enabled = enabled
}

func (beam *BeamGun) Rate() float64 {
    return 3.5
}

func (beam *BeamGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (beam *BeamGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, x float64, y float64, textFace *text.GoTextFace) {
    var pic *ebiten.Image
    animation, err := imageManager.LoadAnimation(gameImages.ImageBeam1)
    if err == nil {
        pic = animation.GetFrame(0)
    }

    drawGunBox(screen, x, y, iconColor(beam.enabled), pic)
}

func (beam *BeamGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if beam.enabled && beam.counter == 0 {
        beam.counter = int(60.0 / beam.Rate())
        velocityY := -2.3

        animation, err := imageManager.LoadAnimation(gameImages.ImageBeam1)
        if err != nil {
            return nil, err
        }

        bullet := Bullet{
            x: x,
            y: y,
            Strength: 2,
            health: 3,
            velocityX: 0,
            velocityY: velocityY,
            animation: animation,
            Gun: beam,
            // pic: pic,
        }

        return []*Bullet{&bullet}, nil
    } else {
        return nil, nil
    }
}

type MissleGun struct {
    enabled bool
    counter int
    level int
    experience float64
}

func (missle *MissleGun) LevelPercent() float64 {
    required := experienceForLevel(missle.level)
    if required == 0 {
        return 0.0
    }
    return missle.experience / required
}

func (missle *MissleGun) IncreaseExperience(experience float64) {
    // TODO
}

func (missle *MissleGun) Update() {
    if missle.counter > 0 {
        missle.counter -= 1
    }
}

func (missle *MissleGun) EnergyUsed() float64 {
    return 5
}

func (missle *MissleGun) IsEnabled() bool {
    return missle.enabled
}

func (missle *MissleGun) SetEnabled(enabled bool) {
    missle.enabled = enabled
}

func (missle *MissleGun) Rate() float64 {
    return 2
}

func (missle *MissleGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (missle *MissleGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, x float64, y float64, textFace *text.GoTextFace) {
    pic, _, err := imageManager.LoadImage(gameImages.ImageMissle1)
    if err != nil {
        pic = nil
    }
    drawGunBox(screen, x, y, iconColor(missle.enabled), pic)
}

func (missle *MissleGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if missle.enabled && missle.counter == 0 {
        missle.counter = int(60.0 / missle.Rate())
        velocityY := -2.1

        pic, _, err := imageManager.LoadImage(gameImages.ImageMissle1)
        if err != nil {
            return nil, err
        }

        bullet := Bullet{
            x: x,
            y: y,
            Strength: 10,
            health: 1,
            velocityX: 0,
            velocityY: velocityY,
            pic: pic,
            Gun: missle,
        }

        return []*Bullet{&bullet}, nil
    } else {
        return nil, nil
    }
}

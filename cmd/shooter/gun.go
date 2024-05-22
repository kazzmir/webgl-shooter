package main

import (
    gameImages "github.com/kazzmir/webgl-shooter/images"
    audioFiles "github.com/kazzmir/webgl-shooter/audio"

    "strconv"
    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Gun interface {
    Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error)
    Rate() float64
    DoSound(soundManager *SoundManager)
    DrawIcon(screen *ebiten.Image, imageManager *ImageManager, font *text.GoTextFaceSource, x float64, y float64)
    IsEnabled() bool
    SetEnabled(bool)
    Update()
}

type BasicGun struct {
    enabled bool
    counter int
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
    return 10
}

func drawGunBox(screen *ebiten.Image, x float64, y float64, n int, font *text.GoTextFaceSource) {
    size := 20
    vector.StrokeRect(screen, float32(x), float32(y), float32(size), float32(size), 2, &color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)

    face := &text.GoTextFace{Source: font, Size: 10}
    op := &text.DrawOptions{}
    op.GeoM.Translate(x + 5, y + 8)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, strconv.Itoa(n), face, op)
}

func (basic *BasicGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, font *text.GoTextFaceSource, x float64, y float64) {
    drawGunBox(screen, x, y, 1, font)
}

func (basic *BasicGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (basic *BasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if basic.enabled && basic.counter == 0 {
        basic.counter = int(60.0 / basic.Rate())
        velocityY := -2.5

        pic, _, err := imageManager.LoadImage(gameImages.ImageBullet)
        if err != nil {
            return nil, err
        }

        bullet := Bullet{
            x: x,
            y: y,
            Strength: 1,
            alive: true,
            velocityX: 0,
            velocityY: velocityY,
            pic: pic,
        }

        return []*Bullet{&bullet}, nil
    } else {
        return nil, nil
    }
}

type DualBasicGun struct {
    enabled bool
    counter int
}

func (dual *DualBasicGun) Update() {
    if dual.counter > 0 {
        dual.counter -= 1
    }
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

func (dual *DualBasicGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, font *text.GoTextFaceSource, x float64, y float64) {
    drawGunBox(screen, x, y, 2, font)
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
            alive: true,
            velocityX: 0,
            velocityY: velocityY,
            pic: pic,
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
}

func (beam *BeamGun) Update() {
    if beam.counter > 0 {
        beam.counter -= 1
    }
}

func (beam *BeamGun) IsEnabled() bool {
    return beam.enabled
}

func (beam *BeamGun) SetEnabled(enabled bool) {
    beam.enabled = enabled
}

func (beam *BeamGun) Rate() float64 {
    return 4
}

func (beam *BeamGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (beam *BeamGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, font *text.GoTextFaceSource, x float64, y float64) {
    drawGunBox(screen, x, y, 3, font)
}

func (beam *BeamGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    if beam.enabled && beam.counter > 0 {
        beam.counter = int(60.0 / beam.Rate())
        velocityY := -2.3

        animation, err := imageManager.LoadAnimation(gameImages.ImageBeam1)
        // animation, err := imageManager.LoadAnimation(gameImages.ImageRotate1)
        if err != nil {
            return nil, err
        }

        bullet := Bullet{
            x: x,
            y: y,
            Strength: 2,
            alive: true,
            velocityX: 0,
            velocityY: velocityY,
            animation: animation,
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
}

func (missle *MissleGun) Update() {
    if missle.counter > 0 {
        missle.counter -= 1
    }
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

func (missle *MissleGun) DrawIcon(screen *ebiten.Image, imageManager *ImageManager, font *text.GoTextFaceSource, x float64, y float64) {
    drawGunBox(screen, x, y, 4, font)
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
            alive: true,
            velocityX: 0,
            velocityY: velocityY,
            pic: pic,
        }

        return []*Bullet{&bullet}, nil
    } else {
        return nil, nil
    }
}

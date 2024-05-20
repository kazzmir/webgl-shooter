package main

import (
    gameImages "github.com/kazzmir/webgl-shooter/images"
    audioFiles "github.com/kazzmir/webgl-shooter/audio"
)

type Gun interface {
    Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error)
    Rate() float64
    DoSound(soundManager *SoundManager)
}

type BasicGun struct {
}

func (basic *BasicGun) Rate() float64 {
    return 10
}

func (basic *BasicGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (basic *BasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
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
}

type DualBasicGun struct {
}

func (dual *DualBasicGun) Rate() float64 {
    return 10
}

func (dual *DualBasicGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (dual *DualBasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
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
}

type BeamGun struct {
}

func (beam *BeamGun) Rate() float64 {
    return 4
}

func (beam *BeamGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (beam *BeamGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
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
}

type MissleGun struct {
}

func (missle *MissleGun) Rate() float64 {
    return 2
}

func (missle *MissleGun) DoSound(soundManager *SoundManager) {
    soundManager.Play(audioFiles.AudioShoot1)
}

func (missle *MissleGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
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
}

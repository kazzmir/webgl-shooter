package main

import (
    gameImages "github.com/kazzmir/webgl-shooter/images"
)

type Gun interface {
    Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error)
}

type BasicGun struct {
}

func (basic *BasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    velocityY := -2.5

    pic, err := imageManager.LoadImage(gameImages.ImageBullet)
    if err != nil {
        return nil, err
    }

    bullet := Bullet{
        x: x,
        y: y,
        alive: true,
        velocityX: 0,
        velocityY: velocityY,
        pic: pic,
    }

    return []*Bullet{&bullet}, nil
}

type DualBasicGun struct {
}

func (dual *DualBasicGun) Shoot(imageManager *ImageManager, x float64, y float64) ([]*Bullet, error) {
    velocityY := -2.5

    pic, err := imageManager.LoadImage(gameImages.ImageBullet)
    if err != nil {
        return nil, err
    }

    bullet1 := Bullet{
        x: x - 10,
        y: y,
        alive: true,
        velocityX: 0,
        velocityY: velocityY,
        pic: pic,
    }

    bullet2 := bullet1
    bullet2.x += 20

    return []*Bullet{&bullet1, &bullet2}, nil
}

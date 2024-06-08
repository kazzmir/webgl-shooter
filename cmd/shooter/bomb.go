package main

import (
    "image/color"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const MaxRadius float64 = 150

type Bomb struct {
    x, y float64
    velocityX float64
    velocityY float64
    // how much time remains until the bomb destructs
    destructTime int
    strength int
    radius float64
    alpha int
}

func MakeBomb(x float64, y float64, velocityX float64, velocityY float64) *Bomb {
    return &Bomb{
        x: x,
        y: y,
        velocityX: velocityX,
        velocityY: velocityY,
        destructTime: 100,
        strength: 30,
        radius: 1,
        alpha: 210,
    }
}

func (bomb *Bomb) IsAlive() bool {
    return bomb.alpha > 0
}

func (bomb *Bomb) Update() {

    if bomb.destructTime > 0 {
        bomb.destructTime -= 1
        bomb.x += bomb.velocityX
        bomb.y += bomb.velocityY
    } else {
        if bomb.radius < MaxRadius {
            bomb.radius += 7
        } else {
            if bomb.alpha > 0 {
                bomb.alpha -= 3
            }
        }
    }
}

func (bomb *Bomb) ShouldExplode() bool {
    return bomb.destructTime == 0
}

func (bomb *Bomb) Draw(screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager){
    if bomb.ShouldExplode() {

        var alpha uint8 = 0
        if bomb.alpha > 0 {
            alpha = uint8(bomb.alpha)
        }
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius), premultiplyAlpha(color.RGBA{R: 255, G: 0, B: 0, A: alpha}), true)
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius/2), premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 0, A: alpha}), true)
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius/4), premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: alpha}), true)
    } else {
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), 15.0, color.RGBA{R: 255, G: 0, B: 0, A: 0}, true)
    }
}

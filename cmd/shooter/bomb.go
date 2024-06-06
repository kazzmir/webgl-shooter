package main

import (
    "github.com/hajimehoshi/ebiten/v2"
)

type Bomb struct {
    x, y float64
    velocityX float64
    velocityY float64
    // how much time remains until the bomb destructs
    destructTime int
    strength int
}

func MakeBomb(x float64, y float64, velocityX float64, velocityY float64) *Bomb {
    return &Bomb{
        x: x,
        y: y,
        velocityX: velocityX,
        velocityY: velocityY,
        destructTime: 100,
        strength: 30,
    }
}

func (bomb *Bomb) Update() {
    bomb.x += bomb.velocityX
    bomb.y += bomb.velocityY
    if bomb.destructTime > 0 {
        bomb.destructTime -= 1
    }
}

func (bomb *Bomb) ShouldExplode() bool {
    return bomb.destructTime == 0
}

func (bomb *Bomb) Draw(screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager){
}

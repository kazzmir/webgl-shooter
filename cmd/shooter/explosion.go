package main

import (
    "math"
    "github.com/hajimehoshi/ebiten/v2"
)

type Explosion interface {
    Move()
    IsAlive() bool
    Draw(shaderManager *ShaderManager, screen *ebiten.Image)
}

type SimpleExplosion struct {
    x, y float64
    velocityX, velocityY float64
    pic *ebiten.Image
    life int
}

func MakeSimpleExplosion(x float64, y float64, pic *ebiten.Image) Explosion {
    return &SimpleExplosion{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 0,
        pic: pic,
        life: 10,
    }
}

func (explosion *SimpleExplosion) Move() {
    explosion.x += explosion.velocityX
    explosion.y += explosion.velocityY
    explosion.life -= 1
}

func (explosion *SimpleExplosion) IsAlive() bool {
    return explosion.life > 0
}

func (explosion *SimpleExplosion) Draw(shaderManager *ShaderManager, screen *ebiten.Image) {
    bounds := explosion.pic.Bounds()
    posX := explosion.x - float64(bounds.Dx()) / 2
    posY := explosion.y - float64(bounds.Dy()) / 2

    /*
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(posX, posY)
    screen.DrawImage(explosion.pic, options)
    */

    options := &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(posX, posY)
    options.Blend = AlphaBlender
    options.Images[0] = explosion.pic
    options.Uniforms = make(map[string]interface{})
    // radians = math.Pi * 90 / 180
    // log.Printf("Red: %v", radians)
    options.Uniforms["Center"] = []float32{float32(explosion.x), float32(explosion.y)}
    // options.Uniforms["Center"] = []float32{float32(bounds.Dx()) / 2, float32(bounds.Dy()) / 2}
    options.Uniforms["InnerRadius"] = float32(math.Max(0, float64(5-explosion.life)))
    options.Uniforms["OuterRadius"] = float32(math.Max(0, float64(10-explosion.life)))

    // log.Printf("Uniforms: %v", options.Uniforms)
    // options.Uniforms["InnerRadius"] = float32(10)
    // options.Uniforms["OuterRadius"] = float32(100)
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaderManager.ExplosionShader, options)
}

type AnimatedExplosion struct {
    x, y float64
    velocityX, velocityY float64
    animation *Animation
}

func (explosion *AnimatedExplosion) Move() {
    explosion.x += explosion.velocityX
    explosion.y += explosion.velocityY
    explosion.animation.Update()
}

func (explosion *AnimatedExplosion) IsAlive() bool {
    return explosion.animation.IsAlive()
}

func (explosion *AnimatedExplosion) Draw(shaderManager *ShaderManager, screen *ebiten.Image) {
    explosion.animation.Draw(screen, explosion.x, explosion.y)
}


func MakeAnimatedExplosion(x float64, y float64, animation *Animation) Explosion {
    return &AnimatedExplosion{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 0,
        animation: animation,
    }
}

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
        alpha: 240,
    }
}

func (bomb *Bomb) IsAlive() bool {
    return bomb.alpha > 0
}

func (bomb *Bomb) Update(onExplode func()) {

    if bomb.destructTime > 0 {
        bomb.destructTime -= 1
        bomb.x += bomb.velocityX
        bomb.y += bomb.velocityY
        if bomb.destructTime == 0 {
            onExplode()
        }
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
        /*
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius), premultiplyAlpha(color.RGBA{R: 255, G: 0, B: 0, A: alpha}), true)
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius/2), premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 0, A: alpha}), true)
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), float32(bomb.radius/4), premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: alpha}), true)
        */

        options := &ebiten.DrawRectShaderOptions{}
        options.GeoM.Translate(float64(bomb.x - bomb.radius), float64(bomb.y - bomb.radius))
        options.Uniforms = make(map[string]interface{})
        options.Uniforms["Center"] = []float32{float32(bomb.x), float32(bomb.y)}
        options.Uniforms["Radius"] = float32(bomb.radius)
        options.Uniforms["CenterAlpha"] = float32(alpha) / 255.0
        options.Uniforms["EdgeAlpha"] = float32(alpha) / (255.0 * 4)
        options.Uniforms["Color"] = []float32{1, 0, 0}
        screen.DrawRectShader(int(bomb.radius * 2), int(bomb.radius * 2), shaderManager.AlphaCircleShader, options)

        options.GeoM.Translate(float64(bomb.radius/2), float64(bomb.radius/2))
        options.Uniforms["Radius"] = float32(bomb.radius/2)
        options.Uniforms["Color"] = []float32{1, 1, 0}
        screen.DrawRectShader(int(bomb.radius), int(bomb.radius), shaderManager.AlphaCircleShader, options)

        options.GeoM.Translate(float64(bomb.radius/4), float64(bomb.radius/4))
        options.Uniforms["Radius"] = float32(bomb.radius/4)
        options.Uniforms["Color"] = []float32{1, 1, 1}
        screen.DrawRectShader(int(bomb.radius/2), int(bomb.radius/2), shaderManager.AlphaCircleShader, options)
    } else {
        vector.DrawFilledCircle(screen, float32(bomb.x), float32(bomb.y), 15.0, color.RGBA{R: 255, G: 0, B: 0, A: 0}, true)
    }
}

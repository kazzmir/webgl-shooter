package main

import (
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/red.kage
var RedShaderData []byte

//go:embed shaders/shadow.kage
var ShadowShaderData []byte

//go:embed shaders/edge.kage
var EdgeShaderData []byte

//go:embed shaders/explosion.kage
var ExplosionShaderData []byte

func LoadRedShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(RedShaderData)
}

func LoadEdgeShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(EdgeShaderData)
}

func LoadShadowShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(ShadowShaderData)
}

func LoadExplosionShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(ExplosionShaderData)
}

package main

import (
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/red.kage
var RedShaderData []byte

func LoadRedShader() (*ebiten.Shader, error) {
    return ebiten.NewShader(RedShaderData)
}

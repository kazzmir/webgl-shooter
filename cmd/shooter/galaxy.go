package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func DrawGalaxy(screen *ebiten.Image, galaxyShader *ebiten.Shader, galaxyImage *ebiten.Image, useTime float32, tilt float32, x float32, y float32, scale float32) {
    bounds := galaxyImage.Bounds()
    screenBounds := screen.Bounds()

    galaxyX := x / float32(screenBounds.Dx())
    galaxyY := y / float32(screenBounds.Dy())

    op := &ebiten.DrawRectShaderOptions{
		Images: [4]*ebiten.Image{galaxyImage},
		Uniforms: map[string]any{
			"GalaxyCenter": []float32{galaxyX, galaxyY},
            "GalaxyScale":  scale,
			"Time":         useTime,
			"Tilt":         tilt,
		},
		Blend: ebiten.BlendLighter,
	}
	op.GeoM.Scale(float64(screenBounds.Dx())/float64(bounds.Dx()), float64(screenBounds.Dy())/float64(bounds.Dy()))

	screen.DrawRectShader(bounds.Dx(), bounds.Dy(), galaxyShader, op)
}

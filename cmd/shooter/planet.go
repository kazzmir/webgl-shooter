package main

import (
    _ "fmt"
    "image"
    "math/rand/v2"
    _ "image/png"
    _ "image/jpeg"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

type Vector3 struct {
    X float32
    Y float32
    Z float32
}

func DrawPlanet(screen *ebiten.Image, x float64, y float64, scale float64, axis Vector3, planetImage *ebiten.Image, cloudImage *ebiten.Image, timeSeconds float64, shader *ebiten.Shader) {
    bounds := planetImage.Bounds()
    w, h := bounds.Dx(), bounds.Dy()

    opts := &ebiten.DrawRectShaderOptions{}
    opts.GeoM.Translate(-float64(w) * 0.5, -float64(h) * 0.5)
    opts.GeoM.Scale(scale, scale)
    opts.GeoM.Translate(x, y)

    // fmt.Printf("Resolution: %v x %v\n", w, h)

    rotationSpeed := timeSeconds / 600.0

    opts.Uniforms = map[string]interface{}{
        "Rotation":       float32(rotationSpeed),
        "Axis": []float32{axis.X, axis.Y, axis.Z},
    }
    opts.Images[0] = planetImage

    screen.DrawRectShader(w, h, shader, opts)

    if cloudImage != nil {
        opts.Blend = ebiten.BlendLighter
        opts.Images[0] = cloudImage
        opts.ColorScale.ScaleAlpha(0.2)
        opts.Uniforms["Rotation"] = float32(rotationSpeed * 0.7)
        screen.DrawRectShader(w, h, shader, opts)
    }
}

func makeCloudImage(bounds image.Rectangle) *ebiten.Image {
    w, h := bounds.Dx(), bounds.Dy()
    img := ebiten.NewImage(w, h)
    // img.Fill(color.RGBA{B: 255, A: 255})

    clouds := []string{"cloud1.png", "cloud-a.png"}

    for _, cloudFile := range clouds {
        cloud, _, err := ebitenutil.NewImageFromFile(cloudFile)
        if err != nil {
            panic(err)
        }

        cloudBounds := cloud.Bounds()

        for range 10 {
            var opts ebiten.DrawImageOptions
            opts.GeoM.Translate(float64(w - cloudBounds.Dx()) * rand.Float64(), float64(h - cloudBounds.Dy()) * rand.Float64())
            img.DrawImage(cloud, &opts)
        }
    }

    return img
}

/*
func MakeGame() *Game {
    data, err := os.ReadFile("planet.kage")
    if err != nil {
        panic(err)
    }
    shader, err := ebiten.NewShader(data)
    if err != nil {
        panic(err)
    }

    planetImage, _, err := ebitenutil.NewImageFromFile("earth.jpg")
    if err != nil {
        panic(err)
    }

    cloudImage := makeCloudImage(planetImage.Bounds())

    axis := Vector3{
        X: float32(rand.Float64() - 0.5),
        Y: float32(rand.Float64() - 0.5),
        Z: float32(rand.Float64() - 0.5),
    }

    return &Game{
        Counter: 0,
        Shader: shader,
        Planet: planetImage,
        CurrentPlanet: Earth,
        Axis: axis,
        drawClouds: true,
        CloudImage: cloudImage,
        Scale: 0.5,
    }
}
*/

/*
func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.NRGBA{R: 32, G: 0, B: 0, A: 255})

    x := float64(screen.Bounds().Dx() / 2)
    y := float64(screen.Bounds().Dy() / 2)

    / *
    vector.StrokeLine(screen, float32(x), 0, float32(x), float32(screen.Bounds().Dy()), 1, color.White, false)
    vector.StrokeLine(screen, 0, float32(y), float32(screen.Bounds().Dx()), float32(y), 1, color.White, false)
    * /

    cloud := g.CloudImage
    if !g.drawClouds {
        cloud = nil
    }

    ebitenutil.DebugPrintAt(screen, "Left/Right: Change Planet", 0, 0)
    ebitenutil.DebugPrintAt(screen, "Space: Toggle clouds", 0, 20)
    ebitenutil.DebugPrintAt(screen, "Mouse wheel: zoom in/out", 0, 40)

    draw(screen, x, y, g.Scale, g.Axis, g.Planet, cloud, float64(g.Counter), g.Shader)
}
*/

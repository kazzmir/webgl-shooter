package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"

    "github.com/kazzmir/webgl-shooter/lib/resize"

    "os"
    "image"
    "image/color"
    _ "image/png"
    "log"
)

type Run struct {
    img *ebiten.Image
    // img2 *ebiten.Image
    blurred *ebiten.Image
}

func (run *Run) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    // log.Printf("FPS: %v TPS: %v", ebiten.ActualFPS(), ebiten.ActualTPS())

    return nil
}

func drawCenter(screen *ebiten.Image, img *ebiten.Image, x, y float64) {
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(x, y)
    options.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
    screen.DrawImage(img, options)
}

func (run *Run) Draw(screen *ebiten.Image) {

    screen.Fill(color.RGBA{R: 32, G: 32, B: 32, A: 255})

    drawCenter(screen, run.blurred, 100, 200)
    drawCenter(screen, run.img, 100, 200)
    /*
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(50, 50)
    screen.DrawImage(run.img, options)
    */

    drawCenter(screen, run.blurred, 100, 500)
    // drawCenter(screen, run.img2, 300, 200)
    /*
    options = &ebiten.DrawImageOptions{}
    options.GeoM.Translate(250, 50)
    screen.DrawImage(run.img2, options)
    */

}

func (run *Run) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return 800, 640
}

func loadPng(path string) (image.Image, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    defer file.Close()
    img, _, err := image.Decode(file)
    if err != nil {
        return nil, err
    }
    return img, nil
}

func findAverage(img image.Image, x int, y int, size int) int {
    total := 0
    count := 0
    for i := x - size; i <= x + size; i++ {
        for j := y - size; j <= y + size; j++ {
            if i < 0 || j < 0 || i >= img.Bounds().Dx() || j >= img.Bounds().Dy() {
                continue
            }
            _, _, _, a := img.At(i, j).RGBA()
            total += int(a)
            count++
        }
    }
    if total == 0 {
        return 0
    }
    return total / count / 255
}

func premultiplyAlpha(value color.RGBA) color.RGBA {
    a := float32(value.A) / 256.0

    return color.RGBA{
        R: uint8(float32(value.R) * a),
        G: uint8(float32(value.G) * a),
        B: uint8(float32(value.B) * a),
        A: value.A,
    }
}

func makeBlur(img image.Image, factor float64, blur int, blurColor color.Color) image.Image {
    blurred := resize.ResizeBy(img, factor, resize.Bilinear)
    out := image.NewRGBA(blurred.Bounds())

    r, g, b, _ := blurColor.RGBA()

    r = r >> 8
    g = g >> 8
    b = b >> 8
    if r > 255 {
        r = 255
    }
    if g > 255 {
        g = 255
    }
    if b > 255 {
        b = 255
    }

    for x := 0; x < blurred.Bounds().Dx(); x++ {
        for y := 0; y < blurred.Bounds().Dy(); y++ {
            /*
            _, _, _, a := img.At(x, y).RGBA()
            if a == 0 {
                out.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
            } else {
                */
                average := findAverage(blurred, x, y, blur)
                if average > 255 {
                    average = 255
                }
                out.Set(x, y, premultiplyAlpha(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(average)}))
                // out.Set(x, y, premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 0, A: uint8(32)}))
            // }
        }
    }
    return out
}

func main(){
    img, err := loadPng("images/player/player.png")
    if err != nil {
        log.Printf("Error loading png: %v", err)
        return
    }

    // img2 := resize.ResizeBy(img, 1.2, resize.Bilinear)

    blurred := makeBlur(img, 1.2, 3, color.RGBA{R: 255, G: 255, B: 0, A: 255})

    run := Run{
        img: ebiten.NewImageFromImage(img),
        // img2: ebiten.NewImageFromImage(img2),
        blurred: ebiten.NewImageFromImage(blurred),
    }

    ebiten.SetWindowSize(800, 600)
    ebiten.RunGame(&run)
}

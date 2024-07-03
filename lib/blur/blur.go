package blur

import (
    "github.com/kazzmir/webgl-shooter/lib/resize"

    "image"
    "image/color"
)

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
    a := float32(value.A) / 255.0

    return color.RGBA{
        R: uint8(float32(value.R) * a),
        G: uint8(float32(value.G) * a),
        B: uint8(float32(value.B) * a),
        A: value.A,
    }
}

func MakeBlur(img image.Image, factor float64, blur int, blurColor color.Color) image.Image {
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
            average := findAverage(blurred, x, y, blur)
            if average > 255 {
                average = 255
            }
            out.Set(x, y, premultiplyAlpha(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(average)}))
        }
    }

    return out
}

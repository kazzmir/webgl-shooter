package main

import (
    "os"
    "log"
    "image/png"
    "image"
    "math"
)

func loadPng(path string) (image.Image, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    out, _, err := image.Decode(file)
    return out, err
}

func copyImageCenter(source image.Image, destination *image.RGBA, x int, y int, width int, height int){
    bounds := source.Bounds()

    padX := (width - bounds.Dx()) / 2
    padY := (height - bounds.Dy()) / 2

    for i := 0; i < bounds.Dx(); i++ {
        for j := 0; j < bounds.Dy(); j++ {
            destination.Set(x + i + padX, y + j + padY, source.At(i, j))
        }
    }
}

func createSheet(imagePaths []string, out string) error {
    var images []image.Image
    for _, path := range imagePaths {
        img, err := loadPng(path)
        if err != nil {
            return err
        }
        images = append(images, img)
        log.Printf("Loaded image %s", path)
    }

    maxWidth := 0
    maxHeight := 0

    for _, img := range images {
        bounds := img.Bounds()
        if bounds.Dx() > maxWidth {
            maxWidth = bounds.Dx()
        }
        if bounds.Dy() > maxHeight {
            maxHeight = bounds.Dy()
        }
    }

    log.Printf("Max width: %d, max height: %d", maxWidth, maxHeight)

    rows := int(math.Sqrt(float64(len(images))))
    columns := int(math.Ceil(float64(len(images)) / float64(rows)))

    log.Printf("Images: %d Rows: %d, columns: %d", len(images), rows, columns)

    sheetWidth := maxWidth * columns
    sheetHeight := maxHeight * rows

    sheet := image.NewRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))

    x := 0
    y := 0
    for _, img := range images {
        copyImageCenter(img, sheet, x, y, maxWidth, maxHeight)
        x += maxWidth
        if x >= sheetWidth {
            x = 0
            y += maxHeight
        }
    }

    file, err := os.Create(out)
    if err != nil {
        return err
    }

    err = png.Encode(file, sheet)
    if err != nil {
        return err
    }

    log.Printf("Created %s", out)
    return nil
}

func main(){
    out := "sheet.png"

    images := os.Args[1:]

    if len(images) == 0 {
        log.Printf("No images provided")
        return
    }

    err := createSheet(images, out)
    if err != nil {
        log.Printf("Error creating sheet: %s", err)
    }
}

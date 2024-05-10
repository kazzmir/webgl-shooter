package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "image"
)

type Animation struct {
    Frames []*ebiten.Image
    CurrentFrame int
    Loop bool
    FPS float64
    fpsCounter float64
}

func NewAnimation(sheet *ebiten.Image, frameRows int, frameColumns int, fps float64) *Animation {
    var frames []*ebiten.Image

    yMax := float64(sheet.Bounds().Dy())
    xMax := float64(sheet.Bounds().Dx())

    frameHeight := yMax / float64(frameRows)
    frameWidth := xMax / float64(frameColumns)

    for y := float64(0); y < yMax; y += frameHeight {
        for x := float64(0); x < xMax; x += frameWidth {
            frames = append(frames, sheet.SubImage(image.Rect(int(x), int(y), int(x + frameWidth), int(y + frameHeight))).(*ebiten.Image))
        }
    }
    return &Animation{
        Frames: frames,
        CurrentFrame: 0,
        Loop: false,
        FPS: fps,
        fpsCounter: 0,
    }
}

func (animation *Animation) IsAlive() bool {
    return animation.CurrentFrame < len(animation.Frames)
}

func (animation *Animation) Update() {
    animation.fpsCounter += 1.0
    if animation.fpsCounter < animation.FPS {
        return
    }

    animation.fpsCounter -= animation.FPS

    if animation.CurrentFrame < len(animation.Frames) {
        animation.CurrentFrame += 1
    }

    if animation.Loop && animation.CurrentFrame >= len(animation.Frames) {
        animation.CurrentFrame = 0
    }
}

func (animation *Animation) Draw(screen *ebiten.Image, x float64, y float64) {
    if animation.CurrentFrame >= len(animation.Frames) {
        return
    }

    frame := animation.Frames[animation.CurrentFrame]

    options := ebiten.DrawImageOptions{}
    options.GeoM.Translate(x - float64(frame.Bounds().Dx()) / 2.0, y - float64(frame.Bounds().Dy()) / 2.0)
    screen.DrawImage(frame, &options)
}

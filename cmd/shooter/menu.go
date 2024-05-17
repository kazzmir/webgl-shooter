package main

import (
    "math"
    "image/color"

    fontLib "github.com/kazzmir/webgl-shooter/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/audio"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type MenuAction func(run *Run) error

type MenuOption struct {
    Text string
    Action MenuAction
}

type Menu struct {
    Font *text.GoTextFaceSource
    Counter uint64
    Options []*MenuOption
    Selected int
}

func (menu *Menu) Update(run *Run) error {
    menu.Counter = (menu.Counter + 1)

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock: return ebiten.Termination
            case ebiten.KeyArrowUp:
                menu.Selected -= 1
                if menu.Selected < 0 {
                    menu.Selected = len(menu.Options) - 1
                }
            case ebiten.KeyArrowDown:
                menu.Selected = (menu.Selected + 1) % len(menu.Options)
            case ebiten.KeyEnter:
                err := menu.Options[menu.Selected].Action(run)
                if err != nil {
                    return err
                }
        }
    }

    return nil
}

func (menu *Menu) Draw(screen *ebiten.Image) {
    // screen.Fill(color.RGBA{0, 0, 0, 0xff})

    var x float64 = 200
    var y float64 = 200

    angle := float64(menu.Counter % 360) * math.Pi / 180.0 * 9
    // log.Printf("Counter: %v angle: %v", menu.Counter % 360, angle)

    // angle = 90.0 * math.Pi / 180.0
    a := int((math.Sin(angle) + 1) * 128)
    if a > 255 {
        a = 255
    }

    // vector.DrawFilledRect(screen, float32(x - 10), float32(y - 10), 100, 30, &color.RGBA{R: 255, G: 255, B: 255, A: uint8(menu.Counter % 255)}, true)

    face := text.GoTextFace{Source: menu.Font, Size: 20}

    _, height := text.Measure("X", &face, 0)

    // FIXME: compute this
    optionWidth := 150

    for i, option := range menu.Options {
        drawColor := color.RGBA{R: 255, G: 255, B: 255, A: 32}
        if menu.Selected == i {
            drawColor = color.RGBA{R: 255, G: 255, B: 255, A: uint8(a)}
        }
        vector.DrawFilledRect(screen, float32(x - 10), float32(y - 10), float32(optionWidth), float32(height + 10 + 10), premultiplyAlpha(drawColor), true)

        op := &text.DrawOptions{}
        op.GeoM.Translate(x, y)
        red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
        op.ColorScale.ScaleWithColor(red)
        text.Draw(screen, option.Text, &face, op)

        y += float64(height + 40)
    }
}

func createMenu(audioContext *audio.Context) (*Menu, error) {

    var options []*MenuOption

    options = append(options, &MenuOption{
        Text: "New game",
        Action: func(run *Run) error {
            run.Mode = RunGame

            if run.Game != nil {
                run.Game.Cancel()
            }

            game, err := MakeGame(audioContext)
            if err != nil {
                return err
            }

            run.Game = game

            return nil
        },

    })

    options = append(options, &MenuOption{
        Text: "Continue",
        Action: func(run *Run) error {
            run.Mode = RunGame

            if run.Game == nil {
                game, err := MakeGame(audioContext)
                if err != nil {
                    return err
                }

                run.Game = game
            }

            return nil
        },
    })

    options = append(options, &MenuOption{
        Text: "Quit",
        Action: func(run *Run) error {
            return ebiten.Termination
        },
    })

    font, err := fontLib.LoadFont()
    if err != nil {
        return nil, err
    }

    return &Menu{
        Font: font,
        Options: options,
    }, nil
}

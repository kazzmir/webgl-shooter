package main

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const peerRoomIDMaxLength = 100

type PeerEditor struct {
	Active bool
	Title  string
	Value  string
	Apply  func(string)
	MaxLength int
}

func (editor *PeerEditor) Handle(chars []rune, keys []ebiten.Key) {
	if !editor.Active {
		return
	}

	controlPressed := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)

	for _, key := range keys {
		if controlPressed {
			switch key {
			case ebiten.KeyU:
				editor.Value = ""
				continue
			case ebiten.KeyW:
				editor.deleteLastWord()
				continue
			}
		}

		switch key {
		case ebiten.KeyEscape:
			editor.Active = false
			return
		case ebiten.KeyEnter:
			editor.Active = false
			if editor.Apply != nil {
				editor.Apply(strings.TrimSpace(editor.Value))
			}
			return
		case ebiten.KeyBackspace:
			if editor.Value != "" {
				_, size := utf8.DecodeLastRuneInString(editor.Value)
				editor.Value = editor.Value[:len(editor.Value)-size]
			}
		}
	}

	for _, char := range chars {
		if char < 32 || char == 127 {
			continue
		}
		if editor.MaxLength > 0 && utf8.RuneCountInString(editor.Value) >= editor.MaxLength {
			break
		}
		editor.Value += string(char)
	}

	backspaceDuration := inpututil.KeyPressDuration(ebiten.KeyBackspace)
	if backspaceDuration >= 20 && backspaceDuration%3 == 0 {
		editor.deleteLastRune()
	}
}

func (editor *PeerEditor) Draw(screen *ebiten.Image, font *text.GoTextFaceSource, counter uint64) {
	if !editor.Active {
		return
	}

	vector.DrawFilledRect(screen, 150, 180, 900, 220, color.RGBA{R: 10, G: 18, B: 28, A: 240}, true)
	vector.StrokeRect(screen, 150, 180, 900, 220, 2, color.RGBA{R: 220, G: 230, B: 255, A: 255}, true)

	titleFace := text.GoTextFace{Source: font, Size: 22}
	bodyFace := text.GoTextFace{Source: font, Size: 16}
	valueFace := text.GoTextFace{Source: font, Size: 18}

	drawText(screen, titleFace, 180, 220, editor.Title, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	drawText(screen, bodyFace, 180, 255, "Type a value, press Enter to save, or Escape to cancel.", color.RGBA{R: 190, G: 210, B: 255, A: 255})
	drawText(screen, bodyFace, 180, 280, "Ctrl+U clears the line. Ctrl+W deletes the last word.", color.RGBA{R: 190, G: 210, B: 255, A: 255})
	if editor.MaxLength > 0 {
		drawText(screen, bodyFace, 180, 305, fmt.Sprintf("Maximum length: %d characters.", editor.MaxLength), color.RGBA{R: 190, G: 210, B: 255, A: 255})
	}

	value := editor.Value
	if (counter/30)%2 == 0 {
		value += "_"
	}
	drawText(screen, valueFace, 180, 340, value, color.RGBA{R: 255, G: 120, B: 120, A: 255})
}

func (menu *Menu) openPeerEditor(title string, value string, apply func(string)) {
	menu.PeerEditor = &PeerEditor{
		Active: true,
		Title:  title,
		Value:  value,
		Apply:  apply,
	}
}

func (menu *Menu) openPeerRoomEditor(value string, apply func(string)) {
	menu.PeerEditor = &PeerEditor{
		Active: true,
		Title:  "Peer room ID",
		Value:  value,
		Apply:  apply,
		MaxLength: peerRoomIDMaxLength,
	}
}

func (editor *PeerEditor) deleteLastRune() {
	if editor.Value == "" {
		return
	}

	_, size := utf8.DecodeLastRuneInString(editor.Value)
	editor.Value = editor.Value[:len(editor.Value)-size]
}

func (editor *PeerEditor) deleteLastWord() {
	for strings.HasSuffix(editor.Value, " ") {
		editor.deleteLastRune()
	}

	for editor.Value != "" && !strings.HasSuffix(editor.Value, " ") {
		editor.deleteLastRune()
	}
}

func (menu *Menu) peerServerLabel() string {
	value := menu.PeerConnector.ServerURL()
	if value == "" {
		value = "<unset>"
	}
	return fmt.Sprintf("Peer server: %s", value)
}

func (menu *Menu) peerRoomLabel() string {
	value := menu.PeerConnector.RoomID()
	if value == "" {
		value = "<unset>"
	}
	return fmt.Sprintf("Peer room: %s", value)
}

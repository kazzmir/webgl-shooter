package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"sync"

	audioFiles "github.com/kazzmir/webgl-shooter/audio"
	fontLib "github.com/kazzmir/webgl-shooter/font"
	gameImages "github.com/kazzmir/webgl-shooter/images"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func drawText(screen *ebiten.Image, face text.GoTextFace, x, y float64, str string, color color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color)
	text.Draw(screen, str, &face, op)
}

type MenuAction func(self *MenuOption, run *Run, key ebiten.Key) error

type MenuOption struct {
	Text     string
	TextFunc func() string
	Action   MenuAction
	Respond  []ebiten.Key
}

type HintFunc func(*Hint, *ebiten.Image, *ImageManager, *ShaderManager, *text.GoTextFaceSource, ebiten.GeoM) error

type Hint struct {
	Active  bool
	Time    int
	FadeIn  int
	FadeOut int
	Show    HintFunc
}

func (hint *Hint) Activate() {
	hint.Active = true
	hint.Time = 250
	hint.FadeIn = 50
}

func (hint *Hint) FontColor() color.Color {
	if hint.FadeIn > 0 {
		maxFade := 50
		alpha := (maxFade - hint.FadeIn) * 255 / maxFade
		if alpha > 255 {
			alpha = 255
		}
		return premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: uint8(alpha)})
	}

	if hint.Time < 50 {
		maxFade := 50
		alpha := hint.Time * 255 / maxFade
		if alpha > 255 {
			alpha = 255
		}
		return premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: uint8(alpha)})
	}

	return premultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

func (hint *Hint) Update() {

	if hint.Time > 0 {
		hint.Time -= 1
	}

	if hint.FadeIn > 0 {
		hint.FadeIn -= 1
	}

	if hint.Time == 0 {
		hint.Active = false
	}
}

type Menu struct {
	Font          *text.GoTextFaceSource
	Counter       uint64
	Options       []*MenuOption
	MultiplayerOptions []*MenuOption
	MultiplayerStartOption *MenuOption
	Selected      int
	MultiplayerSelected int
	MultiplayerOpen bool
	SoundManager  *SoundManager
	ImageManager  *ImageManager
	ShaderManager *ShaderManager
	PeerConnector PeerConnector
	PeerEditor    *PeerEditor

	Hints      []*Hint
	ActiveHint int
}

func (option *MenuOption) Label() string {
	if option.TextFunc != nil {
		return option.TextFunc()
	}

	return option.Text
}

func (option *MenuOption) DoesRespond(key ebiten.Key) bool {
	for _, respond := range option.Respond {
		if respond == key {
			return true
		}
	}

	return false
}

func (menu *Menu) currentOptions() []*MenuOption {
	if menu.MultiplayerOpen {
		if menu.PeerConnector != nil && menu.PeerConnector.IsConnected() && menu.MultiplayerStartOption != nil {
			options := make([]*MenuOption, 0, len(menu.MultiplayerOptions)+1)
			last := len(menu.MultiplayerOptions) - 1
			if last < 0 {
				return menu.MultiplayerOptions
			}
			options = append(options, menu.MultiplayerOptions[:last]...)
			options = append(options, menu.MultiplayerStartOption)
			options = append(options, menu.MultiplayerOptions[last])
			return options
		}

		return menu.MultiplayerOptions
	}

	return menu.Options
}

func (menu *Menu) currentSelected() *int {
	if menu.MultiplayerOpen {
		return &menu.MultiplayerSelected
	}

	return &menu.Selected
}

func (menu *Menu) ChooseHint() {
	if len(menu.Hints) == 1 {
		menu.ActiveHint = 0
	} else {
		for _, i := range rand.Perm(len(menu.Hints)) {
			if i != menu.ActiveHint {
				menu.ActiveHint = i
				break
			}
		}
	}

	menu.Hints[menu.ActiveHint].Activate()
}

func (menu *Menu) Update(run *Run) error {
	menu.Counter = (menu.Counter + 1)

	keys := make([]ebiten.Key, 0)
	keys = inpututil.AppendJustPressedKeys(keys)
	chars := make([]rune, 0)
	chars = ebiten.AppendInputChars(chars)

	if menu.PeerEditor != nil && menu.PeerEditor.Active {
		menu.PeerEditor.Handle(chars, keys)
		return nil
	}

	if menu.ActiveHint == -1 || menu.Hints[menu.ActiveHint].Active == false {
		menu.ChooseHint()
	}

	menu.Hints[menu.ActiveHint].Update()

	for _, key := range keys {
		options := menu.currentOptions()
		selected := menu.currentSelected()
		if *selected >= len(options) {
			*selected = len(options) - 1
		}
		switch key {
		case ebiten.KeyEscape, ebiten.KeyCapsLock:
			if menu.MultiplayerOpen {
				menu.MultiplayerOpen = false
				return nil
			}
			if run.Game != nil {
				run.Mode = RunGame
			}
		case ebiten.KeyArrowUp:
			*selected -= 1
			if *selected < 0 {
				*selected = len(options) - 1
			}
			menu.SoundManager.Play(audioFiles.AudioBeep)
		case ebiten.KeyArrowDown:
			*selected = (*selected + 1) % len(options)
			menu.SoundManager.Play(audioFiles.AudioBeep)
		default:
			option := options[*selected]
			if option.DoesRespond(key) {
				menu.SoundManager.Play(audioFiles.AudioBeep)
				err := option.Action(option, run, key)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (menu *Menu) Draw(screen *ebiten.Image) {
	// screen.Fill(color.RGBA{0, 0, 0, 0xff})

	var x float64 = 100
	var y float64 = 100

	angle := float64(menu.Counter%360) * math.Pi / 180.0 * 9
	// log.Printf("Counter: %v angle: %v", menu.Counter % 360, angle)

	// angle = 90.0 * math.Pi / 180.0
	a := int((math.Sin(angle) + 1) * 128)
	if a > 255 {
		a = 255
	}

	// vector.DrawFilledRect(screen, float32(x - 10), float32(y - 10), 100, 30, &color.RGBA{R: 255, G: 255, B: 255, A: uint8(menu.Counter % 255)}, true)

	face := text.GoTextFace{Source: menu.Font, Size: 20}

	_, height := text.Measure("X", &face, 0)

	options := menu.currentOptions()
	selected := *menu.currentSelected()
	if selected >= len(options) {
		selected = len(options) - 1
	}

	optionWidth := 150.0
	for _, option := range options {
		width, _ := text.Measure(option.Label(), &face, 0)
		optionWidth = math.Max(optionWidth, width+20)
	}

	for i, option := range options {
		label := option.Label()
		drawColor := color.RGBA{R: 255, G: 255, B: 255, A: 32}
		if selected == i {
			drawColor = color.RGBA{R: 255, G: 255, B: 255, A: uint8(a)}
		}
		vector.FillRect(screen, float32(x-10), float32(y-10), float32(optionWidth), float32(height+10+10), premultiplyAlpha(drawColor), true)

		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		op.ColorScale.ScaleWithColor(red)
		text.Draw(screen, label, &face, op)

		y += float64(height + 40)
	}

	if menu.MultiplayerOpen && menu.PeerConnector != nil {
		drawText(screen, text.GoTextFace{Source: menu.Font, Size: 28}, x, 60, "Multiplayer", color.RGBA{R: 255, G: 255, B: 255, A: 255})
		drawText(screen, text.GoTextFace{Source: menu.Font, Size: 14}, x, y, menu.PeerConnector.StatusLine(menu.Counter), color.RGBA{R: 200, G: 220, B: 255, A: 255})
	}

	hintX := 500
	hintY := 400
	hintWidth := 500
	hintHeight := 240
	hintArea := screen.SubImage(image.Rect(hintX, hintY, hintX+hintWidth, hintY+hintHeight)).(*ebiten.Image)
	hintArea.Fill(color.RGBA{0x11, 0x21, 0x32, 0xff})
	vector.StrokeRect(hintArea, float32(hintX), float32(hintY), float32(hintWidth), float32(hintHeight), 1, color.RGBA{0xff, 0xff, 0xff, 0xff}, true)

	if menu.ActiveHint != -1 && menu.Hints[menu.ActiveHint].Active {
		var geom ebiten.GeoM
		geom.Translate(float64(hintX+5), float64(hintY+3))
		hint := menu.Hints[menu.ActiveHint]
		err := hint.Show(hint, hintArea, menu.ImageManager, menu.ShaderManager, menu.Font, geom)
		if err != nil {
			log.Printf("Warning: error rendering hint %v: %v", menu.ActiveHint, err)
		}
	}

	drawText(screen, text.GoTextFace{Source: menu.Font, Size: 15}, ScreenWidth-170, ScreenHeight-20, "Made by Jon Rafkind", color.RGBA{R: 255, G: 255, B: 255, A: 255})

	if menu.PeerEditor != nil && menu.PeerEditor.Active {
		menu.PeerEditor.Draw(screen, menu.Font, menu.Counter)
	}
}

func makeHintKeys() *Hint {
	return &Hint{
		Active: false,
		Time:   0,
		Show: func(self *Hint, screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager, font *text.GoTextFaceSource, geometry ebiten.GeoM) error {

			var fontSize float64 = 15

			face := text.GoTextFace{Source: font, Size: fontSize}

			op := &text.DrawOptions{}
			op.GeoM = geometry
			fontColor := self.FontColor()
			op.ColorScale.ScaleWithColor(fontColor)
			text.Draw(screen, "Keys", &face, op)
			op.GeoM.Translate(5, 0)
			all := []string{
				"Arrow up: move ship up",
				"Arrow down: move ship down",
				"Arrow left: move ship left",
				"Arrow right: move ship right",
				"Spacebar: shoot",
				"Left Shift: increase speed",
				"B: release bomb",
			}
			for _, s := range all {
				op.GeoM.Translate(0, fontSize+1)
				text.Draw(screen, s, &face, op)
			}

			return nil
		},
	}
}

func makeHintHealth() *Hint {
	var healthImage *ebiten.Image
	var load sync.Once
	return &Hint{
		Active: false,
		Time:   0,
		Show: func(self *Hint, screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager, font *text.GoTextFaceSource, geometry ebiten.GeoM) error {
			load.Do(func() {
				var err error
				healthImage, _, err = imageManager.LoadImage(gameImages.ImageHealthBar)
				if err != nil {
					healthImage = nil
					log.Printf("Unable to create energy bar: %v", err)
				}
			})

			face := text.GoTextFace{Source: font, Size: 16}
			op := &text.DrawOptions{}
			op.GeoM = geometry
			fontColor := self.FontColor()
			op.ColorScale.ScaleWithColor(fontColor)
			text.Draw(screen, "Health", &face, op)

			if healthImage != nil {
				options := &ebiten.DrawImageOptions{}
				options.GeoM.Scale(0.5, 0.5)
				x, y := geometry.Apply(0, 0)
				options.GeoM.Translate(x, y)
				options.GeoM.Translate(10, 30)
				screen.DrawImage(healthImage, options)

				op.GeoM.Translate(40, 40)
				text.Draw(screen, "This is your health bar.", &face, op)
				op.GeoM.Translate(0, 20)
				text.Draw(screen, "Colliding with bullets or enemies will lower your health.", &face, op)
				op.GeoM.Translate(0, 20)
				text.Draw(screen, "When your health is depleted your ship will explode!", &face, op)
			}

			return nil
		},
	}
}

func makeHintEnergy() *Hint {
	var energyImage *ebiten.Image
	var load sync.Once
	return &Hint{
		Active: false,
		Time:   0,
		Show: func(self *Hint, screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager, font *text.GoTextFaceSource, geometry ebiten.GeoM) error {
			load.Do(func() {
				var err error
				energyImage, _, err = imageManager.LoadImage(gameImages.ImageEnergyBar)
				if err != nil {
					energyImage = nil
					log.Printf("Unable to create energy bar: %v", err)
				}
			})

			face := text.GoTextFace{Source: font, Size: 16}
			op := &text.DrawOptions{}
			op.GeoM = geometry
			fontColor := self.FontColor()
			op.ColorScale.ScaleWithColor(fontColor)
			text.Draw(screen, "Shooting requires energy. Energy is restored over time.", &face, op)

			if energyImage != nil {
				options := &ebiten.DrawImageOptions{}
				options.GeoM.Scale(0.5, 0.5)
				x, y := geometry.Apply(0, 0)
				options.GeoM.Translate(x, y)
				options.GeoM.Translate(10, 30)
				screen.DrawImage(energyImage, options)

				op.GeoM.Translate(40, 40)
				text.Draw(screen, "This is your energy bar.", &face, op)
				op.GeoM.Translate(0, 20)
				text.Draw(screen, "When it is depleted you will not be able to shoot.", &face, op)
			}

			return nil
		},
	}
}

func makeHintPowerups() *Hint {
	return &Hint{
		Active: false,
		Time:   0,
		Show: func(self *Hint, screen *ebiten.Image, imageManager *ImageManager, shaderManager *ShaderManager, font *text.GoTextFaceSource, geometry ebiten.GeoM) error {

			face := text.GoTextFace{Source: font, Size: 14}

			op := &text.DrawOptions{}
			op.GeoM = geometry
			fontColor := self.FontColor()
			op.ColorScale.ScaleWithColor(fontColor)
			text.Draw(screen, "Powerups", &face, op)

			x, y := geometry.Apply(0, 0)

			var scaler ebiten.GeoM
			scaler.Scale(0.7, 0.7)

			powerup := MakePowerupEnergy(x+30, y+40)
			powerup.Draw(screen, imageManager, shaderManager, scaler)

			op.GeoM.Translate(60, 27)
			text.Draw(screen, "Energy stays at the maximum for a few seconds", &face, op)

			powerup = MakePowerupHealth(x+30, y+80)
			powerup.Draw(screen, imageManager, shaderManager, scaler)

			op.GeoM.Translate(0, 40)
			text.Draw(screen, "Increase health by some amount", &face, op)

			powerup = MakePowerupWeapon(x+30, y+120)
			powerup.Draw(screen, imageManager, shaderManager, scaler)
			op.GeoM.Translate(0, 40)
			text.Draw(screen, "Enable the next weapon slot", &face, op)

			powerup = MakePowerupBomb(x+30, y+160)
			powerup.Draw(screen, imageManager, shaderManager, scaler)
			op.GeoM.Translate(0, 40)
			text.Draw(screen, "Add a bomb to your arsenal", &face, op)

			/*
			   powerup = MakePowerupEnergyIncrease(x + 30, y + 200)
			   powerup.Draw(screen, imageManager, shaderManager, scaler)
			   op.GeoM.Translate(0, 40)
			   text.Draw(screen, "Increase maximum energy and fill rate", &face, op)
			*/

			return nil
		},
	}
}

func createMenu(quit context.Context, soundManager *SoundManager, initialVolume float64, cheats bool, peerConnector PeerConnector) (*Menu, error) {

	var options []*MenuOption
	var multiplayerOptions []*MenuOption
	var multiplayerStartOption *MenuOption
	var menu *Menu

	startNewGame := func(run *Run) error {
		run.Mode = RunGame

		if run.Game != nil {
			run.Game.Cancel()
		}

		player, err := MakePlayer(0, 0, cheats)
		if err != nil {
			return err
		}
		run.Player = player

		game, err := MakeGame(soundManager, run, 1)
		if err != nil {
			return err
		}

		run.Game = game
		return nil
	}

	options = append(options, &MenuOption{
		Text: "New game",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			return startNewGame(run)
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	soundMuted := false
	lastVolume := initialVolume
	options = append(options, &MenuOption{
		Text: fmt.Sprintf("Sound %v", initialVolume),
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			switch key {
			case ebiten.KeyArrowLeft:
				if !soundMuted {
					run.DecreaseVolume()
					lastVolume = run.GetVolume()
				}
			case ebiten.KeyArrowRight:
				if !soundMuted {
					run.IncreaseVolume()
					lastVolume = run.GetVolume()
				}
			case ebiten.KeyEnter:
				soundMuted = !soundMuted
				if soundMuted {
					run.SetVolume(0)
				} else {
					run.SetVolume(lastVolume)
				}
			}

			if soundMuted {
				self.Text = "Sound Muted"
			} else {
				self.Text = fmt.Sprintf("Sound %v", run.GetVolume())
			}
			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyEnter},
	})

	options = append(options, &MenuOption{
		Text: "Fullscreen",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			ebiten.SetFullscreen(!ebiten.IsFullscreen())

			if ebiten.IsFullscreen() {
				self.Text = "Windowed"
			} else {
				self.Text = "Fullscreen"
			}

			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	options = append(options, &MenuOption{
		Text: "Continue",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			run.Mode = RunGame

			if run.Game == nil {
				if run.Player == nil {
					player, err := MakePlayer(0, 0, cheats)
					if err != nil {
						return err
					}
					run.Player = player
				}

				game, err := MakeGame(soundManager, run, 1)
				if err != nil {
					return err
				}

				run.Game = game
			}

			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	options = append(options, &MenuOption{
		Text: "Multiplayer",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			menu.MultiplayerOpen = true
			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	multiplayerOptions = append(multiplayerOptions, &MenuOption{
		TextFunc: func() string {
			return menu.peerServerLabel()
		},
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			menu.openPeerEditor("Peer signaling server URL", peerConnector.ServerURL(), peerConnector.SetServerURL)
			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	multiplayerOptions = append(multiplayerOptions, &MenuOption{
		TextFunc: func() string {
			return menu.peerRoomLabel()
		},
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			menu.openPeerRoomEditor(peerConnector.RoomID(), peerConnector.SetRoomID)
			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	multiplayerOptions = append(multiplayerOptions, &MenuOption{
		TextFunc: func() string {
			return peerConnector.MenuLabel()
		},
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			return peerConnector.Action()
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	multiplayerStartOption = &MenuOption{
		Text: "Start game",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			return startNewGame(run)
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	}

	multiplayerOptions = append(multiplayerOptions, &MenuOption{
		Text: "Back",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			menu.MultiplayerOpen = false
			return nil
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	options = append(options, &MenuOption{
		Text: "Quit",
		Action: func(self *MenuOption, run *Run, key ebiten.Key) error {
			return ebiten.Termination
		},
		Respond: []ebiten.Key{ebiten.KeyEnter},
	})

	font, err := fontLib.LoadFont()
	if err != nil {
		return nil, err
	}

	// FIXME: re-use game shader manager
	shaderManager, err := MakeShaderManager()
	if err != nil {
		return nil, err
	}

	hints := []*Hint{
		makeHintKeys(),
		makeHintPowerups(),
		makeHintEnergy(),
		makeHintHealth(),
	}

	menu = &Menu{
		Font:               font,
		Options:            options,
		MultiplayerOptions: multiplayerOptions,
		MultiplayerStartOption: multiplayerStartOption,
		ImageManager:       MakeImageManager(),
		ShaderManager:      shaderManager,
		PeerConnector:      peerConnector,
		SoundManager:       soundManager,
		Hints:              hints,
		ActiveHint:         -1,
	}

	return menu, nil
}

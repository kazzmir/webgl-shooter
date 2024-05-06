package audio

import (
    _ "embed"
    "fmt"
    "bytes"

    "github.com/hajimehoshi/ebiten/v2/audio/mp3"
    // libAudio "github.com/hajimehoshi/ebiten/v2/audio"
)

//go:embed effects/hit1.mp3
var Hit1Data []byte

//go:embed effects/shoot1.mp3
var Shoot1Data []byte

//go:embed music/stellar-pulse.mp3
var SongStellarPulseData []byte

type AudioName string

const AudioHit1 = AudioName("hit1")
const AudioShoot1 = AudioName("shoot1")
const AudioStellarPulseSong = AudioName("stellar-pulse")

var AllSounds []AudioName = []AudioName{AudioHit1, AudioShoot1, AudioStellarPulseSong}

func LoadSound(name AudioName, sampleRate int) (*mp3.Stream, error) {
    switch name {
        case AudioHit1: return mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(Hit1Data))
        case AudioShoot1: return mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(Shoot1Data))
        case AudioStellarPulseSong: return mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(SongStellarPulseData))
    }

    return nil, fmt.Errorf("No such audio effect %v", name)
}

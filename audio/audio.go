package audio

import (
    _ "embed"
    "fmt"
    "bytes"
    "io"

    // "github.com/hajimehoshi/ebiten/v2/audio/mp3"
    "github.com/hajimehoshi/ebiten/v2/audio/vorbis"
    // libAudio "github.com/hajimehoshi/ebiten/v2/audio"
)

//go:embed effects/hit1.ogg
var Hit1Data []byte

//go:embed effects/shoot1.ogg
var Shoot1Data []byte

//go:embed music/stellar-pulse.ogg
var SongStellarPulseData []byte

type AudioName string

const AudioHit1 = AudioName("hit1")
const AudioShoot1 = AudioName("shoot1")
const AudioStellarPulseSong = AudioName("stellar-pulse")

var AllSounds []AudioName = []AudioName{AudioHit1, AudioShoot1, AudioStellarPulseSong}

func LoadSound(name AudioName, sampleRate int) (io.Reader, error) {
    switch name {
        case AudioHit1: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Hit1Data))
        case AudioShoot1: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Shoot1Data))
        case AudioStellarPulseSong: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(SongStellarPulseData))
    }

    return nil, fmt.Errorf("No such audio effect %v", name)
}

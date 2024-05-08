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

//go:embed effects/explosion1.ogg
var Explosion1Data []byte

//go:embed effects/explosion2.ogg
var Explosion2Data []byte

//go:embed effects/explosion3.ogg
var Explosion3Data []byte

type AudioName string

const AudioHit1 = AudioName("hit1")
const AudioShoot1 = AudioName("shoot1")
const AudioStellarPulseSong = AudioName("stellar-pulse")
const AudioExplosion1 = AudioName("explosion1")
const AudioExplosion2 = AudioName("explosion2")
const AudioExplosion3 = AudioName("explosion3")

var AllSounds []AudioName = []AudioName{AudioHit1, AudioShoot1, AudioStellarPulseSong, AudioExplosion1, AudioExplosion2, AudioExplosion3}

func LoadSound(name AudioName, sampleRate int) (io.Reader, error) {
    switch name {
        case AudioHit1: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Hit1Data))
        case AudioShoot1: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Shoot1Data))
        case AudioStellarPulseSong: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(SongStellarPulseData))
        case AudioExplosion1: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Explosion1Data))
        case AudioExplosion2: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Explosion2Data))
        case AudioExplosion3: return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(Explosion3Data))
    }

    return nil, fmt.Errorf("No such audio effect %v", name)
}

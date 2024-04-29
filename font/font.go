package font

import (
    _ "embed"
    "bytes"

    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed futura.ttf
var FuturaData []byte

func LoadFont() (*text.GoTextFaceSource, error) {
    return text.NewGoTextFaceSource(bytes.NewReader(FuturaData))
}

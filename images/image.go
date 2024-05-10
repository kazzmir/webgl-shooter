package image

import (
    _ "embed"
    "image"
    "bytes"
    "fmt"
    _ "image/png"
)

type Image string

const ImagePlayer = Image("player")
const ImageStar1 = Image("star1")
const ImageStar2 = Image("star2")
const ImagePlanet = Image("planet")
const ImageBullet = Image("bullet")
const ImageEnemy1 = Image("enemy1")
const ImageEnemy2 = Image("enemy2")
const ImageExplosion1 = Image("explosion1")
const ImageExplosion2 = Image("explosion2")

//go:embed player/player.png
var playerImage []byte

//go:embed player/bullet.png
var bulletImage []byte

//go:embed background/star1.png
var star1Image []byte

//go:embed background/star2.png
var star2Image []byte

//go:embed background/planet1.png
var planetImage []byte

//go:embed enemy/enemy1.png
var enemy1Image []byte

//go:embed enemy/enemy2.png
var enemy2Image []byte

//go:embed misc/explosion1.png
var explosion1Image []byte

//go:embed misc/explosion2-anim.png
var explosion2Animation []byte

func loadPng(data []byte) (image.Image, error) {
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    return img, nil
}

func LoadImage(name Image) (image.Image, error) {
    switch name {
        case ImagePlayer: return loadPng(playerImage)
        case ImageStar1: return loadPng(star1Image)
        case ImageStar2: return loadPng(star2Image)
        case ImagePlanet: return loadPng(planetImage)
        case ImageBullet: return loadPng(bulletImage)
        case ImageEnemy1: return loadPng(enemy1Image)
        case ImageEnemy2: return loadPng(enemy2Image)
        case ImageExplosion1: return loadPng(explosion1Image)
        case ImageExplosion2: return loadPng(explosion2Animation)
    }

    return nil, fmt.Errorf("no such image: %s", name)
}

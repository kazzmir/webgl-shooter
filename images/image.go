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
const ImageEnemy3 = Image("enemy3")
const ImageEnemy4 = Image("enemy4")
const ImageEnemy5 = Image("enemy5")
const ImageAsteroid1 = Image("asteroid1")
const ImageEnergyBar = Image("energy")
const ImageHealthBar = Image("health")
const ImageBoss1 = Image("boss1")
const ImageExplosion1 = Image("explosion1")
const ImageExplosion2 = Image("explosion2")
const ImageExplosion3 = Image("explosion3")
const ImageHit = Image("hit")
const ImageHit2 = Image("hit2")
const ImageBeam1 = Image("beam1")
const ImageWave1 = Image("wave1")
const ImageRotate1 = Image("rotate1")
const ImageMissle1 = Image("missle1")
const ImageBulletSmallBlue = Image("bullet-small-blue")
const ImageFire1 = Image("fire1")
const ImagePowerup1 = Image("powerup1")
const ImagePowerup2 = Image("powerup2")
const ImagePowerup3 = Image("powerup3")
const ImagePowerup4 = Image("powerup4")

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

//go:embed bullet/small-blue.png
var bulletSmallBlueImage []byte

//go:embed enemy/enemy1.png
var enemy1Image []byte

//go:embed enemy/enemy2.png
var enemy2Image []byte

//go:embed enemy/enemy3.png
var enemy3Image []byte

//go:embed enemy/enemy4.png
var enemy4Image []byte

//go:embed enemy/enemy5.png
var enemy5Image []byte

//go:embed enemy/boss1.png
var boss1Image []byte

//go:embed misc/asteroid1.png
var asteroid1Image []byte

//go:embed misc/explosion1.png
var explosion1Image []byte

//go:embed misc/explosion2-anim.png
var explosion2Animation []byte

//go:embed misc/explosion3-anim.png
var explosion3Animation []byte

//go:embed bullet/fire1-anim.png
var fire1Animation []byte

//go:embed misc/powerup1.png
var powerup1Image []byte

//go:embed misc/powerup2.png
var powerup2Image []byte

//go:embed misc/powerup3.png
var powerup3Image []byte

//go:embed misc/powerup4.png
var powerup4Image []byte

//go:embed bullet/missle1.png
var missle1Image []byte

//go:embed player/beam1.png
var beam1Image []byte

//go:embed bullet/wave1.png
var wave1Image []byte

//go:embed misc/hit.png
var hitImage []byte

//go:embed misc/hit2.png
var hit2Image []byte

//go:embed bullet/rotate1.png
var rotate1Image []byte

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
        case ImageEnemy3: return loadPng(enemy3Image)
        case ImageEnemy4: return loadPng(enemy4Image)
        case ImageEnemy5: return loadPng(enemy5Image)
        case ImageBoss1: return loadPng(boss1Image)
        case ImageExplosion1: return loadPng(explosion1Image)
        case ImageExplosion2: return loadPng(explosion2Animation)
        case ImageExplosion3: return loadPng(explosion3Animation)
        case ImageHit: return loadPng(hitImage)
        case ImageBeam1: return loadPng(beam1Image)
        case ImageWave1: return loadPng(wave1Image)
        case ImageRotate1: return loadPng(rotate1Image)
        case ImageHit2: return loadPng(hit2Image)
        case ImageMissle1: return loadPng(missle1Image)
        case ImageBulletSmallBlue: return loadPng(bulletSmallBlueImage)
        case ImagePowerup1: return loadPng(powerup1Image)
        case ImagePowerup2: return loadPng(powerup2Image)
        case ImagePowerup3: return loadPng(powerup3Image)
        case ImagePowerup4: return loadPng(powerup4Image)
        case ImageAsteroid1: return loadPng(asteroid1Image)
        case ImageFire1: return loadPng(fire1Animation)
    }

    return nil, fmt.Errorf("no such image: %s", name)
}

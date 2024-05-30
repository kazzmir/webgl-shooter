package main

import (
    "math"
    "log"
    "image"

    gameImages "github.com/kazzmir/webgl-shooter/images"
    "github.com/hajimehoshi/ebiten/v2"
)

type Asteroid struct {
    x, y float64
    velocityX float64
    velocityY float64
    rotation uint64
    rotationSpeed float64
    health float64
}

func MakeAsteroid(x float64, y float64) *Asteroid {
    angle := randomFloat(90 - 45, 90 + 45)
    speed := randomFloat(1, 3)
    return &Asteroid{
        x: x,
        y: y,
        velocityX: speed * math.Cos(angle * math.Pi / 180),
        velocityY: speed * math.Sin(angle * math.Pi / 180),
        rotation: 0,
        rotationSpeed: randomFloat(1, 4),
        health: randomFloat(5, 20),
    }
}

func (asteroid *Asteroid) IsAlive() bool {
    return asteroid.health > 0 && asteroid.y < ScreenHeight + 100
}

func (asteroid *Asteroid) Damage(amount float64){
    asteroid.health -= amount
}

func (asteroid *Asteroid) Move() {
    asteroid.x += asteroid.velocityX
    asteroid.y += asteroid.velocityY
    asteroid.rotation += 1
}

func (asteroid *Asteroid) Collision(x float64, y float64, imageManager *ImageManager) bool {
    _, raw, err := imageManager.LoadImage(gameImages.ImageAsteroid1)
    if err != nil {
        return false
    }

    bounds := raw.Bounds()

    useX := asteroid.x - float64(bounds.Dx()) / 2
    useY := asteroid.y - float64(bounds.Dy()) / 2

    // FIXME: to handle rotation, we could counter-rotate the x/y coordinate

    if x >= useX && x <= useX + float64(bounds.Dx()) && y >= useY && y <= useY + float64(bounds.Dy()) {
        _, _, _, a := raw.At(int(x - useX), int(y - useY)).RGBA()
        return a > 200 * 255
    }

    return false
}

func (asteroid *Asteroid) Collide(player *Player, imageManager *ImageManager) bool {
    _, raw, err := imageManager.LoadImage(gameImages.ImageAsteroid1)
    if err != nil {
        return false
    }

    from := raw.Bounds().Add(image.Point{
        X: int(asteroid.x - float64(raw.Bounds().Dx()) / 2),
        Y: int(asteroid.y - float64(raw.Bounds().Dy()) / 2),
    })

    return isColliding(from, player)
}

func (asteroid *Asteroid) Draw(screen *ebiten.Image, imageManager *ImageManager, shaders *ShaderManager) {
    pic, _, err := imageManager.LoadImage(gameImages.ImageAsteroid1)
    if err != nil {
        log.Printf("Unable to load asteroid image: %v", err)
    } else {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(-float64(pic.Bounds().Dx()) / 2, -float64(pic.Bounds().Dy()) / 2)
        radians := float64(asteroid.rotation) * asteroid.rotationSpeed * math.Pi / 180
        options.GeoM.Rotate(radians)
        options.GeoM.Translate(asteroid.x, asteroid.y)
        screen.DrawImage(pic, options)
    }
}

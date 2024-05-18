package main

import (
    "log"
    "math"
    "math/rand"

    gameImages "github.com/kazzmir/webgl-shooter/images"

    "github.com/hajimehoshi/ebiten/v2"
)

type Movement interface {
    Move(x float64, y float64) (float64, float64)
    Coords(x float64, y float64) (float64, float64)
    Copy() Movement
}

type LinearMovement struct {
    velocityX, velocityY float64
}

func (linear *LinearMovement) Copy() Movement {
    return &LinearMovement{
        velocityX: linear.velocityX,
        velocityY: linear.velocityY,
    }
}

func (linear *LinearMovement) Move(x float64, y float64) (float64, float64) {
    return x + linear.velocityX, y + linear.velocityY
}

func (linear *LinearMovement) Coords(x float64, y float64) (float64, float64) {
    return x, y
}

type CircularMovement struct {
    radius float64
    angle uint64
    speed float64
    velocityX float64
    velocityY float64
}

func (circular *CircularMovement) Copy() Movement {
    return &CircularMovement{
        radius: circular.radius,
        angle: circular.angle,
        speed: circular.speed,
        velocityX: circular.velocityX,
        velocityY: circular.velocityY,
    }
}

func (circular *CircularMovement) Move(x float64, y float64) (float64, float64) {
    circular.angle += 1
    return x + circular.velocityX, y + circular.velocityY
}

func (circular *CircularMovement) Coords(x float64, y float64) (float64, float64) {
    radians := float64(circular.angle) * circular.speed * math.Pi / 180.0
    return x + circular.radius * math.Cos(radians), y + circular.radius * math.Sin(radians)
}

func makeMovement() Movement {
    switch rand.Intn(2) {
        case 0:
            return &LinearMovement{
                velocityX: randomFloat(-1, 1),
                velocityY: 2,
            }
        case 1:
            return &CircularMovement{
                radius: 75,
                angle: 0,
                speed: randomFloat(0.8, 2.8),
                velocityX: 0,
                velocityY: 2,
            }
    }

    return nil
}

type Enemy interface {
    Move(player *Player, imageManager *ImageManager) []*Bullet
    Hit(bullet *Bullet)
    Coords() (float64, float64)
    IsAlive() bool
    Draw(screen *ebiten.Image, shaders *ShaderManager)
    // returns true if this enemy is colliding with the point
    Collision(x, y float64) bool
}

type NormalEnemy struct {
    x, y float64
    // velocityX, velocityY float64
    Life float64
    pic *ebiten.Image
    Flip bool
    hurt int
    move Movement
}

func (enemy *NormalEnemy) Coords() (float64, float64) {
    return enemy.move.Coords(enemy.x, enemy.y)
}

func (enemy *NormalEnemy) IsAlive() bool {
    return enemy.Life > 0 && (enemy.y < 0 || onScreen(enemy.x, enemy.y, 100))
}

func (enemy *NormalEnemy) Hit(bullet *Bullet) {
    enemy.hurt = 10
    enemy.Life -= bullet.Strength
    if enemy.Life <= 0 {
        /*
        enemy.x = randomFloat(50, ScreenWidth - 50)
        enemy.y = randomFloat(-500, -50)
        enemy.Life = 10
        */
    }
}

func (enemy *NormalEnemy) Move(player *Player, imageManager *ImageManager) []*Bullet {
    enemy.x, enemy.y = enemy.move.Move(enemy.x, enemy.y)

    /*
    enemy.x += enemy.velocityX
    enemy.y += enemy.velocityY
    */

    /*
    if enemy.y > ScreenHeight + 50 {
        enemy.y = -100
    }
    */

    if enemy.hurt > 0 {
        enemy.hurt -= 1
    }

    if rand.Intn(100) == 0 {
        bulletPic, err := imageManager.LoadAnimation(gameImages.ImageRotate1)
        if err != nil {
            log.Printf("Unable to load bullet: %v", err)
        } else {
            useX, useY := enemy.move.Coords(enemy.x, enemy.y)
            bullet := Bullet{
                x: useX,
                y: useY + float64(enemy.pic.Bounds().Dy()) / 2,
                Strength: 1,
                velocityX: 0,
                velocityY: 1.5,
                pic: nil,
                animation: bulletPic,
                alive: true,
            }

            return []*Bullet{&bullet}
        }
    }

    return nil
}

func (enemy* NormalEnemy) Collision(x float64, y float64) bool {
    bounds := enemy.pic.Bounds()

    useX, useY := enemy.move.Coords(enemy.x, enemy.y)

    enemyX := useX - float64(bounds.Dx()) / 2
    enemyY := useY - float64(bounds.Dy()) / 2

    return x >= enemyX && x <= enemyX + float64(bounds.Dx()) && y >= enemyY && y <= enemyY + float64(bounds.Dy())
}

func (enemy *NormalEnemy) Draw(screen *ebiten.Image, shaders *ShaderManager) {

    useX, useY := enemy.move.Coords(enemy.x, enemy.y)

    enemyX := useX - float64(enemy.pic.Bounds().Dx()) / 2
    enemyY := useY - float64(enemy.pic.Bounds().Dy()) / 2

    // draw shadow
    shaderOptions := &ebiten.DrawRectShaderOptions{}
    if enemy.Flip {
        shaderOptions.GeoM.Translate(-float64(enemy.pic.Bounds().Dx()) / 2, -float64(enemy.pic.Bounds().Dy()) / 2)
        shaderOptions.GeoM.Rotate(math.Pi)
        shaderOptions.GeoM.Translate(float64(enemy.pic.Bounds().Dx()) / 2, float64(enemy.pic.Bounds().Dy()) / 2)
    }
    shaderOptions.GeoM.Translate(enemyX, enemyY + 10)
    shaderOptions.Blend = AlphaBlender
    shaderOptions.Images[0] = enemy.pic
    bounds := enemy.pic.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.ShadowShader, shaderOptions)

    if enemy.hurt > 0 {
        hurtOptions := &ebiten.DrawRectShaderOptions{}
        if enemy.Flip {
            hurtOptions.GeoM.Translate(-float64(enemy.pic.Bounds().Dx()) / 2, -float64(enemy.pic.Bounds().Dy()) / 2)
            hurtOptions.GeoM.Rotate(math.Pi)
            hurtOptions.GeoM.Translate(float64(enemy.pic.Bounds().Dx()) / 2, float64(enemy.pic.Bounds().Dy()) / 2)
        }
        hurtOptions.GeoM.Translate(enemyX, enemyY)
        hurtOptions.Uniforms = make(map[string]interface{})
        hurtOptions.Uniforms["Red"] = float32(math.Min(1.0, float64(enemy.hurt) / 8.0))
        hurtOptions.Blend = AlphaBlender
        hurtOptions.Images[0] = enemy.pic
        bounds := enemy.pic.Bounds()
        screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, hurtOptions)

    } else {

        options := &ebiten.DrawImageOptions{}
        // flip 180 degrees
        if enemy.Flip {
            options.GeoM.Translate(-float64(enemy.pic.Bounds().Dx()) / 2, -float64(enemy.pic.Bounds().Dy()) / 2)
            options.GeoM.Rotate(math.Pi)
            options.GeoM.Translate(float64(enemy.pic.Bounds().Dx()) / 2, float64(enemy.pic.Bounds().Dy()) / 2)
            // options.GeoM.Rotate(1, -1)
        }
        options.GeoM.Translate(enemyX, enemyY)
        screen.DrawImage(enemy.pic, options)
    }

    /*
    vector.StrokeRect(
        screen,
        float32(enemyX),
        float32(enemyY),
        float32(enemy.pic.Bounds().Dx()),
        float32(enemy.pic.Bounds().Dy()),
        1,
        &color.RGBA{R: 255, G: 0, B: 0, A: 128},
        true)
        */
}

func MakeEnemy1(x float64, y float64, image *ebiten.Image, move Movement) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        move: move,
        Life: 10,
        pic: image,
        Flip: true,
        hurt: 0,
    }, nil
}

func MakeEnemy2(x float64, y float64, pic *ebiten.Image, move Movement) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        move: move,
        Life: 10,
        pic: pic,
        Flip: false,
        hurt: 0,
    }, nil
}

type Coordinate struct {
    x, y float64
}

func MakeGroupGeneratorX() chan Coordinate {
    out := make(chan Coordinate)

    go func(){
        defer close(out)
        out <- Coordinate{x: 0, y: 0}
        out <- Coordinate{x: -50, y: -50}
        out <- Coordinate{x: 50, y: -50}
        out <- Coordinate{x: -50, y: 50}
        out <- Coordinate{x: 50, y: 50}
    }()

    return out
}

func MakeGroupGeneratorVertical(many int) chan Coordinate {
    out := make(chan Coordinate)

    go func(){
        defer close(out)
        y := float64(0)
        for i := 0; i < many; i++ {
            out <- Coordinate{x: 0, y: y}
            y -= 50
        }
    }()

    return out
}

func MakeGroupGenerator1x2() chan Coordinate {
    out := make(chan Coordinate)

    go func(){
        defer close(out)
        out <- Coordinate{x: -50, y: 0}
        out <- Coordinate{x: 50, y: 0}
    }()

    return out
}

func MakeGroupGenerator2x2() chan Coordinate {
    out := make(chan Coordinate)

    go func(){
        defer close(out)
        out <- Coordinate{x: -50, y: 0}
        out <- Coordinate{x: 50, y: 0}
        out <- Coordinate{x: -50, y: -50}
        out <- Coordinate{x: 50, y: -50}
    }()

    return out
}


func MakeGroupGeneratorCircle(radius int, many int) chan Coordinate {
    out := make(chan Coordinate)

    go func(){
        defer close(out)
        for i := 0; i < many; i++ {
            radians := float64(i) * 2 * math.Pi / float64(many)
            x := float64(radius) * math.Cos(radians)
            y := float64(radius) * math.Sin(radians)
            out <- Coordinate{x: x, y: y}
        }
    }()

    return out
}


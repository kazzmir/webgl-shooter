package main

import (
    "log"
    "math"
    "math/rand"

    "image"

    gameImages "github.com/kazzmir/webgl-shooter/images"

    "github.com/hajimehoshi/ebiten/v2"
)

func ptr[T any](obj T) *T { return &obj }

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

type SineMovement struct {
    amplitude float64
    angle float64
    velocityX, velocityY float64
}

func (sine *SineMovement) Copy() Movement {
    return &SineMovement{
        amplitude: sine.amplitude,
        angle: sine.angle,
        velocityX: sine.velocityX,
        velocityY: sine.velocityY,
    }
}

func (sine *SineMovement) Move(x float64, y float64) (float64, float64) {
    sine.angle += 3
    if sine.angle > 360 {
        sine.angle -= 360
    }
    return x + sine.velocityX, y + sine.velocityY
}

func (sine *SineMovement) Coords(x float64, y float64) (float64, float64) {
    radians := sine.angle * math.Pi / 180.0
    return x + sine.amplitude * math.Cos(radians), y
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
    switch rand.Intn(3) {
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
        case 2:
            return &SineMovement{
                amplitude: randomFloat(50, 100),
                velocityX: 0,
                velocityY: randomFloat(1, 2),
            }
    }

    return nil
}

type EnemyGun interface {
    Shoot(x float64, y float64, player *Player, imageManager *ImageManager) []*Bullet
}

type EnemyGun1 struct {
}

func (gun *EnemyGun1) Shoot(x float64, y float64, player *Player, imageManager *ImageManager) []*Bullet {
    if rand.Intn(100) == 0 {
        bulletPic, err := imageManager.LoadAnimation(gameImages.ImageRotate1)
        if err != nil {
            log.Printf("Unable to load bullet: %v", err)
            return nil
        } else {
            bullet := Bullet{
                x: x,
                y: y,
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

type EnemyGun2 struct {
}

func (gun *EnemyGun2) Shoot(x float64, y float64, player *Player, imageManager *ImageManager) []*Bullet {
    if rand.Intn(100) == 0 {
        bulletPic, _, err := imageManager.LoadImage(gameImages.ImageBulletSmallBlue)
        if err != nil {
            log.Printf("Unable to load bullet: %v", err)
            return nil
        } else {
            // in radians
            angleToPlayer := math.Atan2(player.y - y, player.x - x)
            speed := 1.1

            bullet := Bullet{
                x: x,
                y: y,
                Strength: 1,
                velocityX: math.Cos(angleToPlayer) * speed,
                velocityY: math.Sin(angleToPlayer) * speed,
                pic: bulletPic,
                animation: nil,
                alive: true,
            }

            return []*Bullet{&bullet}
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

    Dead() chan struct{}
}

type NormalEnemy struct {
    x, y float64
    // velocityX, velocityY float64
    Life float64
    rawImage image.Image
    pic *ebiten.Image
    Flip bool
    hurt int
    gun EnemyGun
    move Movement
    dead chan struct{}
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
        close(enemy.dead)
    }
}

func (enemy *NormalEnemy) Move(player *Player, imageManager *ImageManager) []*Bullet {
    enemy.x, enemy.y = enemy.move.Move(enemy.x, enemy.y)

    if enemy.hurt > 0 {
        enemy.hurt -= 1
    }

    var bullets []*Bullet

    useX, useY := enemy.move.Coords(enemy.x, enemy.y)
    bullets = enemy.gun.Shoot(useX, useY + float64(enemy.pic.Bounds().Dy()) / 2, player, imageManager)

    return bullets
}

func (enemy* NormalEnemy) Collision(x float64, y float64) bool {
    bounds := enemy.pic.Bounds()

    useX, useY := enemy.move.Coords(enemy.x, enemy.y)

    enemyX := useX - float64(bounds.Dx()) / 2
    enemyY := useY - float64(bounds.Dy()) / 2

    if x >= enemyX && x <= enemyX + float64(bounds.Dx()) && y >= enemyY && y <= enemyY + float64(bounds.Dy()) {
        _, _, _, a := enemy.rawImage.At(int(x - enemyX), int(y - enemyY)).RGBA()
        return a > 200
    }

    return false
}

func (enemy *NormalEnemy) Dead() chan struct{} {
    return enemy.dead
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

func MakeEnemy1(x float64, y float64, rawImage image.Image, image *ebiten.Image, move Movement) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        move: move,
        Life: 10,
        rawImage: rawImage,
        pic: image,
        gun: &EnemyGun2{},
        Flip: true,
        hurt: 0,
        dead: make(chan struct{}),
    }, nil
}

func MakeEnemy2(x float64, y float64, rawImage image.Image, pic *ebiten.Image, move Movement) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        move: move,
        Life: 10,
        rawImage: rawImage,
        pic: pic,
        gun: &EnemyGun1{},
        Flip: false,
        hurt: 0,
        dead: make(chan struct{}),
    }, nil
}

type BossGun1 struct {
    shot1count uint32
    shot2count uint32
}

func MakeBossGun1() BossGun1 {
    return BossGun1{
        shot1count: 0,
        shot2count: 0,
    }
}

func (gun *BossGun1) NormalShot(x float64, y float64, imageManager *ImageManager) *Bullet {
    bulletPic, err := imageManager.LoadAnimation(gameImages.ImageRotate1)
    if err != nil {
        log.Printf("Unable to load bullet: %v", err)
        return nil
    } else {
        bullet := Bullet{
            x: x,
            y: y,
            Strength: 1,
            velocityX: 0,
            velocityY: 1.5,
            pic: nil,
            animation: bulletPic,
            alive: true,
        }

        return &bullet
    }
}

func (gun *BossGun1) AimShot(x float64, y float64, player *Player, imageManager *ImageManager) *Bullet {
    bulletPic, _, err := imageManager.LoadImage(gameImages.ImageBulletSmallBlue)
    if err != nil {
        log.Printf("Unable to load bullet: %v", err)
        return nil
    } else {
        // in radians
        angleToPlayer := math.Atan2(player.y - y, player.x - x)
        speed := 1.8

        bullet := Bullet{
            x: x,
            y: y,
            Strength: 1,
            velocityX: math.Cos(angleToPlayer) * speed,
            velocityY: math.Sin(angleToPlayer) * speed,
            pic: bulletPic,
            animation: nil,
            alive: true,
        }

        return &bullet
    }
}

func (gun *BossGun1) Shoot(x float64, y float64, player *Player, imageManager *ImageManager) []*Bullet {

    const shot1rate = 30
    const shot2rate = 20

    if gun.shot1count == 0 && rand.Intn(50) == 0 {
        gun.shot1count = shot1rate * 3
    }

    if gun.shot2count == 0 && rand.Intn(100) == 0 {
        gun.shot2count = shot2rate * 3
    }

    var bullets []*Bullet

    if gun.shot1count > 0 && gun.shot1count % shot1rate == 0 {
        bullet := gun.NormalShot(x, y, imageManager)
        if bullet != nil {
            bullets = append(bullets, bullet)
        }
    }

    if gun.shot2count > 0 && gun.shot2count % shot2rate == 0 {
        bullet := gun.AimShot(x, y, player, imageManager)
        if bullet != nil {
            bullets = append(bullets, bullet)
        }
    }

    if gun.shot1count > 0 {
        gun.shot1count -= 1
    }

    if gun.shot2count > 0 {
        gun.shot2count -= 1
    }

    return bullets
}

type Boss1Movement struct {
    // location we want to move to
    moveX, moveY float64
    // count how long we are at one position
    counter uint64
}

func distance(x1, y1, x2, y2 float64) float64 {
    return math.Sqrt((x2 - x1) * (x2 - x1) + (y2 - y1) * (y2 - y1))
}

func (boss *Boss1Movement) Copy() Movement {
    return &Boss1Movement{
        moveX: boss.moveX,
        moveY: boss.moveY,
        counter: boss.counter,
    }
}

func (boss *Boss1Movement) Coords(x float64, y float64) (float64, float64) {
    return x, y
}

func (boss *Boss1Movement) Move(x float64, y float64) (float64, float64) {

    speed := 1.5

    if distance(x, y, boss.moveX, boss.moveY) < speed * 2 {
        if boss.counter == 0 {
            boss.moveX = randomFloat(100, ScreenWidth - 100)
            boss.moveY = randomFloat(100, ScreenHeight - 100)
            boss.counter = uint64(rand.Intn(200) + 200)
        } else {
            boss.counter -= 1
        }

        return x, y
    } else {
        angle := math.Atan2(boss.moveY - y, boss.moveX - x)
        return x + math.Cos(angle) * speed, y + math.Sin(angle) * speed
    }
}

func MakeBoss1(x float64, y float64, rawImage image.Image, pic *ebiten.Image) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        move: &Boss1Movement{
            moveX: ScreenWidth / 2,
            moveY: 100,
            counter: 100,
        },
        Life: 500,
        rawImage: rawImage,
        pic: pic,
        gun: ptr(MakeBossGun1()),
        Flip: false,
        hurt: 0,
        dead: make(chan struct{}),
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


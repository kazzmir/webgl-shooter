package main

import (
    "log"
    "fmt"
    "io"
    "time"
    "bytes"
    "math/rand"
    "math"
    "sync"

    "image/color"
    _ "image/png"

    gameImages "github.com/kazzmir/webgl-shooter/images"
    fontLib "github.com/kazzmir/webgl-shooter/font"
    audioFiles "github.com/kazzmir/webgl-shooter/audio"

    "github.com/hajimehoshi/ebiten/v2"
    _ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/audio"
    // "github.com/hajimehoshi/ebiten/v2/audio/mp3"
    // "github.com/hajimehoshi/ebiten/v2/vector"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Explosion interface {
    Move()
    IsAlive() bool
    Draw(shaderManager *ShaderManager, screen *ebiten.Image)
}

type SimpleExplosion struct {
    x, y float64
    velocityX, velocityY float64
    pic *ebiten.Image
    life int
}

func MakeSimpleExplosion(x float64, y float64, pic *ebiten.Image) Explosion {
    return &SimpleExplosion{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 0,
        pic: pic,
        life: 10,
    }
}

func (explosion *SimpleExplosion) Move() {
    explosion.x += explosion.velocityX
    explosion.y += explosion.velocityY
    explosion.life -= 1
}

func (explosion *SimpleExplosion) IsAlive() bool {
    return explosion.life > 0
}

func (explosion *SimpleExplosion) Draw(shaderManager *ShaderManager, screen *ebiten.Image) {
    bounds := explosion.pic.Bounds()
    posX := explosion.x - float64(bounds.Dx()) / 2
    posY := explosion.y - float64(bounds.Dy()) / 2

    /*
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(posX, posY)
    screen.DrawImage(explosion.pic, options)
    */

    options := &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(posX, posY)
    options.Blend = AlphaBlender
    options.Images[0] = explosion.pic
    options.Uniforms = make(map[string]interface{})
    // radians = math.Pi * 90 / 180
    // log.Printf("Red: %v", radians)
    options.Uniforms["Center"] = []float32{float32(explosion.x), float32(explosion.y)}
    // options.Uniforms["Center"] = []float32{float32(bounds.Dx()) / 2, float32(bounds.Dy()) / 2}
    options.Uniforms["InnerRadius"] = float32(math.Max(0, float64(5-explosion.life)))
    options.Uniforms["OuterRadius"] = float32(math.Max(0, float64(10-explosion.life)))

    // log.Printf("Uniforms: %v", options.Uniforms)
    // options.Uniforms["InnerRadius"] = float32(10)
    // options.Uniforms["OuterRadius"] = float32(100)
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaderManager.ExplosionShader, options)
}

type AnimatedExplosion struct {
    x, y float64
    velocityX, velocityY float64
    animation *Animation
}

func (explosion *AnimatedExplosion) Move() {
    explosion.x += explosion.velocityX
    explosion.y += explosion.velocityY
    explosion.animation.Update()
}

func (explosion *AnimatedExplosion) IsAlive() bool {
    return explosion.animation.IsAlive()
}

func (explosion *AnimatedExplosion) Draw(shaderManager *ShaderManager, screen *ebiten.Image) {
    explosion.animation.Draw(screen, explosion.x, explosion.y)
}


func MakeAnimatedExplosion(x float64, y float64, animation *Animation) Explosion {
    return &AnimatedExplosion{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 0,
        animation: animation,
    }
}

type Bullet struct {
    x, y float64
    velocityX, velocityY float64
    pic *ebiten.Image
    alive bool
}

func (bullet *Bullet) Draw(screen *ebiten.Image) {
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(bullet.x - float64(bullet.pic.Bounds().Dx()) / 2, bullet.y - float64(bullet.pic.Bounds().Dy()) / 2)
    screen.DrawImage(bullet.pic, options)
}

func (bullet *Bullet) Move(){
    bullet.x += bullet.velocityX
    bullet.y += bullet.velocityY
}

func (bullet *Bullet) SetDead() {
    bullet.alive = false
}

func (bullet *Bullet) IsAlive() bool {
    return bullet.alive && bullet.y > 0
}

type StarPosition struct {
    x, y float64
    dx, dy float64
    Image *ebiten.Image
}

type Background struct {
    // Star *ebiten.Image
    // Star2 *ebiten.Image
    Stars []*StarPosition
}

func randomFloat(min float64, max float64) float64 {
    return min + rand.Float64() * (max - min)
}

func MakeBackground() (*Background, error) {
    starImage, err := gameImages.LoadImage(gameImages.ImageStar1)
    if err != nil {
        return nil, err
    }

    starImage2, err := gameImages.LoadImage(gameImages.ImageStar2)
    if err != nil {
        return nil, err
    }

    planet1, err := gameImages.LoadImage(gameImages.ImagePlanet)
    if err != nil {
        return nil, err
    }

    images := []*ebiten.Image{
        ebiten.NewImageFromImage(starImage),
        ebiten.NewImageFromImage(starImage2),
        ebiten.NewImageFromImage(planet1),
    }

    stars := make([]*StarPosition, 0)
    for i := 0; i < 50; i++ {
        x := randomFloat(0, float64(ScreenWidth))
        y := randomFloat(0, float64(ScreenHeight))
        dx := 0.0
        dy := randomFloat(0.6, 1.1)

        image := images[rand.Intn(len(images))]

        stars = append(stars, &StarPosition{x: x, y: y, dx: dx, dy: dy, Image: image})
    }

    return &Background{
        // Star: ebiten.NewImageFromImage(starImage),
        Stars: stars,
    }, nil
}

func (background *Background) Update(){
    for _, star := range background.Stars {
        star.y += star.dy
        if star.y > ScreenHeight + 50 {
            star.y = -50
        }
    }
}

func (background *Background) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x1b, 0x22, 0x24, 0xff})

    for _, star := range background.Stars {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(star.x, star.y)
        screen.DrawImage(star.Image, options)
    }
}

type ShaderManager struct {
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
    EdgeShader *ebiten.Shader
    ExplosionShader *ebiten.Shader
}

func MakeShaderManager() (*ShaderManager, error) {
    redShader, err := LoadRedShader()
    if err != nil {
        return nil, err
    }

    shadowShader, err := LoadShadowShader()
    if err != nil {
        return nil, err
    }

    edgeShader, err := LoadEdgeShader()
    if err != nil {
        return nil, err
    }

    explosionShader, err := LoadExplosionShader()
    if err != nil {
        return nil, err
    }

    return &ShaderManager{
        RedShader: redShader,
        ShadowShader: shadowShader,
        EdgeShader: edgeShader,
        ExplosionShader: explosionShader,
    }, nil
}

type Enemy interface {
    Move()
    Hit()
    Coords() (float64, float64)
    IsAlive() bool
    Draw(screen *ebiten.Image, shaders *ShaderManager)
    // returns true if this enemy is colliding with the point
    Collision(x, y float64) bool
}

type NormalEnemy struct {
    x, y float64
    velocityX, velocityY float64
    Life float64
    pic *ebiten.Image
    Flip bool
    hurt int
}

func (enemy *NormalEnemy) Coords() (float64, float64) {
    return enemy.x, enemy.y
}

func (enemy *NormalEnemy) IsAlive() bool {
    return enemy.Life > 0
}

func (enemy *NormalEnemy) Hit() {
    enemy.hurt = 10
    enemy.Life -= 1
    if enemy.Life <= 0 {
        /*
        enemy.x = randomFloat(50, ScreenWidth - 50)
        enemy.y = randomFloat(-500, -50)
        enemy.Life = 10
        */
    }
}

func (enemy *NormalEnemy) Move() {
    enemy.x += enemy.velocityX
    enemy.y += enemy.velocityY

    if enemy.y > ScreenHeight + 50 {
        enemy.y = -100
    }

    if enemy.hurt > 0 {
        enemy.hurt -= 1
    }
}

func (enemy* NormalEnemy) Collision(x float64, y float64) bool {
    bounds := enemy.pic.Bounds()

    enemyX := enemy.x - float64(bounds.Dx()) / 2
    enemyY := enemy.y - float64(bounds.Dy()) / 2

    return x >= enemyX && x <= enemyX + float64(bounds.Dx()) && y >= enemyY && y <= enemyY + float64(bounds.Dy())
}

func (enemy *NormalEnemy) Draw(screen *ebiten.Image, shaders *ShaderManager) {

    enemyX := enemy.x - float64(enemy.pic.Bounds().Dx()) / 2
    enemyY := enemy.y - float64(enemy.pic.Bounds().Dy()) / 2

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

func MakeEnemy1(x float64, y float64, image *ebiten.Image) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 2,
        Life: 10,
        pic: image,
        Flip: true,
        hurt: 0,
    }, nil
}

func MakeEnemy2(x float64, y float64, pic *ebiten.Image) (Enemy, error) {
    return &NormalEnemy{
        x: x,
        y: y,
        velocityX: 0,
        velocityY: 2,
        Life: 10,
        pic: pic,
        Flip: false,
        hurt: 0,
    }, nil
}

type Player struct {
    x, y float64
    Jump int
    velocityX, velocityY float64
    bulletCounter int
    pic *ebiten.Image
    bullet *ebiten.Image
    Score int
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
    Counter int
    SoundShoot chan bool
}

func (player *Player) Move() {
    player.Counter += 1

    player.x += player.velocityX
    player.y += player.velocityY

    accel := 0.23

    if player.velocityX < -accel {
        player.velocityX += accel
    } else if player.velocityX > accel {
        player.velocityX -= accel
    } else {
        player.velocityX = 0
    }

    if player.velocityY < -accel {
        player.velocityY += accel
    } else if player.velocityY > accel {
        player.velocityY -= accel
    } else {
        player.velocityY = 0
    }

    if player.x < 0 {
        player.x = 0
    } else if player.x > ScreenWidth {
        player.x = ScreenWidth
    }

    if player.y < 0 {
        player.y = 0
    } else if player.y > ScreenHeight {
        player.y = ScreenHeight
    }

    if player.bulletCounter > 0 {
        player.bulletCounter -= 1
    }
}

func (player *Player) MakeBullet() *Bullet {

    velocityY := player.velocityY-2
    if velocityY > -1 {
        velocityY = -1
    }

    velocityY = -2.5

    return &Bullet{
        x: player.x,
        y: player.y - float64(player.pic.Bounds().Dy()) / 2,
        alive: true,
        velocityX: 0,
        velocityY: velocityY,
        pic: player.bullet,
    }
}

var AlphaBlender ebiten.Blend = ebiten.Blend{
    BlendFactorSourceRGB:        ebiten.BlendFactorSourceAlpha,
    BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
    BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
    BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
    BlendOperationRGB:           ebiten.BlendOperationAdd,
    BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

func (player *Player) Draw(screen *ebiten.Image, shaders *ShaderManager, font *text.GoTextFaceSource) {
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Score: %v", player.Score), &text.GoTextFace{Source: font, Size: 15}, op)

    playerX := player.x - float64(player.pic.Bounds().Dx()) / 2
    playerY := player.y - float64(player.pic.Bounds().Dy()) / 2

    options := &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(playerX + player.velocityX * 3, playerY + 10)
    options.Blend = AlphaBlender
    options.Images[0] = player.pic
    bounds := player.pic.Bounds()
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.ShadowShader, options)

    /*
    options := &ebiten.DrawImageOptions{}
    options.GeoM.Translate(player.x, player.y)
    screen.DrawImage(player.pic, options)
    */

    if player.Jump > 0 {
        options := &ebiten.DrawRectShaderOptions{}
        options.GeoM.Translate(playerX, playerY)
        options.Blend = AlphaBlender
        options.Images[0] = player.pic
        options.Uniforms = make(map[string]interface{})
        var radians float32 = math.Pi * float32(player.Jump) * 360 / JumpDuration / 180.0
        // radians = math.Pi * 90 / 180
        // log.Printf("Red: %v", radians)
        options.Uniforms["Red"] = radians
        bounds := player.pic.Bounds()
        screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.RedShader, options)
    } else {
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(playerX, playerY)
        screen.DrawImage(player.pic, options)
    }

    options = &ebiten.DrawRectShaderOptions{}
    options.GeoM.Translate(playerX, playerY)
    options.Blend = AlphaBlender
    options.Images[0] = player.pic
    options.Uniforms = make(map[string]interface{})
    options.Uniforms["Color"] = []float32{0, 0, float32((math.Sin(float64(player.Counter) * 7 * math.Pi / 180.0) + 1) / 2), 1}
    // options.Uniforms["Color"] = []float32{0, 0, 1, 1}
    screen.DrawRectShader(bounds.Dx(), bounds.Dy(), shaders.EdgeShader, options)
}

func (player *Player) HandleKeys(game *Game) error {
    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)
    playerAccel := 3.8
    if player.Jump > 0 {
        playerAccel = 5
    }
    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            player.velocityY = -playerAccel;
        } else if key == ebiten.KeyArrowDown {
            player.velocityY = playerAccel;
        } else if key == ebiten.KeyArrowLeft {
            player.velocityX = -playerAccel;
        } else if key == ebiten.KeyArrowRight {
            player.velocityX = playerAccel;
        } else if key == ebiten.KeyShift && player.Jump <= -50 {
            player.Jump = JumpDuration
        // FIXME: make ebiten understand key mapping
        } else if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        } else if key == ebiten.KeySpace && game.Player.bulletCounter == 0 {
            game.Bullets = append(game.Bullets, game.Player.MakeBullet())
            player.bulletCounter = 5

            select {
                case <-player.SoundShoot:
                    game.SoundManager.Play(audioFiles.AudioShoot1)
                    go func(){
                        time.Sleep(10 * time.Millisecond)
                        player.SoundShoot <- true
                    }()
                default:
            }
        }
    }

    if player.Jump > -50 {
        player.Jump -= 1
    }

    return nil
}

/*
func loadPng(path string) (image.Image, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    defer file.Close()

    img, _, err := image.Decode(file)
    return img, err
}
*/

const JumpDuration = 50

func MakePlayer(x, y float64) (*Player, error) {

    playerImage, err := gameImages.LoadImage(gameImages.ImagePlayer)

    if err != nil {
        return nil, err
    }

    bulletImage, err := gameImages.LoadImage(gameImages.ImageBullet)
    if err != nil {
        return nil, err
    }

    soundChan := make(chan bool, 2)
    soundChan <- true

    return &Player{
        x: x,
        y: y,
        pic: ebiten.NewImageFromImage(playerImage),
        bullet: ebiten.NewImageFromImage(bulletImage),
        Jump: -50,
        Score: 0,
        SoundShoot: soundChan,
    }, nil
}

type ImageManager struct {
    Images map[gameImages.Image]*ebiten.Image
}

func MakeImageManager() *ImageManager {
    return &ImageManager{
        Images: make(map[gameImages.Image]*ebiten.Image),
    }
}

func (manager *ImageManager) LoadImage(name gameImages.Image) (*ebiten.Image, error) {
    if image, ok := manager.Images[name]; ok {
        return image, nil
    }

    loaded, err := gameImages.LoadImage(name)
    if err != nil {
        return nil, err
    }

    converted := ebiten.NewImageFromImage(loaded)

    manager.Images[name] = converted
    return converted, nil
}

func (manager *ImageManager) LoadAnimation(name gameImages.Image) (*Animation, error) {
    loaded, err := manager.LoadImage(name)
    if err != nil {
        return nil, err
    }

    switch name {
        case gameImages.ImageExplosion2: return NewAnimation(loaded, 5, 6), nil
    }

    return nil, fmt.Errorf("No such animation %v", name)
}

type SoundHandler struct {
    Make func() *audio.Player
    MakeLoop func() (*audio.Player, error)
    // Players chan *audio.Player
}

type SoundManager struct {
    Sounds map[audioFiles.AudioName]*SoundHandler
    Context *audio.Context
    SampleRate int
}

func MakeSoundManager() (*SoundManager, error) {
    manager := SoundManager{
        Sounds: make(map[audioFiles.AudioName]*SoundHandler),
        SampleRate: 48000,
        Context: audio.NewContext(48000),
    }

    return &manager, manager.LoadAll()
}

func MakeSoundHandler(name audioFiles.AudioName, context *audio.Context, sampleRate int) (*SoundHandler, error) {
    var data []byte

    var create sync.Once

    load := func(){
        log.Printf("Creating sound %v", name)
        stream, err := audioFiles.LoadSound(name, sampleRate)
        if err != nil {
            log.Printf("Error loading sound %v: %v", name, err)
            return
        }

        data, err = io.ReadAll(stream)
        if err != nil {
            log.Printf("Error loading sound %v: %v", name, err)
            return
        }

        log.Printf("  loaded %v", name)
    }

    return &SoundHandler{
        Make: func() *audio.Player {
            create.Do(load)
            return context.NewPlayerFromBytes(data)
        },
        MakeLoop: func() (*audio.Player, error) {
            create.Do(load)
            return context.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data)) + 1000))
        },
    }, nil
}

func (manager *SoundManager) LoadAll() error {

    sounds := audioFiles.AllSounds
    for _, sound := range sounds {
        handler, err := MakeSoundHandler(sound, manager.Context, manager.SampleRate)
        if err != nil {
            return fmt.Errorf("Error loading %v: %v", sound, err)
        }
        manager.Sounds[sound] = handler
    }

    return nil
}

func (manager *SoundManager) Play(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        handler.Make().Play()
    }
}

func (manager *SoundManager) PlayLoop(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        go func(){
            player, err := handler.MakeLoop()
            if err != nil {
                log.Printf("Failed to play audio loop %v: %v", name, err)
            } else {
                player.Play()
            }
        }()
    }
}


type Game struct {
    Player *Player
    Background *Background
    Bullets []*Bullet
    Font *text.GoTextFaceSource
    Enemies []Enemy
    Explosions []Explosion
    ShaderManager *ShaderManager
    ImageManager *ImageManager
    SoundManager *SoundManager

    MusicPlayer sync.Once
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

func (game *Game) MakeEnemy(x float64, y float64, kind int) error {
    var enemy Enemy
    var err error

    switch kind {
        case 0:
            pic, err := game.ImageManager.LoadImage(gameImages.ImageEnemy1)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy1(x, y, pic)
        case 1:
            pic, err := game.ImageManager.LoadImage(gameImages.ImageEnemy2)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, pic)
    }

    if err != nil {
        return err
    }

    game.Enemies = append(game.Enemies, enemy)

    return nil
}

func (game *Game) MakeEnemies(count int) error {

    for i := 0; i < count; i++ {
        var generator chan Coordinate
        switch rand.Intn(5) {
            case 0: generator = MakeGroupGeneratorX()
            case 1: generator = MakeGroupGeneratorVertical(rand.Intn(3) + 3)
            case 2: generator = MakeGroupGeneratorCircle(100, 6)
            case 3: generator = MakeGroupGenerator1x2()
            case 4: generator = MakeGroupGenerator2x2()
        }

        x := randomFloat(50, ScreenWidth - 50)
        y := float64(-500)
        kind := rand.Intn(2)

        for coord := range generator {
            err := game.MakeEnemy(x + coord.x, y + coord.y, kind)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (game *Game) Update() error {

    game.MusicPlayer.Do(func(){
        game.SoundManager.PlayLoop(audioFiles.AudioStellarPulseSong)
    })

    game.Background.Update()

    err := game.Player.HandleKeys(game)
    if err != nil {
        return err
    }

    game.Player.Move()

    for _, enemy := range game.Enemies {
        enemy.Move()
    }

    explosionOut := make([]Explosion, 0)
    for _, explosion := range game.Explosions {
        explosion.Move()
        if explosion.IsAlive() {
            explosionOut = append(explosionOut, explosion)
        }
    }
    game.Explosions = explosionOut

    for i := 0; i < 3; i++ {
        var outBullets []*Bullet
        for _, bullet := range game.Bullets {
            bullet.Move()

            for _, enemy := range game.Enemies {
                if enemy.IsAlive() && enemy.Collision(bullet.x, bullet.y) {
                    game.Player.Score += 1
                    enemy.Hit()
                    if ! enemy.IsAlive() {
                        game.SoundManager.Play(audioFiles.AudioExplosion3)

                        animation, err := game.ImageManager.LoadAnimation(gameImages.ImageExplosion2)
                        if err == nil {
                            x, y := enemy.Coords()
                            game.Explosions = append(game.Explosions, MakeAnimatedExplosion(x, y, animation))
                        } else {
                            log.Printf("Could not load explosion sheet: %v", err)
                        }
                    }
                    bullet.SetDead()

                    game.SoundManager.Play(audioFiles.AudioHit1)

                    explosionPic, err := game.ImageManager.LoadImage(gameImages.ImageExplosion1)
                    if err != nil {
                        log.Printf("Could not load explosion: %v", err)
                    } else {
                        explosion := MakeSimpleExplosion(bullet.x, bullet.y, explosionPic)
                        game.Explosions = append(game.Explosions, explosion)
                    }
                    break
                }
            }

            if bullet.IsAlive() {
                outBullets = append(outBullets, bullet)
            }
        }
        game.Bullets = outBullets
    }

    enemyOut := make([]Enemy, 0)
    for _, enemy := range game.Enemies {
        if enemy.IsAlive() {
            enemyOut = append(enemyOut, enemy)
        }
    }
    game.Enemies = enemyOut

    if len(game.Enemies) == 0 || (len(game.Enemies) < 50 && rand.Intn(100) == 0) {
        game.MakeEnemies(1)
    }

    return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
    game.Background.Draw(screen)

    for _, enemy := range game.Enemies {
        enemy.Draw(screen, game.ShaderManager)
    }

    for _, explosion := range game.Explosions {
        explosion.Draw(game.ShaderManager, screen)
    }

    // ebitenutil.DebugPrint(screen, "debugging")
    game.Player.Draw(screen, game.ShaderManager, game.Font)

    for _, bullet := range game.Bullets {
        bullet.Draw(screen)
    }

    /*
    ePic, err := game.ImageManager.LoadImage(gameImages.ImageExplosion1)
    if err == nil {
        e := MakeExplosion(600, 400, ePic)
        e.life = -40
        e.Draw(game.ShaderManager, screen)
    } else {
        log.Printf("Failed to make explosion: %v", err)
    }
    */
    
    // vector.StrokeRect(screen, 0, 0, 100, 100, 3, &color.RGBA{R: 255, G: 0, B: 0, A: 128}, true)
    // vector.DrawFilledRect(screen, 0, 0, 100, 100, &color.RGBA{R: 255, G: 0, B: 0, A: 64}, true)

    /*
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, "Hello, World!", &text.GoTextFace{Source: game.Font, Size: 15}, op)
    */
}

func (game *Game) PreloadAssets() error {
    // preload assets
    _, err := game.ImageManager.LoadAnimation(gameImages.ImageExplosion2)
    if err != nil {
        return err
    }

    return nil
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("Shooter")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    log.Printf("Loading objects")

    player, err := MakePlayer(ScreenWidth / 2, ScreenHeight - 100)
    if err != nil {
        log.Printf("Failed to make player: %v", err)
        return
    }

    background, err := MakeBackground()
    if err != nil {
        log.Printf("Failed to make background: %v", err)
        return
    }

    font, err := fontLib.LoadFont()
    if err != nil {
        log.Printf("Failed to load font: %v", err)
        return
    }

    shaderManager, err := MakeShaderManager()
    if err != nil {
        log.Printf("Failed to make shaders: %v", err)
        return
    }

    soundManager, err := MakeSoundManager()
    if err != nil {
        log.Printf("Failed to make sound manager: %v", err)
        return
    }

    game := Game{
        Background: background,
        Player: player,
        Font: font,
        ShaderManager: shaderManager,
        ImageManager: MakeImageManager(),
        SoundManager: soundManager,
    }

    err = game.MakeEnemies(5)
    if err != nil {
        log.Printf("Failed to make enemies: %v", err)
        return
    }

    err = game.PreloadAssets()
    if err != nil {
        log.Printf("Failed to preload assets: %v", err)
        return
    }

    log.Printf("Running")
    err = ebiten.RunGame(&game)
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }

    log.Printf("Bye!")
}

package main

import (
    "log"
    "fmt"
    "io"
    "time"
    "os"
    "bytes"
    "errors"
    "math/rand"
    "math"
    "sync"
    "sync/atomic"
    "context"
    "runtime/pprof"

    "image"
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
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const debugForceBoss = false

const ScreenWidth = 1024
const ScreenHeight = 768

func onScreen(x float64, y float64, margin float64) bool {
    return x > -margin && x < ScreenWidth + margin && y > -margin && y < ScreenHeight + margin
}

type Bullet struct {
    x, y float64
    Strength float64
    velocityX, velocityY float64
    pic *ebiten.Image
    animation *Animation
    alive bool
}

func (bullet *Bullet) Draw(screen *ebiten.Image) {

    if bullet.animation != nil {
        bullet.animation.Draw(screen, bullet.x, bullet.y)
    } else if bullet.pic != nil {
        x1 := bullet.x - float64(bullet.pic.Bounds().Dx()) / 2
        y1 := bullet.y - float64(bullet.pic.Bounds().Dy()) / 2
        options := &ebiten.DrawImageOptions{}
        options.GeoM.Translate(x1, y1)
        screen.DrawImage(bullet.pic, options)
    }
}

func (bullet *Bullet) Move(){
    bullet.x += bullet.velocityX
    bullet.y += bullet.velocityY

    if bullet.animation != nil {
        bullet.animation.Update()
    }
}

func (bullet *Bullet) SetDead() {
    bullet.alive = false
}

func (bullet *Bullet) IsAlive() bool {
    return bullet.alive && onScreen(bullet.x, bullet.y, 10)
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

type Player struct {
    x, y float64
    Jump int
    velocityX, velocityY float64
    rawImage image.Image
    pic *ebiten.Image
    Guns []Gun
    Score uint64
    Kills uint64
    RedShader *ebiten.Shader
    ShadowShader *ebiten.Shader
    Counter int
    SoundShoot chan bool
}

func (player *Player) Bounds() image.Rectangle {
    bounds := player.rawImage.Bounds()

    x1 := player.x - float64(bounds.Dx()) / 2
    y1 := player.y - float64(bounds.Dy()) / 2
    x2 := x1 + float64(bounds.Dx())
    y2 := y1 + float64(bounds.Dy())

    return image.Rect(int(x1), int(y1), int(x2), int(y2))
}

func (player *Player) Collide(x float64, y float64) bool {
    bounds := player.Bounds()
    if int(x) >= bounds.Min.X && int(x) <= bounds.Max.X && int(y) >= bounds.Min.Y && int(y) <= bounds.Max.Y {
        cx := int(x) - bounds.Min.X
        cy := int(y) - bounds.Min.Y
        c := player.rawImage.At(cx, cy)
        _, _, _, a := c.RGBA()
        if a > 200 {
            return true
        }
    }

    return false
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

    for _, gun := range player.Guns {
        gun.Update()
    }
}

func (player *Player) Shoot(imageManager *ImageManager, soundManager *SoundManager) []*Bullet {

    var bullets []*Bullet

    for _, gun := range player.Guns {
        if gun.IsEnabled() {
            more, err := gun.Shoot(imageManager, player.x, player.y - float64(player.pic.Bounds().Dy()) / 2)
            if err != nil {
                log.Printf("Could not create bullets: %v", err)
            } else {
                if more != nil {
                    bullets = append(bullets, more...)

                    select {
                        case <-player.SoundShoot:
                            // soundManager.Play(audioFiles.AudioShoot1)
                            gun.DoSound(soundManager)
                            go func(){
                                time.Sleep(10 * time.Millisecond)
                                player.SoundShoot <- true
                            }()
                        default:
                    }
                }
            }
        }
    }

    return bullets

    /*
    velocityY := player.velocityY-2
    if velocityY > -1 {
        velocityY = -1
    }
    */

    /*
    velocityY := -2.5

    return &Bullet{
        x: player.x,
        y: player.y - float64(player.pic.Bounds().Dy()) / 2,
        alive: true,
        velocityX: 0,
        velocityY: velocityY,
        pic: player.bullet,
    }
    */
}

var AlphaBlender ebiten.Blend = ebiten.Blend{
    BlendFactorSourceRGB:        ebiten.BlendFactorSourceAlpha,
    BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
    BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
    BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
    BlendOperationRGB:           ebiten.BlendOperationAdd,
    BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

func (player *Player) Draw(screen *ebiten.Image, shaders *ShaderManager, imageManger *ImageManager, font *text.GoTextFaceSource) {
    face := &text.GoTextFace{Source: font, Size: 15} 

    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Score: %v", player.Score), face, op)

    op.GeoM.Translate(1, 20)
    text.Draw(screen, fmt.Sprintf("Kills: %v", player.Kills), face, op)

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

    var iconX float64 = 150
    var iconY float64 = 3
    for _, gun := range player.Guns {
        gun.DrawIcon(screen, imageManger, font, iconX, iconY)
        iconX += 30
    }
}

func enableGun[T Gun] (guns []Gun) {
    for _, gun := range guns {
        switch gun.(type) {
            case T:
                gun.SetEnabled(!gun.IsEnabled())
        }
    }
}

func (player *Player) HandleKeys(game *Game, run *Run) error {
    keys := make([]ebiten.Key, 0)

    keys = inpututil.AppendPressedKeys(keys)

    maxVelocity := 3.8

    playerAccel := 0.9
    if player.Jump > 0 {
        playerAccel = 3
        maxVelocity = 5.5
    }

    for _, key := range keys {
        if key == ebiten.KeyArrowUp {
            player.velocityY -= playerAccel
        } else if key == ebiten.KeyArrowDown {
            player.velocityY += playerAccel
        } else if key == ebiten.KeyArrowLeft {
            player.velocityX -= playerAccel
        } else if key == ebiten.KeyArrowRight {
            player.velocityX += playerAccel
            // player.Gun = &BasicGun{}
        } else if key == ebiten.KeyDigit3 {
            // player.Gun = &BeamGun{}
        } else if key == ebiten.KeyDigit4 {
            // player.Gun = &MissleGun{}
        } else if key == ebiten.KeyShift && player.Jump <= -50 {
            player.Jump = JumpDuration
        // FIXME: make ebiten understand key mapping
        } else if key == ebiten.KeySpace {
            game.Bullets = append(game.Bullets, game.Player.Shoot(game.ImageManager, game.SoundManager)...)
        }
    }

    player.velocityX = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityX))
    player.velocityY = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityY))

    moreKeys := make([]ebiten.Key, 0)
    moreKeys = inpututil.AppendJustPressedKeys(moreKeys)
    for _, key := range moreKeys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            // return ebiten.Termination
            run.Mode = RunMenu
        } else if key == ebiten.KeyDigit1 {
            enableGun[*BasicGun](player.Guns)
        } else if key == ebiten.KeyDigit2 {
            enableGun[*DualBasicGun](player.Guns)
        }
    }

    if player.Jump > -50 {
        player.Jump -= 1
    }

    return nil
}

const JumpDuration = 50

func MakePlayer(x, y float64) (*Player, error) {

    playerImage, err := gameImages.LoadImage(gameImages.ImagePlayer)

    if err != nil {
        return nil, err
    }

    soundChan := make(chan bool, 2)
    soundChan <- true

    return &Player{
        x: x,
        y: y,
        rawImage: playerImage,
        pic: ebiten.NewImageFromImage(playerImage),
        // Gun: &BasicGun{},
        // Gun: &DualBasicGun{},
        Guns: []Gun{
            &BasicGun{enabled: true},
            &DualBasicGun{enabled: false},
        },
        // Gun: &BeamGun{},
        Jump: -50,
        Score: 0,
        SoundShoot: soundChan,
    }, nil
}

type ImagePair struct {
    Image *ebiten.Image
    Raw image.Image
}

type ImageManager struct {
    Images map[gameImages.Image]ImagePair
}

func MakeImageManager() *ImageManager {
    return &ImageManager{
        Images: make(map[gameImages.Image]ImagePair),
    }
}

func (manager *ImageManager) LoadImage(name gameImages.Image) (*ebiten.Image, image.Image, error) {
    if image, ok := manager.Images[name]; ok {
        return image.Image, image.Raw, nil
    }

    loaded, err := gameImages.LoadImage(name)
    if err != nil {
        return nil, nil, err
    }

    converted := ebiten.NewImageFromImage(loaded)

    manager.Images[name] = ImagePair{
        Image: converted,
        Raw: loaded,
    }

    return converted, loaded, nil
}

func (manager *ImageManager) LoadAnimation(name gameImages.Image) (*Animation, error) {
    loaded, _, err := manager.LoadImage(name)
    if err != nil {
        return nil, err
    }

    switch name {
        case gameImages.ImageExplosion2: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageHit: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageHit2: return NewAnimation(loaded, 5, 6, 1.5, false), nil
        case gameImages.ImageBeam1:
            return NewAnimationCoordinates(loaded, 2, 3, 0.13, []SheetCoordinate{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}, {0, 1}, {2, 0}, {1, 0}}, true), nil
        case gameImages.ImageWave1:
            return NewAnimationCoordinates(loaded, 1, 3, 0.10, []SheetCoordinate{{0, 0}, {1, 0}, {2, 0}, {1, 0}}, true), nil
        case gameImages.ImageRotate1: return NewAnimation(loaded, 2, 2, 0.14, true), nil
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
    Quit context.Context
    Volume float64
}

func (manager *SoundManager) SetVolume(volume float64){
    manager.Volume = volume
}

func (manager *SoundManager) GetVolume() float64 {
    return manager.Volume
}

func MakeSoundManager(quit context.Context, audioContext *audio.Context, volume float64) (*SoundManager, error) {
    manager := SoundManager{
        Sounds: make(map[audioFiles.AudioName]*SoundHandler),
        SampleRate: 48000,
        Context: audioContext,
        Quit: quit,
        Volume: volume,
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

        go handler.Make()
    }

    return nil
}

func (manager *SoundManager) Play(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        player := handler.Make()
        player.SetVolume(manager.GetVolume() / 100.0)
        player.Play()
    }
}

func (manager *SoundManager) PlayLoop(name audioFiles.AudioName) {
    if handler, ok := manager.Sounds[name]; ok {
        go func(){
            player, err := handler.MakeLoop()
            if err != nil {
                log.Printf("Failed to play audio loop %v: %v", name, err)
            } else {
                player.SetVolume(manager.GetVolume() / 100.0)
                go func(){
                    for {
                        select {
                            case <-manager.Quit.Done():
                                player.Close()
                                return
                            case <-time.After(100 * time.Millisecond):
                                if !player.IsPlaying() {
                                    return
                                }

                                player.SetVolume(manager.GetVolume() / 100.0)
                        }
                    }
                }()

                player.Play()
            }
        }()
    }
}

const GameFadeIn = 20
const GameFadeOut = 40

type GameCounter struct {
    Limit int
    Counter int
}

func (counter *GameCounter) Do(f func()) {
    if counter.Counter == 0 {
        f()
        counter.Counter = counter.Limit
    }
}

func (counter *GameCounter) Update() {
    if counter.Counter > 0 {
        counter.Counter -= 1
    }
}

type Game struct {
    Counters map[string]*GameCounter
    Player *Player
    Background *Background
    Bullets []*Bullet
    EnemyBullets []*Bullet
    Font *text.GoTextFaceSource
    Enemies []Enemy
    Explosions []Explosion
    ShaderManager *ShaderManager
    ImageManager *ImageManager
    SoundManager *SoundManager
    FadeIn int
    FadeOut int

    BossMode bool
    // runs one time when the boss should appear
    DoBoss sync.Once
    // runs one time when the level ends
    DoEnd sync.Once
    End atomic.Bool

    MusicPlayer sync.Once

    Quit context.Context
    Cancel context.CancelFunc

    // number of ticks the game has run
    Counter uint64
}

func (game *Game) GetCounter(name string, limit int) *GameCounter {
    use, ok := game.Counters[name]
    if ok {
        return use
    }

    counter := GameCounter{
        Counter: 0,
        Limit: limit,
    }

    game.Counters[name] = &counter

    return game.Counters[name]
}

func (game *Game) Close() {
    game.Cancel()
}

func (game *Game) MakeEnemy(x float64, y float64, kind int, move Movement) error {
    var enemy Enemy
    var err error

    switch kind {
        case 0:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy1)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy1(x, y, raw, pic, move)
        case 1:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy2)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 2:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy3)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 3:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy4)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)
        case 4:
            pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageEnemy5)
            if err != nil {
                return err
            }
            enemy, err = MakeEnemy2(x, y, raw, pic, move)

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
        y := float64(-200)
        kind := rand.Intn(5)

        move := makeMovement()

        for coord := range generator {
            err := game.MakeEnemy(x + coord.x, y + coord.y, kind, move.Copy())
            if err != nil {
                return err
            }
        }
    }

    return nil
}

var LevelEnd error = errors.New("end of level")

func (game *Game) UpdateCounters() {
}

func (game *Game) Update(run *Run) error {

    game.UpdateCounters()

    game.Counter += 1

    if game.End.Load() {
        game.DoEnd.Do(func(){
            game.FadeOut = GameFadeOut * 3
        })
    }

    if game.FadeOut > 0 {
        game.FadeOut -= 1

        if game.FadeOut == 0 {
            return LevelEnd
        }
    }

    if game.FadeIn < GameFadeIn {
        game.FadeIn += 1
    }

    game.MusicPlayer.Do(func(){
        game.SoundManager.PlayLoop(audioFiles.AudioStellarPulseSong)
    })

    game.Background.Update()

    err := game.Player.HandleKeys(game, run)
    if err != nil {
        return err
    }

    game.Player.Move()

    for _, enemy := range game.Enemies {
        bullets := enemy.Move(game.Player, game.ImageManager)
        game.EnemyBullets = append(game.EnemyBullets, bullets...)

        collideX, collideY, isCollide := enemy.CollidePlayer(game.Player)

        if isCollide {
            game.GetCounter("player hit enemy", 30).Do(func(){
                game.SoundManager.Play(audioFiles.AudioHit1)
            })

            animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit)
            if err != nil {
                log.Printf("Could not load hit animation: %v", err)
            } else {
                game.Explosions = append(game.Explosions, MakeAnimatedExplosion(collideX, collideY, animation))
            }

            enemy.Damage(1)

            if ! enemy.IsAlive() {
                game.Player.Score += 1
                game.Player.Kills += 1
                game.SoundManager.Play(audioFiles.AudioExplosion3)

                animation, err := game.ImageManager.LoadAnimation(gameImages.ImageExplosion2)
                if err == nil {
                    x, y := enemy.Coords()
                    game.Explosions = append(game.Explosions, MakeAnimatedExplosion(x, y, animation))
                } else {
                    log.Printf("Could not load explosion sheet: %v", err)
                }
            }
        }
    }

    explosionOut := make([]Explosion, 0)
    for _, explosion := range game.Explosions {
        explosion.Move()
        if explosion.IsAlive() {
            explosionOut = append(explosionOut, explosion)
        }
    }
    game.Explosions = explosionOut

    // run bullet physics at 3x
    for i := 0; i < 3; i++ {
        var outBullets []*Bullet
        for _, bullet := range game.Bullets {
            bullet.Move()

            for _, enemy := range game.Enemies {
                if enemy.IsAlive() && enemy.Collision(bullet.x, bullet.y) {
                    game.Player.Score += 1
                    enemy.Hit(bullet)
                    if ! enemy.IsAlive() {
                        game.Player.Kills += 1
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

                    animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit)
                    if err != nil {
                        log.Printf("Could not load hit animation: %v", err)
                    } else {
                        game.Explosions = append(game.Explosions, MakeAnimatedExplosion(bullet.x, bullet.y, animation))
                    }
                    break
                }
            }

            if bullet.IsAlive() {
                outBullets = append(outBullets, bullet)
            }
        }
        game.Bullets = outBullets

        var outEnemyBullets []*Bullet
        for _, bullet := range game.EnemyBullets {
            bullet.Move()

            if game.Player.Collide(bullet.x, bullet.y) {
                game.SoundManager.Play(audioFiles.AudioHit2)

                animation, err := game.ImageManager.LoadAnimation(gameImages.ImageHit2)
                if err == nil {
                    game.Explosions = append(game.Explosions, MakeAnimatedExplosion(bullet.x, bullet.y, animation))
                } else {
                    log.Printf("Could not load explosion sheet: %v", err)
                }

                bullet.SetDead()
            }

            if bullet.IsAlive() {
                outEnemyBullets = append(outEnemyBullets, bullet)
            }
        }
        game.EnemyBullets = outEnemyBullets
    }

    enemyOut := make([]Enemy, 0)
    for _, enemy := range game.Enemies {
        if enemy.IsAlive() {
            enemyOut = append(enemyOut, enemy)
        }
    }
    game.Enemies = enemyOut

    if !game.BossMode && !game.End.Load(){
        if len(game.Enemies) == 0 || (len(game.Enemies) < 10 && rand.Intn(100) == 0) {
            game.MakeEnemies(1)
        }

        // create the boss after 2 minutes
        const bossTime = 60 * 120
        // const bossTime = 60 * 1
        if debugForceBoss || (game.Counter > bossTime && rand.Intn(1000) == 0) {
            game.BossMode = true
            game.DoBoss.Do(func(){
                log.Printf("Created boss!")
                boss1Pic, rawImage, err := game.ImageManager.LoadImage(gameImages.ImageBoss1)
                if err != nil {
                    log.Printf("Unable to load boss: %v", err)
                } else {
                    boss, err := MakeBoss1(ScreenWidth / 2, -150, rawImage, boss1Pic)
                    if err != nil {
                        log.Printf("Unable to make boss: %v", err)
                    } else {
                        game.Enemies = append(game.Enemies, boss)

                        go func(){
                            for {
                                select {
                                    case <-game.Quit.Done():
                                        return
                                    case <-boss.Dead():
                                        game.End.Store(true)
                                        return
                                }
                            }
                        }()

                    }
                }
            })
        }

    }

    return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
    game.Background.Draw(screen)

    for _, enemy := range game.Enemies {
        enemy.Draw(screen, game.ShaderManager)
    }

    for _, explosion := range game.Explosions {
        explosion.Draw(screen, game.ShaderManager)
    }

    // ebitenutil.DebugPrint(screen, "debugging")
    game.Player.Draw(screen, game.ShaderManager, game.ImageManager, game.Font)

    for _, bullet := range game.Bullets {
        bullet.Draw(screen)
    }

    for _, bullet := range game.EnemyBullets {
        bullet.Draw(screen)
    }

    if game.FadeIn < GameFadeIn {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, &color.RGBA{R: 0, G: 0, B: 0, A: uint8(255 - game.FadeIn * 255 / GameFadeIn)}, true)
    }

    if game.FadeOut > 0 && game.FadeOut <= GameFadeOut {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, &color.RGBA{R: 0, G: 0, B: 0, A: uint8(255 - game.FadeOut * 255 / GameFadeOut)}, true)
    }

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

func premultiplyAlpha(value color.RGBA) color.RGBA {
    a := float32(value.A) / 255.0

    return color.RGBA{
        R: uint8(float32(value.R) * a),
        G: uint8(float32(value.G) * a),
        B: uint8(float32(value.B) * a),
        A: value.A,
    }
}

type RunMode int
const (
    RunGame RunMode = iota
    RunMenu RunMode = iota
)

type Run struct {
    Game *Game
    Menu *Menu
    Mode RunMode
    Volume float64
}

func (run *Run) GetVolume() float64 {
    return run.Volume
}

func (run *Run) updateVolume(){
    if run.Game != nil {
        run.Game.SoundManager.SetVolume(run.Volume)
    }
}

func (run *Run) SetVolume(volume float64){
    run.Volume = volume
    run.updateVolume()
}

func (run *Run) IncreaseVolume() {
    run.Volume += 10
    if run.Volume > 100 {
        run.Volume = 100
    }
    run.updateVolume()
}

func (run *Run) DecreaseVolume() {
    run.Volume -= 10
    if run.Volume < 0 {
        run.Volume = 0
    }
    run.updateVolume()
}

func (run *Run) Update() error {
    switch run.Mode {
        case RunGame:
            err := run.Game.Update(run)
            if errors.Is(err, LevelEnd) {
                run.Game.Close()
                run.Game = nil
                run.Mode = RunMenu
                return nil
            } else {
                return err
            }
        case RunMenu: return run.Menu.Update(run)
    }

    return fmt.Errorf("Unknown mode %v", run.Mode)
}

func (run *Run) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (run *Run) Draw(screen *ebiten.Image) {
    if run.Game != nil {
        run.Game.Draw(screen)
    }

    if run.Mode == RunMenu {
        vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 92}, true)
        run.Menu.Draw(screen)
    }

    /*
    switch run.Mode {
        case RunGame: run.Game.Draw(screen)
        case RunMenu: run.Menu.Draw(screen)
    }
    */
}

func MakeGame(audioContext *audio.Context, run *Run) (*Game, error) {
    player, err := MakePlayer(ScreenWidth / 2, ScreenHeight - 100)
    if err != nil {
        return nil, err
    }

    background, err := MakeBackground()
    if err != nil {
        return nil, err
    }

    font, err := fontLib.LoadFont()
    if err != nil {
        return nil, err
    }

    shaderManager, err := MakeShaderManager()
    if err != nil {
        return nil, err
    }

    quitContext, cancel := context.WithCancel(context.Background())

    soundManager, err := MakeSoundManager(quitContext, audioContext, run.GetVolume())
    if err != nil {
        cancel()
        return nil, err
    }

    game := Game{
        Counters: make(map[string]*GameCounter),
        Background: background,
        Player: player,
        Font: font,
        ShaderManager: shaderManager,
        ImageManager: MakeImageManager(),
        SoundManager: soundManager,
        FadeIn: 0,
        BossMode: false,
        Quit: quitContext,
        Cancel: cancel,
    }

    err = game.MakeEnemies(5)
    if err != nil {
        cancel()
        return nil, err
    }

    err = game.PreloadAssets()
    if err != nil {
        cancel()
        return nil, err
    }

    return &game, nil
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    profile := true

    if profile {
        cpuProfile, err := os.Create("profile.cpu")
        if err != nil {
            log.Printf("Unable to create profile.cpu: %v", err)
        } else {
            defer cpuProfile.Close()
            pprof.StartCPUProfile(cpuProfile)
            defer pprof.StopCPUProfile()
        }
    }

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("Shooter")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    log.Printf("Loading objects")

    audioContext := audio.NewContext(48000)

    menu, err := createMenu(audioContext)
    if err != nil {
        log.Printf("Unable to create menu: %v", err)
        return
    }

    /*
    game, err := MakeGame()
    if err != nil {
        log.Printf("Unable to create game: %v", err)
        return
    }
    */

    run := Run{
        Mode: RunMenu,
        Game: nil,
        Menu: menu,
        Volume: 100,
    }

    log.Printf("Running")
    err = ebiten.RunGame(&run)
    if err != nil {
        log.Printf("Failed to run: %v", err)
    }

    log.Printf("Bye!")

    if profile {
        memProfile, err := os.Create("profile.mem")
        if err != nil {
            log.Printf("Unable to create profile.mem: %v", err)
        } else {
            defer memProfile.Close()
            pprof.WriteHeapProfile(memProfile)
        }
    }
}

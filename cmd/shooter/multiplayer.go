package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"image/color"

	gameImages "github.com/kazzmir/webgl-shooter/images"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	multiplayerRoleMaster = "master"
	multiplayerRoleSlave  = "slave"
	snapshotInterval      = 30
	multiplayerSpawnOffset = 100
)

type playerInputState struct {
	Up        bool `json:"up"`
	Down      bool `json:"down"`
	Left      bool `json:"left"`
	Right     bool `json:"right"`
	Jump      bool `json:"jump"`
	Bomb      bool `json:"bomb"`
	Shoot     bool `json:"shoot"`
	OpenMenu  bool `json:"open_menu"`
	ToggleGun [5]bool `json:"toggle_gun"`
}

type multiplayerEnvelope struct {
	Kind        string            `json:"kind"`
	StartGame   *startGameMessage `json:"start_game,omitempty"`
	Input       *playerInputState `json:"input,omitempty"`
	PlayerState *playerState      `json:"player_state,omitempty"`
	Spawn       *spawnMessage     `json:"spawn,omitempty"`
	Snapshot    *snapshotMessage  `json:"snapshot,omitempty"`
}

type startGameMessage struct {
	Difficulty float64 `json:"difficulty"`
}

type spawnMessage struct {
	ObjectKind string          `json:"object_kind"`
	CreatedAt  uint64          `json:"created_at"`
	State      json.RawMessage `json:"state"`
}

type snapshotMessage struct {
	Counter      uint64          `json:"counter"`
	Difficulty   float64         `json:"difficulty"`
	Player       playerState     `json:"player"`
	Bullets      []bulletState   `json:"bullets"`
	EnemyBullets []bulletState   `json:"enemy_bullets"`
	Asteroids    []asteroidState `json:"asteroids"`
	Enemies      []enemyState    `json:"enemies"`
	Powerups     []powerupState  `json:"powerups"`
	Bombs        []bombState     `json:"bombs"`
	BossMode     bool            `json:"boss_mode"`
	End          bool            `json:"end"`
}

type gameMultiplayer struct {
	Role        string
	Peer        PeerConnector
	RemoteInput playerInputState
}

type playerState struct {
	X             float64    `json:"x"`
	Y             float64    `json:"y"`
	VelocityX     float64    `json:"velocity_x"`
	VelocityY     float64    `json:"velocity_y"`
	Health        float64    `json:"health"`
	MaxHealth     float64    `json:"max_health"`
	GunEnergy     float64    `json:"gun_energy"`
	Bombs         int        `json:"bombs"`
	BombCounter   int        `json:"bomb_counter"`
	PowerupEnergy int        `json:"powerup_energy"`
	Jump          int        `json:"jump"`
	Counter       int        `json:"counter"`
	Score         uint64     `json:"score"`
	Kills         uint64     `json:"kills"`
	Level         int        `json:"level"`
	Experience    float64    `json:"experience"`
	Guns          []gunState `json:"guns"`
}

type gunState struct {
	Kind       string  `json:"kind"`
	Enabled    bool    `json:"enabled"`
	Level      int     `json:"level"`
	Experience float64 `json:"experience"`
	Counter    int     `json:"counter"`
}

type bulletState struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Strength  float64 `json:"strength"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Health    int     `json:"health"`
	Kind      string  `json:"kind"`
}

type asteroidState struct {
	X             float64         `json:"x"`
	Y             float64         `json:"y"`
	VelocityX     float64         `json:"velocity_x"`
	VelocityY     float64         `json:"velocity_y"`
	Rotation      uint64          `json:"rotation"`
	RotationSpeed float64         `json:"rotation_speed"`
	Health        float64         `json:"health"`
	Pic           gameImages.Image `json:"pic"`
}

type bombState struct {
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	VelocityX    float64 `json:"velocity_x"`
	VelocityY    float64 `json:"velocity_y"`
	DestructTime int     `json:"destruct_time"`
	Strength     int     `json:"strength"`
	Radius       float64 `json:"radius"`
	Alpha        int     `json:"alpha"`
}

type powerupState struct {
	Kind      string  `json:"kind"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Activated bool    `json:"activated"`
	Counter   uint64  `json:"counter"`
	Angle     uint64  `json:"angle"`
}

type enemyState struct {
	Kind     string        `json:"kind"`
	X        float64       `json:"x"`
	Y        float64       `json:"y"`
	Life     float64       `json:"life"`
	Flip     bool          `json:"flip"`
	Hurt     int           `json:"hurt"`
	Movement movementState `json:"movement"`
}

type movementState struct {
	Kind      string  `json:"kind"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Amplitude float64 `json:"amplitude"`
	Angle     float64 `json:"angle"`
	Radius    float64 `json:"radius"`
	Speed     float64 `json:"speed"`
	MoveX     float64 `json:"move_x"`
	MoveY     float64 `json:"move_y"`
	Counter   uint64  `json:"counter"`
}

func gatherLocalPlayerInput() playerInputState {
	input := playerInputState{}
	keys := inpututil.AppendPressedKeys(nil)
	for _, key := range keys {
		switch key {
		case ebiten.KeyArrowUp:
			input.Up = true
		case ebiten.KeyArrowDown:
			input.Down = true
		case ebiten.KeyArrowLeft:
			input.Left = true
		case ebiten.KeyArrowRight:
			input.Right = true
		case ebiten.KeyShift:
			input.Jump = true
		case ebiten.KeyB:
			input.Bomb = true
		case ebiten.KeySpace:
			input.Shoot = true
		}
	}

	for _, key := range inpututil.AppendJustPressedKeys(nil) {
		switch key {
		case ebiten.KeyEscape, ebiten.KeyCapsLock:
			input.OpenMenu = true
		case ebiten.KeyDigit1:
			input.ToggleGun[0] = true
		case ebiten.KeyDigit2:
			input.ToggleGun[1] = true
		case ebiten.KeyDigit3:
			input.ToggleGun[2] = true
		case ebiten.KeyDigit4:
			input.ToggleGun[3] = true
		case ebiten.KeyDigit5:
			input.ToggleGun[4] = true
		}
	}

	return input
}

func (player *Player) ApplyInput(game *Game, run *Run, input playerInputState, allowProjectiles bool) error {
	maxVelocity := 3.8
	playerAccel := 0.9
	if player.Jump > 0 {
		playerAccel = 3
		maxVelocity = 5.5
	}

	if input.Up {
		player.velocityY -= playerAccel
	}
	if input.Down {
		player.velocityY += playerAccel
	}
	if input.Left {
		player.velocityX -= playerAccel
	}
	if input.Right {
		player.velocityX += playerAccel
	}
	if input.Jump && player.Jump <= -50 {
		player.Jump = JumpDuration
	}
	if allowProjectiles && input.Bomb && player.BombCounter == 0 && player.Bombs > 0 {
		bomb := MakeBomb(player.x, player.y-20, 0, -1.8)
		game.AddBomb(bomb)
		player.Bombs -= 1
		player.BombCounter = BombDelay
	}
	if allowProjectiles && input.Shoot {
		game.AddPlayerBullets(game.Player.Shoot(game.ImageManager, game.SoundManager)...)
	}

	player.velocityX = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityX))
	player.velocityY = math.Min(maxVelocity, math.Max(-maxVelocity, player.velocityY))

	if input.OpenMenu {
		run.Mode = RunMenu
	}
	for i, pressed := range input.ToggleGun {
		if pressed {
			enableGun(player.Guns, i)
		}
	}

	if player.Jump > -50 {
		player.Jump -= 1
	}

	return nil
}

func (run *Run) StartGame(role string, notifyPeer bool) error {
	run.Mode = RunGame

	if run.Game != nil {
		run.Game.Cancel()
	}

	player, err := MakePlayer(0, 0, run.Cheats)
	if err != nil {
		return err
	}
	run.Player = player

	game, err := MakeGame(run.SoundManager, run, 1)
	if err != nil {
		return err
	}

	if role != "" {
		remotePlayer, err := MakePlayer(0, 0, false)
		if err != nil {
			return err
		}
		game.Multiplayer = &gameMultiplayer{
			Role: role,
			Peer: run.PeerConnector,
		}
		game.RemotePlayer = remotePlayer
		baseX := float64(LogicalWidth) / 2
		baseY := float64(ScreenHeight - 100)
		if role == multiplayerRoleMaster {
			game.Player.x = baseX - multiplayerSpawnOffset
			game.RemotePlayer.x = baseX + multiplayerSpawnOffset
		} else {
			game.Player.x = baseX + multiplayerSpawnOffset
			game.RemotePlayer.x = baseX - multiplayerSpawnOffset
		}
		game.Player.y = baseY
		game.RemotePlayer.y = baseY
		if role == multiplayerRoleSlave {
			game.Enemies = nil
			game.Asteroids = nil
			game.Powerups = nil
			game.Bullets = nil
			game.EnemyBullets = nil
			game.Bombs = nil
		}
		if notifyPeer && role == multiplayerRoleMaster && run.PeerConnector != nil {
			if err := run.PeerConnector.SendGameMessage(multiplayerEnvelope{
				Kind: "start_game",
				StartGame: &startGameMessage{Difficulty: game.Difficulty},
			}); err != nil {
				log.Printf("Unable to send start game message: %v", err)
			}
			game.maybeSendSnapshot()
		}
	}

	run.Game = game
	return nil
}

func (run *Run) handleMenuMultiplayerMessages(messages [][]byte) error {
	for _, raw := range messages {
		var envelope multiplayerEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			log.Printf("Unable to decode peer message: %v", err)
			continue
		}
		if envelope.Kind == "start_game" && run.PeerConnector != nil && run.PeerConnector.IsSlave() {
			return run.StartGame(multiplayerRoleSlave, false)
		}
	}
	return nil
}

func (game *Game) isMaster() bool {
	return game.Multiplayer != nil && game.Multiplayer.Role == multiplayerRoleMaster
}

func (game *Game) isSlave() bool {
	return game.Multiplayer != nil && game.Multiplayer.Role == multiplayerRoleSlave
}

func (game *Game) resolvePlayerInput() playerInputState {
	return gatherLocalPlayerInput()
}

func (game *Game) processNetworkMessages(run *Run, messages [][]byte) error {
	for _, raw := range messages {
		var envelope multiplayerEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			log.Printf("Unable to decode gameplay message: %v", err)
			continue
		}

		switch envelope.Kind {
		case "player_state":
			if envelope.PlayerState != nil && game.RemotePlayer != nil {
				applyPlayerState(game.RemotePlayer, *envelope.PlayerState)
			}
		case "snapshot":
			if game.isSlave() && envelope.Snapshot != nil {
				if err := game.applySnapshot(*envelope.Snapshot); err != nil {
					return err
				}
			}
		case "spawn":
			if game.isSlave() && envelope.Spawn != nil {
				if err := game.applySpawn(*envelope.Spawn); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (game *Game) maybeSendSnapshot() {
	if !game.isMaster() || game.Counter%snapshotInterval != 0 {
		return
	}

	snapshot := snapshotMessage{
		Counter:      game.Counter,
		Difficulty:   game.Difficulty,
		Player:       serializePlayer(game.Player),
		Bullets:      serializeBullets(game.Bullets),
		EnemyBullets: serializeBullets(game.EnemyBullets),
		Asteroids:    serializeAsteroids(game.Asteroids),
		Enemies:      serializeEnemies(game.Enemies),
		Powerups:     serializePowerups(game.Powerups),
		Bombs:        serializeBombs(game.Bombs),
		BossMode:     game.BossMode,
		End:          game.End.Load(),
	}

	if err := game.Multiplayer.Peer.SendGameMessage(multiplayerEnvelope{
		Kind:     "snapshot",
		Snapshot: &snapshot,
	}); err != nil && game.Counter%120 == 0 {
		log.Printf("Unable to send snapshot: %v", err)
	}
}

func (game *Game) maybeSendPlayerState() {
	if game.Multiplayer == nil || game.Multiplayer.Peer == nil {
		return
	}

	state := serializePlayer(game.Player)
	if err := game.Multiplayer.Peer.SendGameMessage(multiplayerEnvelope{
		Kind:        "player_state",
		PlayerState: &state,
	}); err != nil && game.Counter%120 == 0 {
		log.Printf("Unable to send player state: %v", err)
	}
}

func (game *Game) AddPlayerBullets(bullets ...*Bullet) {
	if len(bullets) == 0 {
		return
	}
	game.Bullets = append(game.Bullets, bullets...)
	if game.isMaster() {
		for _, bullet := range bullets {
			game.sendSpawn("bullet", game.Counter, serializeBullet(bullet))
		}
	}
}

func (game *Game) AddEnemyBullets(bullets ...*Bullet) {
	if len(bullets) == 0 {
		return
	}
	game.EnemyBullets = append(game.EnemyBullets, bullets...)
	if game.isMaster() {
		for _, bullet := range bullets {
			game.sendSpawn("enemy_bullet", game.Counter, serializeBullet(bullet))
		}
	}
}

func (game *Game) AddBomb(bomb *Bomb) {
	if bomb == nil {
		return
	}
	game.Bombs = append(game.Bombs, bomb)
	if game.isMaster() {
		game.sendSpawn("bomb", game.Counter, serializeBomb(bomb))
	}
}

func (game *Game) AddPowerup(powerup Powerup) {
	if powerup == nil {
		return
	}
	game.Powerups = append(game.Powerups, powerup)
	if game.isMaster() {
		game.sendSpawn("powerup", game.Counter, serializePowerup(powerup))
	}
}

func (game *Game) AddAsteroid(asteroid *Asteroid) {
	if asteroid == nil {
		return
	}
	game.Asteroids = append(game.Asteroids, asteroid)
	if game.isMaster() {
		game.sendSpawn("asteroid", game.Counter, serializeAsteroid(asteroid))
	}
}

func (game *Game) AddEnemy(enemy Enemy) {
	if enemy == nil {
		return
	}
	game.Enemies = append(game.Enemies, enemy)
	if game.isMaster() {
		game.sendSpawn("enemy", game.Counter, serializeEnemy(enemy))
	}
}

func (game *Game) sendSpawn(kind string, createdAt uint64, state any) {
	if game.Multiplayer == nil || game.Multiplayer.Peer == nil {
		return
	}
	payload, err := json.Marshal(state)
	if err != nil {
		log.Printf("Unable to encode spawn %s: %v", kind, err)
		return
	}
	if err := game.Multiplayer.Peer.SendGameMessage(multiplayerEnvelope{
		Kind: "spawn",
		Spawn: &spawnMessage{
			ObjectKind: kind,
			CreatedAt:  createdAt,
			State:      payload,
		},
	}); err != nil && game.Counter%120 == 0 {
		log.Printf("Unable to send spawn %s: %v", kind, err)
	}
}

func (game *Game) applySpawn(spawn spawnMessage) error {
	switch spawn.ObjectKind {
	case "bullet":
		var state bulletState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		game.Bullets = append(game.Bullets, game.makeBulletFromState(state))
	case "enemy_bullet":
		var state bulletState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		game.EnemyBullets = append(game.EnemyBullets, game.makeBulletFromState(state))
	case "bomb":
		var state bombState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		game.Bombs = append(game.Bombs, makeBombFromState(state))
	case "powerup":
		var state powerupState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		powerup, err := makePowerupFromState(state)
		if err != nil {
			return err
		}
		game.Powerups = append(game.Powerups, powerup)
	case "asteroid":
		var state asteroidState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		game.Asteroids = append(game.Asteroids, makeAsteroidFromState(state))
	case "enemy":
		var state enemyState
		if err := json.Unmarshal(spawn.State, &state); err != nil {
			return err
		}
		enemy, err := game.makeEnemyFromState(state)
		if err != nil {
			return err
		}
		game.Enemies = append(game.Enemies, enemy)
	}
	return nil
}

func (game *Game) applySnapshot(snapshot snapshotMessage) error {
	if game.RemotePlayer != nil {
		applyPlayerState(game.RemotePlayer, snapshot.Player)
	}
	game.Counter = snapshot.Counter
	game.Difficulty = snapshot.Difficulty
	game.BossMode = snapshot.BossMode
	game.End.Store(snapshot.End)

	game.Bullets = make([]*Bullet, 0, len(snapshot.Bullets))
	for _, bullet := range snapshot.Bullets {
		game.Bullets = append(game.Bullets, game.makeBulletFromState(bullet))
	}

	game.EnemyBullets = make([]*Bullet, 0, len(snapshot.EnemyBullets))
	for _, bullet := range snapshot.EnemyBullets {
		game.EnemyBullets = append(game.EnemyBullets, game.makeBulletFromState(bullet))
	}

	game.Asteroids = make([]*Asteroid, 0, len(snapshot.Asteroids))
	for _, asteroid := range snapshot.Asteroids {
		game.Asteroids = append(game.Asteroids, makeAsteroidFromState(asteroid))
	}

	game.Powerups = make([]Powerup, 0, len(snapshot.Powerups))
	for _, powerupState := range snapshot.Powerups {
		powerup, err := makePowerupFromState(powerupState)
		if err != nil {
			return err
		}
		game.Powerups = append(game.Powerups, powerup)
	}

	game.Bombs = make([]*Bomb, 0, len(snapshot.Bombs))
	for _, bomb := range snapshot.Bombs {
		game.Bombs = append(game.Bombs, makeBombFromState(bomb))
	}

	game.Enemies = make([]Enemy, 0, len(snapshot.Enemies))
	for _, enemyState := range snapshot.Enemies {
		enemy, err := game.makeEnemyFromState(enemyState)
		if err != nil {
			return err
		}
		game.Enemies = append(game.Enemies, enemy)
	}

	return nil
}

func serializePlayer(player *Player) playerState {
	return playerState{
		X:             player.x,
		Y:             player.y,
		VelocityX:     player.velocityX,
		VelocityY:     player.velocityY,
		Health:        player.Health,
		MaxHealth:     player.MaxHealth,
		GunEnergy:     player.GunEnergy,
		Bombs:         player.Bombs,
		BombCounter:   player.BombCounter,
		PowerupEnergy: player.PowerupEnergy,
		Jump:          player.Jump,
		Counter:       player.Counter,
		Score:         player.Score,
		Kills:         player.Kills,
		Level:         player.Level,
		Experience:    player.Experience,
		Guns:          serializeGuns(player.Guns),
	}
}

func applyPlayerState(player *Player, state playerState) {
	player.x = state.X
	player.y = state.Y
	player.velocityX = state.VelocityX
	player.velocityY = state.VelocityY
	player.Health = state.Health
	player.MaxHealth = state.MaxHealth
	player.GunEnergy = state.GunEnergy
	player.Bombs = state.Bombs
	player.BombCounter = state.BombCounter
	player.PowerupEnergy = state.PowerupEnergy
	player.Jump = state.Jump
	player.Counter = state.Counter
	player.Score = state.Score
	player.Kills = state.Kills
	player.Level = state.Level
	player.Experience = state.Experience
	player.Guns = makeGunsFromState(state.Guns)
}

func serializeGuns(guns []Gun) []gunState {
	out := make([]gunState, 0, len(guns))
	for _, gun := range guns {
		switch current := gun.(type) {
		case *BasicGun:
			out = append(out, gunState{Kind: "basic", Enabled: current.enabled, Level: current.level, Experience: current.experience, Counter: current.counter})
		case *BeamGun:
			out = append(out, gunState{Kind: "beam", Enabled: current.enabled, Level: current.level, Experience: current.experience, Counter: current.counter})
		case *MissleGun:
			out = append(out, gunState{Kind: "missile", Enabled: current.enabled, Level: current.level, Experience: current.experience, Counter: current.counter})
		case *LightningGun:
			out = append(out, gunState{Kind: "lightning", Enabled: current.enabled, Level: current.level, Experience: current.experience, Counter: current.counter})
		case *DualBasicGun:
			out = append(out, gunState{Kind: "dual-basic", Enabled: current.enabled, Level: current.level, Experience: current.experience, Counter: current.counter})
		}
	}
	return out
}

func makeGunsFromState(states []gunState) []Gun {
	out := make([]Gun, 0, len(states))
	for _, state := range states {
		switch state.Kind {
		case "basic":
			out = append(out, &BasicGun{enabled: state.Enabled, level: state.Level, experience: state.Experience, counter: state.Counter})
		case "beam":
			out = append(out, &BeamGun{enabled: state.Enabled, level: state.Level, experience: state.Experience, counter: state.Counter})
		case "missile":
			out = append(out, &MissleGun{enabled: state.Enabled, level: state.Level, experience: state.Experience, counter: state.Counter})
		case "lightning":
			out = append(out, &LightningGun{enabled: state.Enabled, level: state.Level, experience: state.Experience, counter: state.Counter})
		case "dual-basic":
			out = append(out, &DualBasicGun{enabled: state.Enabled, level: state.Level, experience: state.Experience, counter: state.Counter})
		}
	}
	return out
}

func serializeBullet(bullet *Bullet) bulletState {
	return bulletState{
		X:         bullet.x,
		Y:         bullet.y,
		Strength:  bullet.Strength,
		VelocityX: bullet.velocityX,
		VelocityY: bullet.velocityY,
		Health:    bullet.health,
		Kind:      bulletKind(bullet),
	}
}

func serializeBullets(bullets []*Bullet) []bulletState {
	out := make([]bulletState, 0, len(bullets))
	for _, bullet := range bullets {
		out = append(out, serializeBullet(bullet))
	}
	return out
}

func bulletKind(bullet *Bullet) string {
	switch bullet.Gun.(type) {
	case *BeamGun:
		return "beam"
	case *MissleGun:
		return "missile"
	case *LightningGun:
		return "lightning"
	case *BasicGun, *DualBasicGun:
		return "basic"
	}

	if bullet.animation != nil {
		return "enemy-rotate"
	}
	if bullet.pic != nil {
		return "enemy-aim"
	}
	return "basic"
}

func (game *Game) makeBulletFromState(state bulletState) *Bullet {
	bullet := &Bullet{
		x:         state.X,
		y:         state.Y,
		Strength:  state.Strength,
		velocityX: state.VelocityX,
		velocityY: state.VelocityY,
		health:    state.Health,
	}

	switch state.Kind {
	case "beam":
		animation, err := game.ImageManager.LoadAnimation(gameImages.ImageBeam1)
		if err == nil {
			bullet.animation = animation
		}
	case "missile":
		pic, _, err := game.ImageManager.LoadImage(gameImages.ImageMissle1)
		if err == nil {
			bullet.pic = pic
		}
	case "enemy-rotate":
		animation, err := game.ImageManager.LoadAnimation(gameImages.ImageRotate1)
		if err == nil {
			bullet.animation = animation
		}
	case "enemy-aim":
		pic, _, err := game.ImageManager.LoadImage(gameImages.ImageBulletSmallBlue)
		if err == nil {
			bullet.pic = pic
		}
	case "lightning":
		pic := ebiten.NewImage(3, 3)
		pic.Fill(color.White)
		bullet.pic = pic
	default:
		pic, _, err := game.ImageManager.LoadImage(gameImages.ImageBullet)
		if err == nil {
			bullet.pic = pic
		}
	}

	return bullet
}

func serializeAsteroid(asteroid *Asteroid) asteroidState {
	return asteroidState{
		X: asteroid.x, Y: asteroid.y,
		VelocityX: asteroid.velocityX, VelocityY: asteroid.velocityY,
		Rotation: asteroid.rotation, RotationSpeed: asteroid.rotationSpeed,
		Health: asteroid.health, Pic: asteroid.pic,
	}
}

func serializeAsteroids(asteroids []*Asteroid) []asteroidState {
	out := make([]asteroidState, 0, len(asteroids))
	for _, asteroid := range asteroids {
		out = append(out, serializeAsteroid(asteroid))
	}
	return out
}

func makeAsteroidFromState(state asteroidState) *Asteroid {
	return &Asteroid{
		x: state.X, y: state.Y,
		velocityX: state.VelocityX, velocityY: state.VelocityY,
		rotation: state.Rotation, rotationSpeed: state.RotationSpeed,
		health: state.Health, pic: state.Pic,
	}
}

func serializeBomb(bomb *Bomb) bombState {
	return bombState{
		X: bomb.x, Y: bomb.y,
		VelocityX: bomb.velocityX, VelocityY: bomb.velocityY,
		DestructTime: bomb.destructTime, Strength: bomb.strength,
		Radius: bomb.radius, Alpha: bomb.alpha,
	}
}

func serializeBombs(bombs []*Bomb) []bombState {
	out := make([]bombState, 0, len(bombs))
	for _, bomb := range bombs {
		out = append(out, serializeBomb(bomb))
	}
	return out
}

func makeBombFromState(state bombState) *Bomb {
	return &Bomb{
		x: state.X, y: state.Y,
		velocityX: state.VelocityX, velocityY: state.VelocityY,
		destructTime: state.DestructTime, strength: state.Strength,
		radius: state.Radius, alpha: state.Alpha,
	}
}

func serializePowerup(powerup Powerup) powerupState {
	switch current := powerup.(type) {
	case *PowerupEnergy:
		return powerupState{Kind: "energy", X: current.x, Y: current.y, VelocityX: current.velocityX, VelocityY: current.velocityY, Activated: current.activated, Angle: current.angle}
	case *PowerupHealth:
		return powerupState{Kind: "health", X: current.x, Y: current.y, VelocityX: current.velocityX, VelocityY: current.velocityY, Activated: current.activated, Counter: current.counter}
	case *PowerupWeapon:
		return powerupState{Kind: "weapon", X: current.x, Y: current.y, VelocityX: current.velocityX, VelocityY: current.velocityY, Activated: current.activated, Counter: current.counter}
	case *PowerupBomb:
		return powerupState{Kind: "bomb", X: current.x, Y: current.y, VelocityX: current.velocityX, VelocityY: current.velocityY, Activated: current.activated, Counter: current.counter}
	default:
		return powerupState{}
	}
}

func serializePowerups(powerups []Powerup) []powerupState {
	out := make([]powerupState, 0, len(powerups))
	for _, powerup := range powerups {
		out = append(out, serializePowerup(powerup))
	}
	return out
}

func makePowerupFromState(state powerupState) (Powerup, error) {
	switch state.Kind {
	case "energy":
		return &PowerupEnergy{x: state.X, y: state.Y, velocityX: state.VelocityX, velocityY: state.VelocityY, activated: state.Activated, angle: state.Angle}, nil
	case "health":
		return &PowerupHealth{x: state.X, y: state.Y, velocityX: state.VelocityX, velocityY: state.VelocityY, activated: state.Activated, counter: state.Counter}, nil
	case "weapon":
		return &PowerupWeapon{x: state.X, y: state.Y, velocityX: state.VelocityX, velocityY: state.VelocityY, activated: state.Activated, counter: state.Counter}, nil
	case "bomb":
		return &PowerupBomb{x: state.X, y: state.Y, velocityX: state.VelocityX, velocityY: state.VelocityY, activated: state.Activated, counter: state.Counter}, nil
	default:
		return nil, fmt.Errorf("unknown powerup kind %q", state.Kind)
	}
}

func serializeEnemy(enemy Enemy) enemyState {
	current, ok := enemy.(*NormalEnemy)
	if !ok {
		return enemyState{}
	}
	return enemyState{
		Kind: current.Kind,
		X: current.x,
		Y: current.y,
		Life: current.Life,
		Flip: current.Flip,
		Hurt: current.hurt,
		Movement: serializeMovement(current.move),
	}
}

func serializeEnemies(enemies []Enemy) []enemyState {
	out := make([]enemyState, 0, len(enemies))
	for _, enemy := range enemies {
		out = append(out, serializeEnemy(enemy))
	}
	return out
}

func serializeMovement(move Movement) movementState {
	switch current := move.(type) {
	case *LinearMovement:
		return movementState{Kind: "linear", VelocityX: current.velocityX, VelocityY: current.velocityY}
	case *SineMovement:
		return movementState{Kind: "sine", VelocityX: current.velocityX, VelocityY: current.velocityY, Amplitude: current.amplitude, Angle: current.angle}
	case *CircularMovement:
		return movementState{Kind: "circular", VelocityX: current.velocityX, VelocityY: current.velocityY, Radius: current.radius, Angle: float64(current.angle), Speed: current.speed}
	case *Boss1Movement:
		return movementState{Kind: "boss1", MoveX: current.moveX, MoveY: current.moveY, Counter: current.counter}
	default:
		return movementState{}
	}
}

func makeMovementFromState(state movementState) Movement {
	switch state.Kind {
	case "linear":
		return &LinearMovement{velocityX: state.VelocityX, velocityY: state.VelocityY}
	case "sine":
		return &SineMovement{velocityX: state.VelocityX, velocityY: state.VelocityY, amplitude: state.Amplitude, angle: state.Angle}
	case "circular":
		return &CircularMovement{velocityX: state.VelocityX, velocityY: state.VelocityY, radius: state.Radius, angle: uint64(state.Angle), speed: state.Speed}
	case "boss1":
		return &Boss1Movement{moveX: state.MoveX, moveY: state.MoveY, counter: state.Counter}
	default:
		return &LinearMovement{}
	}
}

func (game *Game) makeEnemyFromState(state enemyState) (Enemy, error) {
	move := makeMovementFromState(state.Movement)
	switch state.Kind {
	case "enemy-0", "enemy-1", "enemy-2", "enemy-3", "enemy-4", "enemy-5", "enemy-6", "enemy-7", "enemy-8":
		enemyKind := 0
		fmt.Sscanf(state.Kind, "enemy-%d", &enemyKind)
		var imageName gameImages.Image
		switch enemyKind {
		case 0:
			imageName = gameImages.ImageEnemy1
		case 1:
			imageName = gameImages.ImageEnemy2
		case 2:
			imageName = gameImages.ImageEnemy3
		case 3:
			imageName = gameImages.ImageEnemy4
		case 4:
			imageName = gameImages.ImageEnemy5
		case 5:
			imageName = gameImages.ImageEnemy6
		case 6:
			imageName = gameImages.ImageEnemy7
		case 7:
			imageName = gameImages.ImageEnemy8
		default:
			imageName = gameImages.ImageEnemy9
		}
		pic, raw, err := game.ImageManager.LoadImage(imageName)
		if err != nil {
			return nil, err
		}
		var enemy Enemy
		if enemyKind == 0 {
			enemy, err = MakeEnemy1(state.X, state.Y, raw, pic, move, game.Difficulty)
		} else {
			enemy, err = MakeEnemy2(state.X, state.Y, raw, pic, move, game.Difficulty)
		}
		if err != nil {
			return nil, err
		}
		current := enemy.(*NormalEnemy)
		current.Kind = state.Kind
		current.Life = state.Life
		current.hurt = state.Hurt
		current.Flip = state.Flip
		return current, nil
	case "boss1":
		pic, raw, err := game.ImageManager.LoadImage(gameImages.ImageBoss1)
		if err != nil {
			return nil, err
		}
		enemy, err := MakeBoss1(state.X, state.Y, raw, pic, game.Difficulty)
		if err != nil {
			return nil, err
		}
		current := enemy.(*NormalEnemy)
		current.move = move
		current.Life = state.Life
		current.hurt = state.Hurt
		current.Flip = state.Flip
		return current, nil
	default:
		return nil, fmt.Errorf("unknown enemy kind %q", state.Kind)
	}
}

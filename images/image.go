package image

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

type Image string

const ImagePlayer = Image("player")
const ImageStar1 = Image("star1")
const ImageStar2 = Image("star2")
const ImagePlanet = Image("planet")
const ImageGalaxy = Image("galaxy")
const ImageBullet = Image("bullet")
const ImageEnemy1 = Image("enemy1")
const ImageEnemy2 = Image("enemy2")
const ImageEnemy3 = Image("enemy3")
const ImageEnemy4 = Image("enemy4")
const ImageEnemy5 = Image("enemy5")
const ImageEnemy6 = Image("enemy6")
const ImageEnemy7 = Image("enemy7")
const ImageEnemy8 = Image("enemy8")
const ImageEnemy9 = Image("enemy9")
const ImageAsteroid1 = Image("asteroid1")
const ImageAsteroid2 = Image("asteroid2")
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
const ImageBomb = Image("bomb")
const ImagePowerup1 = Image("powerup1")
const ImagePowerup2 = Image("powerup2")
const ImagePowerup3 = Image("powerup3")
const ImagePowerup4 = Image("powerup4")
const ImagePowerup5 = Image("powerup5")
const ImagePowerupBomb = Image("powerup-bomb")
const ImageLightningIcon = Image("lightning")

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

//go:embed background/galaxy3.jpg
var galaxyImage []byte

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

//go:embed enemy/enemy6.png
var enemy6Image []byte

//go:embed enemy/enemy7.png
var enemy7Image []byte

//go:embed enemy/enemy8.png
var enemy8Image []byte

//go:embed enemy/enemy9.png
var enemy9Image []byte

//go:embed enemy/boss1.png
var boss1Image []byte

//go:embed bullet/bomb.png
var bombImage []byte

//go:embed misc/asteroid1.png
var asteroid1Image []byte

//go:embed misc/asteroid2.png
var asteroid2Image []byte

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

//go:embed misc/powerup5.png
var powerup5Image []byte

//go:embed misc/powerup-bomb.png
var powerupBombImage []byte

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

//go:embed player/bolt1.png
var lightningImage []byte

func loadEmbeddedImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func LoadImage(name Image) (image.Image, error) {
	switch name {
	case ImagePlayer:
		return loadEmbeddedImage(playerImage)
	case ImageStar1:
		return loadEmbeddedImage(star1Image)
	case ImageStar2:
		return loadEmbeddedImage(star2Image)
	case ImagePlanet:
		return loadEmbeddedImage(planetImage)
	case ImageGalaxy:
		return loadEmbeddedImage(galaxyImage)
	case ImageBullet:
		return loadEmbeddedImage(bulletImage)
	case ImageEnemy1:
		return loadEmbeddedImage(enemy1Image)
	case ImageEnemy2:
		return loadEmbeddedImage(enemy2Image)
	case ImageEnemy3:
		return loadEmbeddedImage(enemy3Image)
	case ImageEnemy4:
		return loadEmbeddedImage(enemy4Image)
	case ImageEnemy5:
		return loadEmbeddedImage(enemy5Image)
	case ImageEnemy6:
		return loadEmbeddedImage(enemy6Image)
	case ImageEnemy7:
		return loadEmbeddedImage(enemy7Image)
	case ImageEnemy8:
		return loadEmbeddedImage(enemy8Image)
	case ImageEnemy9:
		return loadEmbeddedImage(enemy9Image)
	case ImageBoss1:
		return loadEmbeddedImage(boss1Image)
	case ImageExplosion1:
		return loadEmbeddedImage(explosion1Image)
	case ImageExplosion2:
		return loadEmbeddedImage(explosion2Animation)
	case ImageExplosion3:
		return loadEmbeddedImage(explosion3Animation)
	case ImageHit:
		return loadEmbeddedImage(hitImage)
	case ImageBeam1:
		return loadEmbeddedImage(beam1Image)
	case ImageWave1:
		return loadEmbeddedImage(wave1Image)
	case ImageRotate1:
		return loadEmbeddedImage(rotate1Image)
	case ImageHit2:
		return loadEmbeddedImage(hit2Image)
	case ImageMissle1:
		return loadEmbeddedImage(missle1Image)
	case ImageBulletSmallBlue:
		return loadEmbeddedImage(bulletSmallBlueImage)
	case ImagePowerup1:
		return loadEmbeddedImage(powerup1Image)
	case ImagePowerup2:
		return loadEmbeddedImage(powerup2Image)
	case ImagePowerup3:
		return loadEmbeddedImage(powerup3Image)
	case ImagePowerup4:
		return loadEmbeddedImage(powerup4Image)
	case ImagePowerup5:
		return loadEmbeddedImage(powerup5Image)
	case ImageAsteroid1:
		return loadEmbeddedImage(asteroid1Image)
	case ImageAsteroid2:
		return loadEmbeddedImage(asteroid2Image)
	case ImageFire1:
		return loadEmbeddedImage(fire1Animation)
	case ImageBomb:
		return loadEmbeddedImage(bombImage)
	case ImagePowerupBomb:
		return loadEmbeddedImage(powerupBombImage)
	case ImageLightningIcon:
		return loadEmbeddedImage(lightningImage)
	}

	return nil, fmt.Errorf("no such image: %s", name)
}

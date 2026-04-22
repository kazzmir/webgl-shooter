package main

import (
	"testing"
)

func TestGunElementTypes(t *testing.T) {
	tests := []struct {
		name string
		gun  Gun
		want ElementType
	}{
		{name: "basic", gun: &BasicGun{}, want: ElementPhysical},
		{name: "dual-basic", gun: &DualBasicGun{}, want: ElementPhysical},
		{name: "beam", gun: &BeamGun{}, want: ElementPlasma},
		{name: "missile", gun: &MissleGun{}, want: ElementPhysical},
		{name: "lightning", gun: &LightningGun{}, want: ElementLightning},
	}

	for _, test := range tests {
		if got := test.gun.ElementType(); got != test.want {
			t.Fatalf("%s element type mismatch: got %q want %q", test.name, got, test.want)
		}
	}
}

func TestSerializeBulletPreservesElementType(t *testing.T) {
	bullet := &Bullet{
		Strength:    1,
		health:      1,
		Gun:         &BeamGun{},
		ElementType: ElementPlasma,
	}

	state := serializeBullet(bullet)
	if state.ElementType != ElementPlasma {
		t.Fatalf("serialized element type mismatch: got %q want %q", state.ElementType, ElementPlasma)
	}
}

func TestLightningGunShootWithSeedIsDeterministic(t *testing.T) {
	gun := &LightningGun{
		level:       4,
		enabled:     true,
		elementType: ElementLightning,
	}

	left, err := gun.ShootWithSeed(nil, 100, 200, 12345)
	if err != nil {
		t.Fatalf("first ShootWithSeed failed: %v", err)
	}
	right, err := gun.ShootWithSeed(nil, 100, 200, 12345)
	if err != nil {
		t.Fatalf("second ShootWithSeed failed: %v", err)
	}

	if len(left) != len(right) {
		t.Fatalf("bullet count mismatch: %d != %d", len(left), len(right))
	}
	if len(left) == 0 {
		t.Fatal("expected lightning bullets")
	}

	for i := range left {
		if left[i].x != right[i].x || left[i].y != right[i].y {
			t.Fatalf("bullet %d mismatch: (%v,%v) != (%v,%v)", i, left[i].x, left[i].y, right[i].x, right[i].y)
		}
		if left[i].RemainingLife != right[i].RemainingLife {
			t.Fatalf("bullet %d life mismatch: %d != %d", i, left[i].RemainingLife, right[i].RemainingLife)
		}
	}
}

// benchmark creation of lightning
func BenchmarkCreateLightning(bench *testing.B) {
	rng := newLightningRand(1)
	for bench.Loop() {
		makeLightningSegments(rng, 1, 1, 300, 300, 0.9, 20, 100)
	}
}

func BenchmarkLightningGun(bench *testing.B) {
	gun := LightningGun{
		level:   0,
		enabled: true,
	}

	for bench.Loop() {
		gun.Shoot(nil, 0, 0)
		gun.counter = 0
	}
}

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

// benchmark creation of lightning
func BenchmarkCreateLightning(bench *testing.B) {
	for bench.Loop() {
		makeLightningSegments(1, 1, 300, 300, 0.9, 20, 100)
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

package main

import (
    "testing"
)

// benchmark creation of lightning
func BenchmarkCreateLightning(bench *testing.B) {
    for bench.Loop() {
        makeLightningSegments(1, 1, 300, 300, 0.9, 20, 100)
    }
}

func BenchmarkLightningGun(bench *testing.B) {
    gun := LightningGun{
        level: 0,
        enabled: true,
    }

    for bench.Loop() {
        gun.Shoot(nil, 0, 0)
        gun.counter = 0
    }
}

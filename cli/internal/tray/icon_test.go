package tray

import "testing"

func TestGaragePixel(t *testing.T) {
	for _, c := range []struct {
		x, y int
		in   bool
	}{
		{16, 4, true},   // under the apex
		{2, 4, false},   // beside the apex
		{2, 13, true},   // eave
		{5, 20, true},   // wall
		{16, 17, true},  // a slat
		{16, 19, false}, // a gap between slats
		{16, 30, false}, // below the garage
		{0, 0, false},
	} {
		if got := garagePixel(c.x, c.y); got != c.in {
			t.Errorf("garagePixel(%d, %d) = %v, want %v", c.x, c.y, got, c.in)
		}
	}
}

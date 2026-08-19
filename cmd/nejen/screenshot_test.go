package main

import "testing"

// The rects feed slurp, so a wrong number here means a screenshot region that
// silently misses part of the screen. These cases mirror what the jq
// implementation produced before the geometry moved into Go.
func TestLogicalRect(t *testing.T) {
	cases := []struct {
		name string
		mon  hyprMonitor
		want string
	}{
		{
			// The machine this was developed on: 4K panel at 2x.
			name: "hidpi scaled",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 3840, Height: 2400, Scale: 2, Transform: 0},
			want: "0,0 1920x1200",
		},
		{
			name: "unscaled",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 1920, Height: 1080, Scale: 1, Transform: 0},
			want: "0,0 1920x1080",
		},
		{
			// Fractional scaling does not divide evenly; jq floored, so do we.
			name: "fractional scale floors",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 2560, Height: 1440, Scale: 1.5, Transform: 0},
			want: "0,0 1706x960",
		},
		{
			// Odd transforms are the 90 and 270 degree rotations, which swap
			// the axes; an unswapped rect captures the wrong half of a
			// portrait monitor.
			name: "rotated 90 swaps axes",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 3840, Height: 2400, Scale: 2, Transform: 1},
			want: "0,0 1200x1920",
		},
		{
			name: "rotated 180 does not swap",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 1920, Height: 1080, Scale: 1, Transform: 2},
			want: "0,0 1920x1080",
		},
		{
			// Second monitor in a side-by-side layout keeps its origin.
			name: "offset origin preserved",
			mon:  hyprMonitor{X: 1920, Y: 300, Width: 1920, Height: 1080, Scale: 1, Transform: 0},
			want: "1920,300 1920x1080",
		},
		{
			// hyprctl has never reported 0, but dividing by it would panic
			// the whole screenshot path rather than degrade.
			name: "zero scale falls back to raw size",
			mon:  hyprMonitor{X: 0, Y: 0, Width: 1920, Height: 1080, Scale: 0, Transform: 0},
			want: "0,0 1920x1080",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.mon.logicalRect(); got != c.want {
				t.Errorf("logicalRect() = %q, want %q", got, c.want)
			}
		})
	}
}

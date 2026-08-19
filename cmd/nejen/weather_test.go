package main

import "testing"

func TestWeatherCategory(t *testing.T) {
	cases := map[int]string{
		0: "clear", 2: "partly-cloudy", 3: "overcast", 48: "fog",
		55: "rain", 63: "rain", 82: "rain", 71: "snow", 86: "snow",
		95: "thunderstorm", 99: "thunderstorm", 42: "unknown",
	}
	for code, want := range cases {
		if got := weatherCategory(code); got != want {
			t.Errorf("weatherCategory(%d) = %q, want %q", code, got, want)
		}
	}
}

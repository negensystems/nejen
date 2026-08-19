package main

// Waybar weather feed: one JSON line per invocation, temperature and
// condition for the current location. This is the Go port of the vendored
// waybar-weather script (github.com/takumiymd/waybar-weather), which needed
// bash plus a python interpreter for what is one HTTP request and a table
// lookup; it was the last python in the shipped tree.
//
// Data sources (no API key): Open-Meteo for weather, ipapi.co / ip-api.com
// for IP geolocation when no coordinates are configured.
//
// Coordinates and units come from the optional user config
// ~/.config/nejen/weather.toml -- fixed coordinates are more reliable than
// geolocation and are what you want on a VPN, where the detected location is
// the exit node's.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func init() {
	registerCommand("weather", runWeather)
}

const weatherErrorMark = "!!"

type weatherConfig struct {
	Latitude  float64 `toml:"latitude"`
	Longitude float64 `toml:"longitude"`
	Label     string  `toml:"label"`
	Prefix    string  `toml:"prefix"`
	Units     string  `toml:"units"`     // celsius | fahrenheit
	WindUnit  string  `toml:"wind_unit"` // kmh | ms | mph | kn
}

// wmoConditions maps WMO weather interpretation codes to short readable text.
var wmoConditions = map[int]string{
	0: "Clear", 1: "Mainly clear", 2: "Partly cloudy", 3: "Overcast",
	45: "Fog", 48: "Rime fog",
	51: "Light drizzle", 53: "Drizzle", 55: "Heavy drizzle",
	56: "Freezing drizzle", 57: "Freezing drizzle",
	61: "Light rain", 63: "Rain", 65: "Heavy rain",
	66: "Freezing rain", 67: "Freezing rain",
	71: "Light snow", 73: "Snow", 75: "Heavy snow", 77: "Snow grains",
	80: "Light showers", 81: "Showers", 82: "Violent showers",
	85: "Snow showers", 86: "Heavy snow showers",
	95: "Thunderstorm", 96: "Thunderstorm, hail", 99: "Thunderstorm, hail",
}

// weatherCategory maps a WMO code to the CSS class styled in style.css.
func weatherCategory(code int) string {
	switch {
	case code == 0:
		return "clear"
	case code == 1 || code == 2:
		return "partly-cloudy"
	case code == 3:
		return "overcast"
	case code == 45 || code == 48:
		return "fog"
	case (code >= 51 && code <= 57) || (code >= 61 && code <= 67) || (code >= 80 && code <= 82):
		return "rain"
	case (code >= 71 && code <= 77) || code == 85 || code == 86:
		return "snow"
	case code == 95 || code == 96 || code == 99:
		return "thunderstorm"
	}
	return "unknown"
}

func weatherCacheDir() string {
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		home, _ := os.UserHomeDir()
		cache = filepath.Join(home, ".cache")
	}
	return filepath.Join(cache, "nejen", "weather")
}

func loadWeatherConfig() weatherConfig {
	cfg := weatherConfig{Units: "celsius", WindUnit: "kmh"}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	toml.DecodeFile(filepath.Join(configHome, "nejen", "weather.toml"), &cfg)
	if cfg.Units == "" {
		cfg.Units = "celsius"
	}
	if cfg.WindUnit == "" {
		cfg.WindUnit = "kmh"
	}
	return cfg
}

func weatherHTTPJSON(rawURL string, out interface{}) error {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "nejen-weather/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

type weatherLocation struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Label string  `json:"label"`
	TS    int64   `json:"ts"`
}

// resolveWeatherLocation: manual coordinates always win; then a cached
// detection under an hour old; then IP geolocation, HTTPS provider first.
func resolveWeatherLocation(cfg weatherConfig) (weatherLocation, error) {
	if cfg.Latitude != 0 || cfg.Longitude != 0 {
		return weatherLocation{Lat: cfg.Latitude, Lon: cfg.Longitude, Label: cfg.Label}, nil
	}

	locCache := filepath.Join(weatherCacheDir(), "location.json")
	if b, err := os.ReadFile(locCache); err == nil {
		var c weatherLocation
		if json.Unmarshal(b, &c) == nil && time.Since(time.Unix(c.TS, 0)) < time.Hour {
			return c, nil
		}
	}

	var loc weatherLocation
	var g struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		City      string  `json:"city"`
		Region    string  `json:"region"`
		Country   string  `json:"country_name"`
	}
	if err := weatherHTTPJSON("https://ipapi.co/json/", &g); err == nil && (g.Latitude != 0 || g.Longitude != 0) {
		loc = weatherLocation{Lat: g.Latitude, Lon: g.Longitude, Label: joinNonEmpty(g.City, g.Region, g.Country)}
	} else {
		var f struct {
			Status  string  `json:"status"`
			Message string  `json:"message"`
			Lat     float64 `json:"lat"`
			Lon     float64 `json:"lon"`
			City    string  `json:"city"`
			Region  string  `json:"regionName"`
			Country string  `json:"country"`
		}
		if err := weatherHTTPJSON("http://ip-api.com/json/?fields=status,message,lat,lon,city,regionName,country", &f); err != nil {
			return loc, err
		}
		if f.Status != "success" {
			return loc, fmt.Errorf("IP geolocation failed: %s", f.Message)
		}
		loc = weatherLocation{Lat: f.Lat, Lon: f.Lon, Label: joinNonEmpty(f.City, f.Region, f.Country)}
	}

	loc.TS = time.Now().Unix()
	if b, err := json.Marshal(loc); err == nil {
		os.MkdirAll(weatherCacheDir(), 0755)
		os.WriteFile(locCache, b, 0644)
	}
	return loc, nil
}

func joinNonEmpty(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ", ")
}

type weatherPayload struct {
	Text    string   `json:"text"`
	Class   []string `json:"class"`
	Tooltip string   `json:"tooltip"`
}

func emitWeather(p weatherPayload) {
	b, _ := json.Marshal(p)
	fmt.Println(string(b))
}

// weatherFallback prefers the last good reading so a brief network blip does
// not flash an error in the bar.
func weatherFallback(cfg weatherConfig, message string) {
	if b, err := os.ReadFile(filepath.Join(weatherCacheDir(), "weather.json")); err == nil {
		var cached weatherPayload
		if json.Unmarshal(b, &cached) == nil && cached.Text != "" {
			cached.Tooltip += "\n(showing last reading: " + message + ")"
			emitWeather(cached)
			return
		}
	}
	text := weatherErrorMark
	if cfg.Prefix != "" {
		text = cfg.Prefix + " " + weatherErrorMark
	}
	emitWeather(weatherPayload{Text: text, Class: []string{"error"}, Tooltip: "Weather unavailable:\n" + message})
}

func runWeather(args []string) {
	cfg := loadWeatherConfig()

	loc, err := resolveWeatherLocation(cfg)
	if err != nil {
		weatherFallback(cfg, err.Error())
		return
	}

	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%g", loc.Lat))
	params.Set("longitude", fmt.Sprintf("%g", loc.Lon))
	params.Set("current", "temperature_2m,relative_humidity_2m,apparent_temperature,is_day,weather_code,wind_speed_10m")
	params.Set("temperature_unit", cfg.Units)
	params.Set("wind_speed_unit", cfg.WindUnit)
	params.Set("timezone", "auto")

	var data struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
			Humidity    float64 `json:"relative_humidity_2m"`
			Feels       float64 `json:"apparent_temperature"`
			IsDay       int     `json:"is_day"`
			Code        int     `json:"weather_code"`
			Wind        float64 `json:"wind_speed_10m"`
		} `json:"current"`
		Units struct {
			Temperature string `json:"temperature_2m"`
			Wind        string `json:"wind_speed_10m"`
		} `json:"current_units"`
	}
	if err := weatherHTTPJSON("https://api.open-meteo.com/v1/forecast?"+params.Encode(), &data); err != nil {
		weatherFallback(cfg, err.Error())
		return
	}

	cur := data.Current
	cond, ok := wmoConditions[cur.Code]
	if !ok {
		cond = fmt.Sprintf("Code %d", cur.Code)
	}

	tUnit := data.Units.Temperature
	if tUnit == "" {
		tUnit = "°"
	}
	tempStr := fmt.Sprintf("%.0f%s", cur.Temperature, tUnit)

	text := tempStr + " " + cond
	if cfg.Prefix != "" {
		text = cfg.Prefix + " " + text
	}

	locLine := loc.Label
	if locLine == "" {
		locLine = fmt.Sprintf("%.3f, %.3f", loc.Lat, loc.Lon)
	}
	tooltip := strings.Join([]string{
		locLine,
		"Weather: " + cond,
		fmt.Sprintf("Temperature: %s (feels %.0f%s)", tempStr, cur.Feels, tUnit),
		fmt.Sprintf("Humidity: %.0f%%", cur.Humidity),
		strings.TrimSpace(fmt.Sprintf("Wind: %.0f %s", cur.Wind, data.Units.Wind)),
		"Updated: " + time.Now().Format("15:04"),
	}, "\n")

	dayClass := "day"
	if cur.IsDay != 1 {
		dayClass = "night"
	}
	payload := weatherPayload{Text: text, Class: []string{weatherCategory(cur.Code), dayClass}, Tooltip: tooltip}

	// Cache this reading so a later failure can reuse it.
	if b, err := json.Marshal(payload); err == nil {
		os.MkdirAll(weatherCacheDir(), 0755)
		os.WriteFile(filepath.Join(weatherCacheDir(), "weather.json"), b, 0644)
	}

	emitWeather(payload)
}

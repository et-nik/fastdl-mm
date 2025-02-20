package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                string            `yaml:"host"`
	Port                uint16            `yaml:"port"`
	PortRange           ConfigPortRange   `yaml:"portRange"`
	AutoIndexEnabled    bool              `yaml:"autoIndexEnabled"`
	ServePrecached      bool              `yaml:"servePrecached"`
	ForbiddenRegexp     []string          `yaml:"forbiddenRegexp"`
	ForbiddenExtensions []string          `yaml:"forbiddenExtensions"`
	AllowedExtensions   []string          `yaml:"allowedExtensions"`
	ForbiddenPaths      []string          `yaml:"forbiddenPaths"`
	AllowedPaths        []string          `yaml:"allowedPaths"`
	CacheSize           ConfigCacheSize   `yaml:"cacheSize"`
	CustomDownloadURL   string            `yaml:"customDownloadURL"`
	HTTP                ConfigHTTP        `yaml:"http"`
	RateLimits          []ConfigRateLimit `yaml:"rateLimits"`
	BlockListIP         []string          `yaml:"blockListIP"`
}

type ConfigHTTP struct {
	ReadTimeout  ConfigTimeout `yaml:"readTimeout"`
	WriteTimeout ConfigTimeout `yaml:"writeTimeout"`
}

type ConfigRateLimit struct {
	period string `yaml:"period"`
	limit  int    `yaml:"limit"`
}

func (c ConfigRateLimit) Period() time.Duration {
	d, err := time.ParseDuration(c.period)
	if err != nil {
		return 0
	}

	return d
}

func (c ConfigRateLimit) Limit() int {
	return c.limit
}

type ConfigTimeout string

func (c ConfigTimeout) Duration() time.Duration {
	d, err := time.ParseDuration(string(c))
	if err != nil {
		return 0
	}

	return d
}

type ConfigCacheSize string

func (c ConfigCacheSize) Int64() int64 {
	multipliers := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	str := strings.TrimSpace(strings.ToUpper(string(c)))

	var numberPart string
	var unitPart string
	for i, r := range str {
		if r < '0' || r > '9' {
			numberPart = str[:i]
			unitPart = str[i:]
			break
		}
	}

	if numberPart == "" {
		return 0
	}

	number, err := strconv.ParseInt(numberPart, 10, 64)
	if err != nil {
		return 0
	}

	multiplier, exists := multipliers[unitPart]
	if !exists {
		return number
	}

	return number * multiplier
}

type ConfigPortRange string

func (c ConfigPortRange) IntRange() (int, int) {
	splitted := strings.SplitN(string(c), "-", 2)
	if len(splitted) != 2 {
		return 0, 0
	}

	low, err := strconv.Atoi(splitted[0])
	if err != nil {
		return 0, 0
	}

	high, err := strconv.Atoi(splitted[1])
	if err != nil {
		return 0, 0
	}

	return low, high
}

func ParseConfig(in []byte) (*Config, error) {
	var cfg Config

	err := yaml.Unmarshal(in, &cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal config")
	}

	return &cfg, nil
}

var DefaultConfig = &Config{
	Host:             "",
	Port:             0,
	AutoIndexEnabled: false,
	ForbiddenRegexp: []string{
		"mapcycle.*",
		".*textscheme.*",
	},
	AllowedExtensions: []string{
		"bmp",
		"bsp",
		"gif",
		"jpeg",
		"jpg",
		"lmp",
		"lst",
		"mdl",
		"mp3",
		"png",
		"res",
		"spr",
		"tga",
		"txt",
		"wad",
		"wav",
		"zip",
	},
	AllowedPaths: []string{
		"gfx",
		"maps",
		"media",
		"models",
		"overviews",
		"sound",
		"sprites",
	},
	CacheSize: "100MB",
}

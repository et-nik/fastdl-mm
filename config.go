package main

import (
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

type Config struct {
	Host                string          `yaml:"host"`
	Port                uint16          `yaml:"port"`
	AutoIndexEnabled    bool            `yaml:"autoIndexEnabled"`
	ForbiddenRegexp     []string        `yaml:"forbiddenRegexp"`
	ForbiddenExtentions []string        `yaml:"forbiddenExtentions"`
	AllowedExtentions   []string        `yaml:"allowedExtentions"`
	ForbiddenPaths      []string        `yaml:"forbiddenPaths"`
	AllowedPaths        []string        `yaml:"allowedPaths"`
	CacheSize           ConfigCacheSize `yaml:"cacheSize"`
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

func ParseConfig(in []byte) (*Config, error) {
	var cfg Config

	err := yaml.Unmarshal(in, &cfg)
	if err != nil {
		return nil, err
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
	AllowedExtentions: []string{
		"lmp",
		"lst",
		"wad",
		"bmp",
		"tga",
		"jpg",
		"jpeg",
		"png",
		"gif",
		"txt",
		"zip",
		"bsp",
		"res",
		"wav",
		"mp3",
		"spr",
	},
	AllowedPaths: []string{
		"gfx",
		"maps",
		"media",
		"models",
		"sound",
		"sprites",
	},
	CacheSize: "100MB",
}

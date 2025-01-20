package main

import (
	"gopkg.in/yaml.v3"
)

type Config struct {
	Host                string   `yaml:"host"`
	Port                uint16   `yaml:"port"`
	AutoIndexEnabled    bool     `yaml:"autoIndexEnabled"`
	ForbiddenRegexp     []string `yaml:"forbiddenRegexp"`
	ForbiddenExtentions []string `yaml:"forbiddenExtentions"`
	AllowedExtentions   []string `yaml:"allowedExtentions"`
	ForbiddenPaths      []string `yaml:"forbiddenPaths"`
	AllowedPaths        []string `yaml:"allowedPaths"`
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
}

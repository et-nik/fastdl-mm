package main

import (
	"gopkg.in/yaml.v3"
)

type Config struct {
	FastDLHost          string   `yaml:"fastDLHost"`
	FastDLPort          uint16   `yaml:"fastDLPort"`
	AutoIndexEnabled    bool     `yaml:"autoIndexEnabled"`
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
	FastDLHost:       "",
	FastDLPort:       0,
	AutoIndexEnabled: true,
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

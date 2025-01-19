package main

import (
	"fmt"
	metamod "github.com/et-nik/metamod-go"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	err := metamod.SetPluginInfo(&metamod.PluginInfo{
		InterfaceVersion: metamod.MetaInterfaceVersion,
		Name:             "FastDL",
		Version:          strings.TrimPrefix(Version, "v"),
		Date:             BuildDate,
		Author:           "KNiK",
		Url:              "https://github.com/et-nik/fastdl-mm",
		LogTag:           "FastDL",
		Loadable:         metamod.PluginLoadTimeStartup,
		Unloadable:       metamod.PluginLoadTimeAnyTime,
	})
	if err != nil {
		panic(err)
	}

	err = metamod.SetMetaCallbacks(&metamod.MetaCallbacks{
		MetaQuery: MetaQuery,
	})
	if err != nil {
		panic(err)
	}
}

func MetaQuery() int {
	engineFuncs, err := metamod.GetEngineFuncs()
	if err != nil {
		log.Fatalf("Failed to get engine funcs: %s", err.Error())
	}

	gameDir := engineFuncs.GetGameDir()

	cfg := loadConfig(gameDir)

	if cfg.FastDLHost == "" {
		ip := engineFuncs.CVarGetString("ip")
		cfg.FastDLHost = ip
	}

	if cfg.FastDLHost == "" || cfg.FastDLHost == "0.0.0.0" {
		panic("FastDLHost is not set")
	}

	if cfg.FastDLPort == 0 {
		setRandomPort(cfg)
	}

	go runServer(gameDir, cfg)

	engineFuncs.ServerCommand(
		fmt.Sprintf(
			"sv_downloadurl \"%s\"",
			fmt.Sprintf("http://%s:%d", cfg.FastDLHost, cfg.FastDLPort),
		),
	)
	engineFuncs.ServerExecute()

	return 1
}

func loadConfig(gameDir string) *Config {
	fileNames := []string{
		"fastdl.yml",
		"fastdl.yaml",
	}

	dirs := []string{
		gameDir,
		filepath.Join(gameDir, "addons", "fastdl"),
	}

	var files []string
	for _, dir := range dirs {
		for _, fileName := range fileNames {
			files = append(files, filepath.Join(dir, fileName))
		}
	}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			configContents, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read config file: %s", err.Error())
			}

			cfg, err := ParseConfig(configContents)
			if err != nil {
				log.Fatalf("Failed to parse config file: %s", err.Error())
			}

			return cfg
		}
	}

	return loadDefaultConfig()
}

func loadDefaultConfig() *Config {
	return DefaultConfig
}

func setRandomPort(cfg *Config) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to get random port: %s", err.Error())
	}

	cfg.FastDLPort = uint16(listener.Addr().(*net.TCPAddr).Port)

	err = listener.Close()
	if err != nil {
		log.Fatalf("Failed to close listener: %s", err.Error())
	}
}

func main() {}

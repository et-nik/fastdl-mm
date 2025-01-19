package main

import (
	"fmt"
	metamod "github.com/et-nik/metamod-go"
	"log"
	"net"
	"os"
	"path/filepath"
)

func init() {
	err := metamod.SetPluginInfo(&metamod.PluginInfo{
		InterfaceVersion: metamod.MetaInterfaceVersion,
		Name:             "FastDL",
		Version:          Version,
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
		filepath.Join(gameDir, "addons/fastdl"),
	}

	var files []string
	for _, dir := range dirs {
		for _, fileName := range fileNames {
			files = append(files, filepath.Join(dir, fileName))
		}
	}

	for _, file := range files {
		path := filepath.Join(gameDir, file)
		if _, err := os.Stat(path); err == nil {
			configContents, err := os.ReadFile(path)
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
	defConf := DefaultConfig

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to get random port: %s", err.Error())
	}

	defConf.FastDLPort = uint16(listener.Addr().(*net.TCPAddr).Port)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Failed to close listener: %s", err.Error())
		}
	}(listener)

	return defConf
}

func main() {}

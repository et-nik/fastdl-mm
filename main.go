package main

import (
	"fmt"
	metamod "github.com/et-nik/metamod-go"
	"github.com/pkg/errors"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
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

	plugin := NewPlugin()

	err = metamod.SetMetaCallbacks(&metamod.MetaCallbacks{
		MetaQuery:  metaQueryFn(plugin),
		MetaDetach: metaDetachFn(plugin),
	})
	if err != nil {
		panic(err)
	}

	err = metamod.SetApiCallbacks(&metamod.APICallbacks{
		GameDLLInit: func() metamod.APICallbackResult {
			err := plugin.Init()
			if err != nil {
				slog.Error("Failed to init plugin: ", "error", err)

				return metamod.APICallbackResultHandled
			}

			slog.Debug("Plugin initialized")

			return metamod.APICallbackResultHandled
		},
		ServerDeactivate: func() metamod.APICallbackResult {
			err := plugin.Reset()
			if err != nil {
				slog.Error("Failed to reset plugin: ", "error", err)

				return metamod.APICallbackResultHandled
			}

			slog.Debug("Server deactivated")

			return metamod.APICallbackResultHandled
		},
		ServerActivate: func(_ *metamod.Edict, _ int, _ int) metamod.APICallbackResult {
			slog.Debug("Server activated")

			if plugin.cfg.ServePrecached {
				processMapRelatedResource(plugin)
			}

			return metamod.APICallbackResultHandled
		},
	})

	if err != nil {
		panic(err)
	}

	err = metamod.SetEngineHooks(&metamod.EngineHooks{
		PrecacheGeneric: func(filePath string) (metamod.EngineHookResult, int) {
			slog.Debug("Precaching generic", "filePath", filePath)

			plugin.AppendPrecached(filePath)

			return metamod.EngineHookResultHandled, 0
		},
		PrecacheModel: func(modelPath string) (metamod.EngineHookResult, int) {
			slog.Debug("Precaching model", "filePath", modelPath)

			plugin.AppendPrecached(modelPath)

			return metamod.EngineHookResultHandled, 0
		},
		PrecacheSound: func(soundPath string) (metamod.EngineHookResult, int) {
			fullPath := filepath.Join("sound", soundPath)

			slog.Debug("Precaching sound", "filePath", fullPath)

			plugin.AppendPrecached(fullPath)

			return metamod.EngineHookResultHandled, 0
		},
	})
}

func metaQueryFn(p *Plugin) func() int {
	return func() int {
		engineFuncs, err := metamod.GetEngineFuncs()
		if err != nil {
			slog.Error("Failed to get engine funcs: ", "error", err)

			return 0
		}

		metaUtilFn, err := metamod.GetMetaUtilFuncs()
		if err != nil {
			slog.Error("Failed to get meta util funcs: ", "error", err)

			return 0
		}

		logLevel := slog.LevelInfo

		developerCvarValue := engineFuncs.CVarGetFloat("developer")
		if developerCvarValue > 0 {
			logLevel = slog.LevelDebug
		}

		slog.SetDefault(
			slog.New(
				slog.NewTextHandler(
					NewMetaLogWriter(metaUtilFn),
					&slog.HandlerOptions{
						Level: logLevel,
						ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
							if a.Key == slog.TimeKey {
								return slog.Attr{}
							}

							return a
						},
					},
				),
			),
		)

		gameDir := engineFuncs.GetGameDir()

		cfg, err := loadConfig(gameDir)
		if err != nil {
			slog.Error(
				"Failed to load config",
				"error", err,
			)

			return 0
		}

		if cfg.Host == "" {
			ip := engineFuncs.CVarGetString("ip")
			cfg.Host = ip
		}

		if (cfg.Host == "" || cfg.Host == "0.0.0.0") && cfg.CustomDownloadURL == "" {
			panic("host is not set, please set host or customDownloadURL in config")
		}

		if cfg.Port == 0 {
			err = setRandomPort(cfg)
			if err != nil {
				panic(err)
			}
		}

		p.SetConfig(cfg)
		p.SetGameDir(gameDir)

		go func() {
			runtime.LockOSThread()

			err := p.RunServer(gameDir)
			if err != nil {
				panic(err)
			}
		}()

		var svDownloadUrl string

		if cfg.CustomDownloadURL != "" {
			svDownloadUrl = cfg.CustomDownloadURL
		} else {
			svDownloadUrl = fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
		}

		slog.Debug(
			"Change sv_downloadurl",
			"sv_downloadurl", svDownloadUrl,
		)

		engineFuncs.ServerCommand(
			fmt.Sprintf(
				"sv_downloadurl \"%s\"",
				svDownloadUrl,
			),
		)
		engineFuncs.ServerExecute()

		return 1
	}
}

func metaDetachFn(p *Plugin) func(now int, reason int) int {
	return func(now int, reason int) int {
		err := p.Shutdown()
		if err != nil {
			slog.Error(
				"Failed to shutdown plugin: ",
				"error", err,
			)
		}

		return 1
	}
}

func loadConfig(gameDir string) (*Config, error) {
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
				return nil, errors.WithMessage(err, "failed to read config file")
			}

			cfg, err := ParseConfig(configContents)
			if err != nil {
				return nil, errors.WithMessage(err, "failed to parse config file")
			}

			return cfg, nil
		}
	}

	return DefaultConfig, nil
}

func setRandomPort(cfg *Config) error {
	randomPort := 0
	minPort, maxPort := cfg.PortRange.IntRange()

	var listener net.Listener
	var err error

	for {
		if maxPort > 0 && minPort < maxPort {
			randomPort = rand.Intn(maxPort-minPort+1) + minPort
		}

		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", randomPort))
		if err != nil {
			continue
		}

		break
	}

	cfg.Port = uint16(listener.Addr().(*net.TCPAddr).Port)

	err = listener.Close()
	if err != nil {
		return errors.WithMessage(err, "failed to close listener")
	}

	return nil
}

func processMapRelatedResource(p *Plugin) {
	engineFuncs, err := metamod.GetEngineFuncs()
	if err != nil {
		slog.Error(
			"Failed to get engine funcs: ",
			"error", err,
		)

		return
	}

	wordspawn := engineFuncs.EntityOfEntIndex(0)
	if wordspawn == nil {
		slog.Error("Failed to get worldspawn entity")

		return
	}

	// Append map specified resources
	mapPath := wordspawn.EntVars().Model()
	p.AppendPrecached(mapPath)

	mapName := strings.TrimSuffix(filepath.Base(mapPath), filepath.Ext(mapPath))

	p.AppendPrecached(filepath.Join("overviews", fmt.Sprintf("%s.txt", mapName)))
	p.AppendPrecached(filepath.Join("overviews", fmt.Sprintf("%s.bmp", mapName)))
	p.AppendPrecached(filepath.Join("overviews", fmt.Sprintf("%s.tga", mapName)))

	p.AppendPrecached(filepath.Join("maps", fmt.Sprintf("%s.txt", mapName)))
	p.AppendPrecached(filepath.Join("maps", fmt.Sprintf("%s_detail.txt", mapName)))

	// Append sky
	skyName := engineFuncs.CVarGetString("sv_skyname")

	if skyName != "" {
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sbk.tga", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sdn.tga", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sft.tga", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%slf.tga", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%srt.tga", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sup.tga", skyName)))

		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sbk.bmp", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sdn.bmp", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sft.bmp", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%slf.bmp", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%srt.bmp", skyName)))
		p.AppendPrecached(filepath.Join("gfx", "env", fmt.Sprintf("%sup.bmp", skyName)))
	}
}

func main() {}

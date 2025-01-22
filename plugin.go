package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"path/filepath"
)

type Plugin struct {
	cfg     *Config
	gameDir string

	server *http.Server

	precachedFiles *map[string]struct{}
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) SetConfig(cfg *Config) {
	p.cfg = cfg
}

func (p *Plugin) SetGameDir(gameDir string) {
	p.gameDir = gameDir
}

func (p *Plugin) GameDir() string {
	return p.gameDir
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) AppendPrecached(filePath string) {
	if !p.cfg.ServePrecached {
		return
	}

	if p.precachedFiles == nil {
		precachedFiles := make(map[string]struct{}, 250)
		p.precachedFiles = &precachedFiles
	}

	(*p.precachedFiles)[filePath] = struct{}{}

	dir := filepath.Dir(filePath)

	for {
		if dir == "." || dir == "/" {
			break
		}

		(*p.precachedFiles)[dir] = struct{}{}
		dir = filepath.Dir(dir)
	}
}

func (p *Plugin) Reset() error {
	if p.cfg.ServePrecached {
		precachedFiles := make(map[string]struct{}, 250)
		p.precachedFiles = &precachedFiles
	}

	return nil
}

func (p *Plugin) Shutdown() error {
	if p.server == nil {
		return nil
	}

	err := p.server.Shutdown(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}

	return nil
}

func (p *Plugin) RunServer(gameDir string) error {
	h := newFileHandler(gameDir, p)

	http.HandleFunc("/", h.ServeHTTP)

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)

	p.server = &http.Server{
		Addr: addr,
	}

	slog.Info(fmt.Sprintf("FastDL HTTP Starting server on %s...\n", addr))

	err := p.server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}

	slog.Info("FastDL HTTP Server stopped")

	return nil
}

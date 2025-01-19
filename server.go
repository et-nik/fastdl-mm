package main

import (
	"fmt"
	"github.com/et-nik/fastdl-mm/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type fileHandler struct {
	baseDir string
	config  *Config

	allowedExtensions   map[string]struct{}
	forbiddenExtensions map[string]struct{}
	allowedPaths        map[string]struct{}
	forbiddenPaths      map[string]struct{}
}

func newFileHandler(baseDir string, cfg *Config) *fileHandler {
	allowedExtensions := make(map[string]struct{}, len(cfg.AllowedExtentions))
	for _, ext := range cfg.AllowedExtentions {
		allowedExtensions[ext] = struct{}{}
	}

	forbiddenExtensions := make(map[string]struct{}, len(cfg.ForbiddenExtentions))
	for _, ext := range cfg.ForbiddenExtentions {
		forbiddenExtensions[ext] = struct{}{}
	}

	allowedPaths := make(map[string]struct{}, len(cfg.AllowedPaths))
	for _, path := range cfg.AllowedPaths {
		allowedPaths[path] = struct{}{}
	}

	forbiddenPaths := make(map[string]struct{}, len(cfg.ForbiddenPaths))
	for _, path := range cfg.ForbiddenPaths {
		forbiddenPaths[path] = struct{}{}
	}

	return &fileHandler{
		baseDir: baseDir,
		config:  cfg,
	}
}

//var allowedExtensions = map[string]bool{
//	".txt":  true,
//	".jpg":  true,
//	".png":  true,
//	".json": true,
//}

//var baseDir string

//func isAllowedExtension(filePath string) bool {
//	ext := strings.ToLower(filepath.Ext(filePath))
//	return allowedExtensions[ext]
//}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.baseDir == "" {
		http.Error(w, "Base directory is not set", http.StatusInternalServerError)
	}

	requestedPath := filepath.Join(h.baseDir, filepath.Clean(r.URL.Path))

	info, err := os.Stat(requestedPath)
	if err != nil {
		http.NotFound(w, r)

		return
	}

	if info.IsDir() {
		if !h.config.AutoIndexEnabled {
			http.NotFound(w, r)

			return
		}

		if !h.pathAllowed(requestedPath) {
			http.NotFound(w, r)

			return
		}

		entries, err := os.ReadDir(requestedPath)
		if err != nil {
			http.Error(w, "Failed to read directory", http.StatusInternalServerError)
			return
		}

		var items []string
		for _, entry := range entries {
			entryPath := filepath.Join(r.URL.Path, entry.Name())
			fullPath := filepath.Join(requestedPath, entry.Name())

			if entry.IsDir() || h.fileAllowed(fullPath) {
				items = append(items, entryPath+"/")
			}
		}

		w.Header().Set("Content-Type", "text/html")
		err = template.AutoIndexTemplate.Execute(w, struct {
			Path  string
			Items []string
		}{
			Path:  r.URL.Path,
			Items: items,
		})
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}

		return
	}

	if !h.fileAllowed(requestedPath) {
		http.NotFound(w, r)

		return
	}

	http.ServeFile(w, r, requestedPath)
}

func (h *fileHandler) fileAllowed(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	if _, ok := h.forbiddenExtensions[ext]; ok {
		return false
	}

	if _, ok := h.allowedExtensions[ext]; ok {
		return true
	}

	// If no extensions are allowed, then all extensions are allowed
	return len(h.allowedExtensions) == 0
}

func (h *fileHandler) pathAllowed(filePath string) bool {
	for path := range h.forbiddenPaths {
		if strings.HasPrefix(filePath, path) {
			return false
		}
	}

	for path := range h.allowedPaths {
		if strings.HasPrefix(filePath, path) {
			return true
		}
	}

	// If no paths are allowed, then all paths are allowed
	return len(h.allowedPaths) == 0
}

func runServer(gameDir string, cfg *Config) {
	h := newFileHandler(gameDir, cfg)

	http.HandleFunc("/", h.ServeHTTP)

	addr := fmt.Sprintf("%s:%d", cfg.FastDLHost, cfg.FastDLPort)

	log.Printf("Starting server on %s...\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

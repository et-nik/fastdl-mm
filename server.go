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

var allowedExtensions = map[string]bool{
	".txt":  true,
	".jpg":  true,
	".png":  true,
	".json": true,
}

var baseDir string

func isAllowedExtension(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return allowedExtensions[ext]
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	if baseDir == "" {
		http.Error(w, "Base directory is not set", http.StatusInternalServerError)
	}

	requestedPath := filepath.Join(baseDir, filepath.Clean(r.URL.Path))

	info, err := os.Stat(requestedPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		entries, err := os.ReadDir(requestedPath)
		if err != nil {
			http.Error(w, "Failed to read directory", http.StatusInternalServerError)
			return
		}

		var items []string
		for _, entry := range entries {
			entryPath := filepath.Join(r.URL.Path, entry.Name())
			fullPath := filepath.Join(requestedPath, entry.Name())

			if entry.IsDir() || isAllowedExtension(fullPath) {
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

	if !isAllowedExtension(requestedPath) {
		http.Error(w, "Forbidden file extension", http.StatusForbidden)

		return
	}

	http.ServeFile(w, r, requestedPath)
}

func runServer(gameDir string, cfg *Config) {
	baseDir = gameDir

	http.HandleFunc("/", fileHandler)

	addr := fmt.Sprintf("%s:%d", cfg.FastDLHost, cfg.FastDLPort)

	log.Printf("Starting server on %s...\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

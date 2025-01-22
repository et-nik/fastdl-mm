package main

import (
	"github.com/et-nik/fastdl-mm/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type fileHandler struct {
	baseDir string
	plugin  *Plugin
	config  *Config

	fileCache *MRUCache

	allowedExtensions   map[string]struct{}
	forbiddenExtensions map[string]struct{}
	allowedPaths        map[string]struct{}
	forbiddenPaths      map[string]struct{}
	forbiddenRegexps    []*regexp.Regexp
}

func newFileHandler(baseDir string, plugin *Plugin) *fileHandler {
	allowedExtensions := make(map[string]struct{}, len(plugin.cfg.AllowedExtentions))
	for _, ext := range plugin.cfg.AllowedExtentions {
		allowedExtensions[ext] = struct{}{}
	}

	forbiddenExtensions := make(map[string]struct{}, len(plugin.cfg.ForbiddenExtentions))
	for _, ext := range plugin.cfg.ForbiddenExtentions {
		forbiddenExtensions[ext] = struct{}{}
	}

	allowedPaths := make(map[string]struct{}, len(plugin.cfg.AllowedPaths))
	for _, path := range plugin.cfg.AllowedPaths {
		allowedPaths[path] = struct{}{}
	}

	forbiddenPaths := make(map[string]struct{}, len(plugin.cfg.ForbiddenPaths))
	for _, path := range plugin.cfg.ForbiddenPaths {
		forbiddenPaths[path] = struct{}{}
	}

	forbiddenRegexps := make([]*regexp.Regexp, 0, len(plugin.cfg.ForbiddenRegexp))
	for _, regxp := range plugin.cfg.ForbiddenRegexp {
		r, err := regexp.Compile(regxp)
		if err != nil {
			slog.Error("Failed to compile firbidden regexp",
				"regexp", regxp,
				"error", err,
			)
		}

		forbiddenRegexps = append(forbiddenRegexps, r)
	}

	return &fileHandler{
		baseDir: baseDir,
		plugin:  plugin,
		config:  plugin.cfg,

		fileCache: NewMRUCache(plugin.cfg.CacheSize.Int64()),

		allowedExtensions:   allowedExtensions,
		forbiddenExtensions: forbiddenExtensions,
		allowedPaths:        allowedPaths,
		forbiddenPaths:      forbiddenPaths,
		forbiddenRegexps:    forbiddenRegexps,
	}
}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.baseDir == "" {
		http.Error(w, "Base directory is not set", http.StatusInternalServerError)

		return
	}

	requestedPath := filepath.Clean(r.URL.Path)
	fullPath := filepath.Join(h.baseDir, requestedPath)

	if h.fileCache.Exists(fullPath) {
		// Serve cached file.

		http.ServeFileFS(w, r, h.fileCache, fullPath)

		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.NotFound(w, r)

		return
	}

	if info.IsDir() {
		h.serveDirInfo(w, r, requestedPath, fullPath)

		return
	}

	if !h.fileAllowed(requestedPath) {
		http.NotFound(w, r)

		return
	}

	if !h.fileCache.Exists(fullPath) {
		contents, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)

			return
		}

		h.fileCache.Put(fullPath, &CacheFile{
			Contents: contents,
			FileInfo: info,
		})
	}

	http.ServeFile(w, r, fullPath)
}

func (h *fileHandler) serveDirInfo(w http.ResponseWriter, r *http.Request, requestedPath, fullPath string) {
	if !h.config.AutoIndexEnabled {
		http.NotFound(w, r)

		return
	}

	if !h.pathAllowed(requestedPath) {
		http.NotFound(w, r)

		return
	}

	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)

		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)

		return
	}

	items := make([]string, 0, len(entries)+1)

	if requestedPath != "/" {
		items = append(items, "../")
	}

	for _, entry := range entries {
		entryPath := filepath.Join(requestedPath, entry.Name())

		if entry.IsDir() && h.pathAllowed(entryPath) {
			items = append(items, entry.Name()+"/")
		} else if h.fileAllowed(entryPath) {
			items = append(items, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "text/html")
	err = template.AutoIndexTemplate.Execute(w, struct {
		Path  string
		Items []string
	}{
		Path:  requestedPath,
		Items: items,
	})
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}

	return
}

func (h *fileHandler) fileAllowed(filePath string) bool {
	filePath = strings.TrimPrefix(filePath, "/")

	fileName := filepath.Base(filePath)
	if strings.HasPrefix(fileName, ".") {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	ext = strings.TrimPrefix(ext, ".")

	if h.config.ServePrecached && ext != "wad" {
		if h.plugin.precachedFiles == nil {
			return false
		}

		if _, ok := (*h.plugin.precachedFiles)[filePath]; !ok {
			return false
		}
	}

	if ext == "cfg" || ext == "ini" {
		return false
	}

	for _, r := range h.forbiddenRegexps {
		if r.MatchString(fileName) {
			return false
		}
	}

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
	if filePath == "/" {
		return true
	}

	filePath = strings.TrimPrefix(filePath, "/")

	if strings.HasSuffix(filePath, "addons") {
		return false
	}

	if h.config.ServePrecached {
		if h.plugin.precachedFiles == nil {
			return false
		}

		if _, ok := (*h.plugin.precachedFiles)[filePath]; !ok {
			return false
		}
	}

	for _, r := range h.forbiddenRegexps {
		if r.MatchString(filePath) {
			return false
		}
	}

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

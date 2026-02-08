package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// Remove /static/ prefix and get the actual file path (legacy support)
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	// Map legacy /static/ to the new ui directory
	fullPath := filepath.Join("ui", filePath)

	file, err := os.Stat(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if file.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Set correct MIME type based on file extension
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".js", ".mjs":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	}

	http.ServeFile(w, r, fullPath)
}

package handler

import (
	"fmt"
	"net/http"
	"strings"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
    
    if strings.HasPrefix(r.URL.Path, "/api/") {
        fmt.Println("hhhh")
        w.WriteHeader(http.StatusNotFound)
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"error":"API endpoint not found"}`))
        return
    }
    // For all other paths, serve SPA index.html
    http.ServeFile(w, r, "./static/index.html")
}
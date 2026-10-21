// Package web serves the built-in static UI.
package web

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/derfenix/webarchive/internal/config"
	"github.com/derfenix/webarchive/ui"
)

func NewUI(cfg config.UI) *UI {
	return &UI{
		prefix: cfg.Prefix,
		theme:  cfg.Theme,
	}
}

type UI struct {
	prefix string
	theme  string
}

func (u *UI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	root, err := fs.Sub(ui.StaticFiles, u.theme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(r.URL.Path, u.prefix) {
		r.URL.Path = "/" + strings.TrimPrefix(r.URL.Path, u.prefix)
	}

	// Everything but static files is handled by the single-page app.
	if !strings.HasPrefix(r.URL.Path, "/static") {
		r.URL.Path = "/"
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")

	http.FileServer(http.FS(root)).ServeHTTP(w, r)
}

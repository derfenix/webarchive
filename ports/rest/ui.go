package rest

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/derfenix/webarchive/config"
	"github.com/derfenix/webarchive/ui"
)

func NewUI(cfg config.UI) *UI {
	return &UI{prefix: cfg.Prefix}
}

type UI struct {
	prefix string
}

func (u *UI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveRoot, err := fs.Sub(ui.StaticFiles, "static")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(r.URL.Path, u.prefix) {
		r.URL.Path = "/" + strings.TrimPrefix(r.URL.Path, u.prefix)
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")

	http.FileServer(http.FS(serveRoot)).ServeHTTP(w, r)
}

func (u *UI) IsUIRequest(r *http.Request) bool {
	return r.URL.Path == u.prefix || strings.HasPrefix(r.URL.Path, "/static/")
}

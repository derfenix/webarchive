package rest

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/derfenix/webarchive/config"
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
	serveRoot, err := fs.Sub(ui.StaticFiles, u.theme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(r.URL.Path, u.prefix) {
		r.URL.Path = "/" + strings.TrimPrefix(r.URL.Path, u.prefix)
	}
	if !strings.HasPrefix(r.URL.Path, "/static") {
		r.URL.Path = "/"
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")

	http.FileServer(http.FS(serveRoot)).ServeHTTP(w, r)
}

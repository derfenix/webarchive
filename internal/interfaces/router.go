// Package interfaces contains the transports of the service.
package interfaces

import (
	"net/http"
	"strings"
)

// APIPrefix is where the REST API is mounted, for both the generated router and
// the handler that splits API requests from UI ones.
const APIPrefix = "/api/v1"

// NewRouter sends API requests to api and everything else to ui.
func NewRouter(api, ui http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)

			return
		}

		ui.ServeHTTP(w, r)
	})
}

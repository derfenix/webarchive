package ui

import (
	"embed"
)

//go:embed */*.html */*.css */*.js
var StaticFiles embed.FS

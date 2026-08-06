package embedfs

import (
	"embed"
	"io/fs"
)

// Dist contains the compiled Vite frontend under frontend/dist.
//
//go:embed all:dist
var dist embed.FS

func Dist() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}

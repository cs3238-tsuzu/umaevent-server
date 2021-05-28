// +build !embed

package main

import (
	"io/fs"
	"os"
)

var staticFS = func() fs.FS {
	dir := os.Getenv("FRONTEND_ROOT")
	if dir == "" {
		dir = "frontend/dist"
	}

	return os.DirFS(dir)
}()

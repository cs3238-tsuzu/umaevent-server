// +build embed

package main

import (
	"embed"
	"io/fs"
)

//go:generate bash -c "cd frontend && npm run build"

//go:embed frontend/dist/*
var embedFS embed.FS

var staticFS = func() fs.FS {
	fs, err := fs.Sub(embedFS, "frontend/dist")

	if err != nil {
		panic(err)
	}

	return fs
}()

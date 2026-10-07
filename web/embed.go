// Package web embeds the static browser files into the Go binary,
// so the server ships as a single file without a separate assets folder.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var files embed.FS

// StaticFS returns the contents of the static/ folder as an http.FileSystem.
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(files, "static")
	if err != nil {
		panic(err) // cannot happen: "static" is embedded at compile time
	}
	return http.FS(sub)
}

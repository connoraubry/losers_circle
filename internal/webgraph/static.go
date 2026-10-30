package webgraph

import (
	"embed"
	"io/fs"
)

//go:embed web
var embeddedWeb embed.FS

// staticFS serves web/'s contents at the root path, e.g. web/index.html
// as "/".
var staticFS = func() fs.FS {
	sub, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		panic(err)
	}
	return sub
}()

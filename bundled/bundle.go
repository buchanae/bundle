package main

// This file is generated. DO NOT EDIT.

import "os"
import "net/http"

type bundleFS map[string]string

var Bundle = bundleFS{
	"index.html": "/Users/abuchanan/projects/bundle/example/index.html", "style": "/Users/abuchanan/projects/bundle/example/style", "style/style.css": "/Users/abuchanan/projects/bundle/example/style/style.css",
}

func (b bundleFS) Open(name string) (http.File, error) {
	f, ok := b[name]
	if !ok {
		// TODO look into https://golang.org/src/net/http/fs.go#L42
		return nil, os.ErrNotExist
	}
	return os.Open(f)
}

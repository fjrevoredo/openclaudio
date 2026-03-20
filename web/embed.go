package webassets

import "embed"

// FS contains the embedded templates and built static assets.
//
//go:embed templates/*.html static/dist/*
var FS embed.FS

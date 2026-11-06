package web

import "embed"

// FS holds the web page, so it is compiled into the program.
//
//go:embed index.html
var FS embed.FS

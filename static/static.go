package static

import "embed"

//go:embed *
var StaticFile embed.FS

package web

import "embed"

// Files embute toda a pasta web na compilação do Go
//go:embed *
var Files embed.FS

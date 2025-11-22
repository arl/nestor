// Package assets provides embedded application assets.
package assets

import "embed"

//go:embed *
var FS embed.FS

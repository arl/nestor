// Package resource provides embedded resources for the application.
package resource

import _ "embed"

//go:embed DejaVuSans.ttf
var DejaVuSansFont []byte

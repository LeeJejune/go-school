// Package migrations embeds reviewed SQL in the migration binary.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS

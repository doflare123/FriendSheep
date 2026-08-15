package migrations

import "embed"

// Files содержит неизменяемую историю миграций notify_service.
//
//go:embed *.sql
var Files embed.FS

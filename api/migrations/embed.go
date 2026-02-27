package migrations

import _ "embed"

// BootstrapSQL contains the idempotent schema bootstrap.
//
//go:embed bootstrap.sql
var BootstrapSQL string

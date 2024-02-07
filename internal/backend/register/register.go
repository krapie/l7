package register

import "github.com/krapie/plumber/internal/backend/registry"

type Register interface {
	SetTarget(target string)
	SetRegistry(registry *registry.BackendRegistry)
	SetAdditionalTable(table registry.Table)
	Observe()
}

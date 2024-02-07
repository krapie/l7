package registry

type Table interface {
	Add(backend string) error
	Remove(backend string) error
}

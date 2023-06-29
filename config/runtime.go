package config

type Runtime interface {
	Version() string
}

// Version TODO: generate that automatically
const (
	Version = "dev-0.0.1"
)

type runtime struct {
	version string
}

func NewRuntime(version string) Runtime {
	return &runtime{
		version: version,
	}
}

func (d runtime) Version() string {
	return d.version
}

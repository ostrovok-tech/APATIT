package version

// These constants define application metadata.
// They can be overridden at build time using ldflags.
// Example: go build -ldflags "-X 'apatit/internal/version.Version=1.1.0'"
var (
	Name    = "apatit"
	Version = "v1.0.0"
	Owner   = "ostrovok.tech"
)

const (
	Language = "go"
)

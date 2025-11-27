package version

// These constants define application metadata.
// They can be overridden at build time using ldflags.
// Example: go build -ldflags "-X 'ping-admin-exporter/internal/version.Version=1.1.0'"
var (
	Name    = "ping_admin_exporter"
	Version = "v1.0.0"
	Owner   = "sre"
)

const (
	Language = "go"
)

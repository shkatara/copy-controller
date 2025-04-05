package types

import "time"

const (
	// DefaultPort is the default port for the server
	DefaultPort = 8080
	// DefaultTimeout is the default timeout for the server
	DefaultTimeout = 30
	// DefaultIdleTimeout is the default idle timeout for the server
	DefaultIdleTimeout = 60
)

type ControllerConfig struct {
	Interval time.Duration
}

var (
	Interval string
)

package server

// Logging provides access to loggers
type Logging interface {
	// New returns the logger for the host. If logToStdout is true, then
	// prints logs also to stdout
	New(h *Host, logToStdout bool) Logger
}

// Logger provides the printing functions. A logger is associated with
// a single host
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
}

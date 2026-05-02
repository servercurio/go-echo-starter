package router

// Config carries cross-cutting dependencies that route/module constructors
// need access to without coupling them to a concrete server implementation.
type Config struct {
	// ReadinessProbe reports whether the application is currently able to
	// serve traffic. Health endpoints invoke it to decide between 200 OK and
	// a 503 Service Unavailable response.
	//
	// The default returned by NewConfig always reports ready, so routes that
	// don't care still work; lifecycle owners (e.g. the application daemon)
	// override it to surface real readiness.
	ReadinessProbe func() bool
}

func NewConfig() *Config {
	return &Config{
		ReadinessProbe: func() bool { return true },
	}
}

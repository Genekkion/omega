package config

var (
	DefaultConfig = Config{
		Ignore:   []string{".git", "bin"},
		Commands: []string{"echo \"Hello from Omega!\""},
		LogLevel: "debug",
		Timeout:  100,
		Delay:    200,
	}
)

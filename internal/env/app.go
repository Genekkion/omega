package env

var (
	ConfigPath    = getStringEnv("CONFIG_PATH", "./.omega.json")
	ProgramName   = getStringEnv("PROGRAM", "omega")
	BaseDirectory = getStringEnv("BASE_DIRECTORY", ".")
	LogFile       = getStringEnv("LOG_FILE", "logs/omega.log")
)

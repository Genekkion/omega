package env

import (
	"omega/internal/log"
	"os"
	"strconv"
	"strings"
)

func getEnv[T any](key string, defaultValue T,
	adapter func(value string) (T, error)) T {
	valueString, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := adapter(valueString)
	if err != nil {
		log.Warn("Environment variable \"" + key + "\" not valid, using default value")
		return defaultValue
	}
	return value
}

func getEnvOrPanic(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal("Required environment variable \"" + key + "\" not set, exiting")
	}
	return value
}

func getStringEnv(key string, defaultValue string) string {
	return getEnv(key, defaultValue, func(x string) (string, error) { return x, nil })
}

func getIntEnv(key string, defaultValue int) int {
	return getEnv(key, defaultValue, strconv.Atoi)
}

func getBoolEnv(key string, defaultValue bool) bool {
	return getEnv(key, defaultValue, func(value string) (bool, error) {
		return strings.ToLower(value) == "true", nil
	})
}

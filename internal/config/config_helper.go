package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}

	return v
}

func getEnv(key string, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}

		log.Printf("config: invalid int %s=%q, using %d", key, v, defaultValue)
	}

	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}

		log.Printf("config: invalid bool %s=%q, using %t", key, v, defaultValue)
	}

	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}

		log.Printf("config: invalid duration %s=%q, using %s", key, v, defaultValue)
	}

	return defaultValue
}

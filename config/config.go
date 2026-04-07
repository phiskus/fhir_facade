package config
import (
	"os"
	"strconv"
	"fmt"
	"github.com/joho/godotenv"
)

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	//Relevant DAtabase and API configuration fields go here
	Env string
	Port string
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	DBSSLMode string
	DBMaxOpenConns int
	DBMaxIdleConns int

}

func Load() (*Config, error) {

	_ = godotenv.Load() // Load .env file if it exists
	cfg := &Config{
		Env:            getEnvOrDefault("ENV", "local"),
		Port:           getEnvOrDefault("PORT", "8080"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         getEnvOrDefault("DB_PORT","5432"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      getEnvOrDefault("DB_SSLMODE", "disable"),
		DBMaxOpenConns: parseInt(getEnvOrDefault("DB_MAX_OPEN_CONNS", "10")),
		DBMaxIdleConns: parseInt(getEnvOrDefault("DB_MAX_IDLE_CONNS", "5")),
	}
	

// Validate required fields — fail loudly rather than silently misbehave
	required := map[string]string{
		
		"DB_HOST":  cfg.DBHost,
		"DB_USER":  cfg.DBUser,
		"DB_PASSWORD": cfg.DBPassword,
		"DB_NAME":  cfg.DBName,
		
	}
	for key, value := range required {
		if value == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	return cfg, nil

}
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
// config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Config
	Port string
	Env  string

	// Database
	UserDB     string
	PasswordDB string
	ServerDB   string
	Database   string
	PortDB     string

	// Base de datos
	DBPath string

	// Auth
	JWTSecret        string
	JWTSalt          string
	JWTExpireTime    int
	CookieSecret     string
	SessionTTL       int
	RefreshThreshold int

	// Redis
	RedisTTL      int
	RedisURL      string
	RedisPort     string
	RedisPassword string
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	// Cargar .env si existe
	err := godotenv.Load()
	if err != nil {
		log.Printf("⚠️  No se pudo cargar el archivo .env: %v", err)
	}

	// Construir Redis URL si no está definida completamente
	redisURL := getEnv("REDIS_URL", "")
	if redisURL == "" {
		// Construir URL desde componentes
		redisHost := getEnv("REDIS_URL", "localhost")
		redisPort := getEnv("REDIS_PORT", "6379")
		redisPassword := getEnv("REDIS_PASSWORD", "")

		if redisPassword != "" {
			redisURL = fmt.Sprintf("redis://:%s@%s:%s", redisPassword, redisHost, redisPort)
		} else {
			redisURL = fmt.Sprintf("redis://%s:%s", redisHost, redisPort)
		}
	}

	return &Config{
		// Server Config
		Port: getEnv("PORT", "3000"),
		Env:  getEnv("ENV", "Production"),

		// Database
		DBPath: getEnv("DB_PATH", "app.db"),

		// Auth
		JWTSecret:     getEnv("JWT_SECRETO", ""),
		JWTSalt:       getEnv("JWT_SALT", "10"),
		JWTExpireTime: getEnvAsInt("JWT_TIEMPO_EXPIRA", 3600),

		//
		CookieSecret:     getEnv("COOKIE_SECRET", "default-cookie-secret-key-32-chars-long-123"),
		SessionTTL:       getEnvAsInt("SESSION_TTL", 3600),
		RefreshThreshold: getEnvAsInt("REFRESH_THRESHOLD", 300),

		// Redis
		RedisTTL:      getEnvAsInt("REDIS_TTL", 3600),
		RedisURL:      redisURL,
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}
}

// GetRedisConfigInfo retorna información de Redis (sin password) para logs
func (c *Config) GetRedisConfigInfo() string {
	// Ocultar password para seguridad
	maskedURL := c.RedisURL
	if c.RedisPassword != "" {
		// Si tenemos password separada, construir URL enmascarada
		maskedURL = fmt.Sprintf("redis://:***@%s:%s",
			getEnv("REDIS_URL", "localhost"),
			c.RedisPort)
	}
	return fmt.Sprintf("URL: %s, TTL: %ds", maskedURL, c.RedisTTL)
}

// getEnv obtiene variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtiene variable de entorno como entero
func getEnvAsInt(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return value
}

// IsDevelopment retorna true si el entorno es desarrollo
func (c *Config) IsDevelopment() bool {
	return c.Env == "Dev" || c.Env == "Development"
}

// IsProduction retorna true si el entorno es producción
func (c *Config) IsProduction() bool {
	return c.Env == "Production" || c.Env == "Prod"
}

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Clerk    ClerkConfig
	Logger   LoggerConfig
	Cache    CacheConfig
	Auth     types.Config
}

type ServerConfig struct {
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RequestTimeout time.Duration
	Middleware     MiddlewareConfig
}

type MiddlewareConfig struct {
	// CORS configuration
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int

	// Rate limiting configuration
	RateLimit struct {
		RequestsPerMinute int
		WindowLength      time.Duration
	}
}

type DatabaseConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Database    string
	Schema      string
	MaxConns    int32
	MinConns    int32
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
	HealthCheck time.Duration
	SSLMode     string
	SearchPath  string
}

type ClerkConfig struct {
	SecretKey     string
	WebhookSecret string
}

type LoggerConfig struct {
	Environment string
	Level       string
}

type CacheConfig struct {
	Host     string
	Port     int
	Password string
}

// Load reads configuration from environment variables and files
func Load() (*Config, error) {
	// Load .env file first if it exists
	if err := godotenv.Load(); err != nil {
		// Only return error if file exists but couldn't be read
		if !strings.Contains(err.Error(), "no such file or directory") {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Set up viper for YAML config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set up environment variable handling
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults()

	// Read YAML config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Parse durations
	if d, err := time.ParseDuration(viper.GetString("server.timeout.read")); err == nil {
		config.Server.ReadTimeout = d
	}
	if d, err := time.ParseDuration(viper.GetString("server.timeout.write")); err == nil {
		config.Server.WriteTimeout = d
	}
	if d, err := time.ParseDuration(viper.GetString("server.timeout.idle")); err == nil {
		config.Server.IdleTimeout = d
	}
	if d, err := time.ParseDuration(viper.GetString("server.timeout.request")); err == nil {
		config.Server.RequestTimeout = d
	}

	// Parse auth durations
	if d, err := time.ParseDuration(viper.GetString("auth.jwt.access_token_ttl")); err == nil {
		config.Auth.JWT.AccessTokenTTL = d
	}
	if d, err := time.ParseDuration(viper.GetString("auth.jwt.refresh_token_ttl")); err == nil {
		config.Auth.JWT.RefreshTokenTTL = d
	}

	fmt.Printf("config: %+v\n", config)
	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.timeout.read", "15s")
	viper.SetDefault("server.timeout.write", "15s")
	viper.SetDefault("server.timeout.idle", "60s")
	viper.SetDefault("server.timeout.request", "60s")

	// Middleware defaults
	viper.SetDefault("server.middleware.allowedOrigins", []string{"https://*", "http://*"})
	viper.SetDefault("server.middleware.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("server.middleware.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	viper.SetDefault("server.middleware.exposedHeaders", []string{"Link"})
	viper.SetDefault("server.middleware.allowCredentials", true)
	viper.SetDefault("server.middleware.maxAge", 300)
	viper.SetDefault("server.middleware.rateLimit.requestsPerMinute", 100)
	viper.SetDefault("server.middleware.rateLimit.windowLength", "1m")

	// Database defaults
	viper.SetDefault("database.maxConns", 25)
	viper.SetDefault("database.minConns", 5)
	viper.SetDefault("database.maxLifetime", "1h")
	viper.SetDefault("database.maxIdleTime", "30m")
	viper.SetDefault("database.healthCheck", "1m")
	viper.SetDefault("database.sslMode", "require")

	// Logger defaults
	viper.SetDefault("logger.environment", "development")
	viper.SetDefault("logger.level", "info")

	// Auth defaults
	viper.SetDefault("auth.jwt.access_token_ttl", "15m")
	viper.SetDefault("auth.jwt.refresh_token_ttl", "7d")
	viper.SetDefault("auth.cookie.path", "/")
	viper.SetDefault("auth.cookie.secure", true)
	viper.SetDefault("auth.cookie.same_site", "strict")
}

// GetDSN returns the formatted database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
		c.SearchPath,
	)
}

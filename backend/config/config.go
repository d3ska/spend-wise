package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

// defaultJWTSecret is the placeholder JWT secret that ships with the codebase.
// It is intentionally weak so that developers notice and override it.
const defaultJWTSecret = "change-me-in-production"

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Auth          AuthConfig
	CORS          CORSConfig
	Invite        InviteConfig
	EnableBanking EnableBankingConfig
}

type InviteConfig struct {
	DefaultExpiryHours int `env:"INVITE_DEFAULT_EXPIRY_HOURS" envDefault:"168"` // 7 days
}

type EnableBankingConfig struct {
	ApplicationID string `env:"ENABLEBANKING_APP_ID"`
	KeyPath       string `env:"ENABLEBANKING_KEY_PATH"`
	BaseURL       string `env:"ENABLEBANKING_BASE_URL" envDefault:"https://api.enablebanking.com"`
}

type ServerConfig struct {
	Host         string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port         int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
}

// Addr returns the host:port address string.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres"`
	Name     string `env:"DB_NAME" envDefault:"spendwise"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns int32  `env:"DB_MAX_CONNS" envDefault:"10"`
}

// DSN returns a PostgreSQL connection string.
// WARNING: The returned string contains the plaintext password.
// Use RedactedDSN() for logging.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// RedactedDSN returns a connection string with the password masked,
// safe for inclusion in log messages.
func (d DatabaseConfig) RedactedDSN() string {
	return fmt.Sprintf("postgres://%s:***@%s:%d/%s?sslmode=%s",
		d.User, d.Host, d.Port, d.Name, d.SSLMode)
}

type AuthConfig struct {
	AuthBypass bool          `env:"AUTH_BYPASS" envDefault:"false"`
	JWTSecret  string        `env:"JWT_SECRET" envDefault:"change-me-in-production"` // MUST be overridden via env var
	JWTExpiry  time.Duration `env:"JWT_EXPIRY" envDefault:"24h"`

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL" envDefault:"https://localhost:3000/auth/google/callback"`

	GitHubClientID     string `env:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string `env:"GITHUB_CLIENT_SECRET"`
	GitHubRedirectURL  string `env:"GITHUB_REDIRECT_URL" envDefault:"https://localhost:3000/auth/github/callback"`

	CookieDomain string `env:"COOKIE_DOMAIN"`
	CookieSecure bool   `env:"COOKIE_SECURE" envDefault:"true"`
}

type CORSConfig struct {
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"https://localhost:3000"`
}

// Load parses environment variables into a Config struct.
// It logs warnings for insecure defaults that should be overridden.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Auth.JWTSecret == defaultJWTSecret {
		slog.Warn("JWT_SECRET is using the insecure default value — set a strong secret via the JWT_SECRET environment variable")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		slog.Warn("JWT_SECRET is shorter than 32 characters — consider using a longer, random secret")
	}

	return cfg, nil
}

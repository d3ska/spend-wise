package config

import (
	"strings"
	"testing"
)

func TestServerConfig_Addr(t *testing.T) {
	tests := map[string]struct {
		config ServerConfig
		want   string
	}{
		"default host and port": {
			config: ServerConfig{Host: "0.0.0.0", Port: 8080},
			want:   "0.0.0.0:8080",
		},
		"localhost with custom port": {
			config: ServerConfig{Host: "127.0.0.1", Port: 3000},
			want:   "127.0.0.1:3000",
		},
		"empty host with port": {
			config: ServerConfig{Host: "", Port: 9000},
			want:   ":9000",
		},
		"IPv6 localhost": {
			config: ServerConfig{Host: "::1", Port: 8080},
			want:   "::1:8080",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.config.Addr()
			if got != tt.want {
				t.Errorf("ServerConfig.Addr() got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := map[string]struct {
		config DatabaseConfig
		want   string
	}{
		"default values": {
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Name:     "spendwise",
				SSLMode:  "disable",
			},
			want: "postgres://postgres:postgres@localhost:5432/spendwise?sslmode=disable",
		},
		"custom values": {
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret123",
				Name:     "production_db",
				SSLMode:  "require",
			},
			want: "postgres://admin:secret123@db.example.com:5433/production_db?sslmode=require",
		},
		"special characters in password": {
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "p@ssw0rd!#$",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			want: "postgres://user:p@ssw0rd!#$@localhost:5432/testdb?sslmode=disable",
		},
		"sslmode verify-full": {
			config: DatabaseConfig{
				Host:     "secure.db.com",
				Port:     5432,
				User:     "secureuser",
				Password: "securepass",
				Name:     "securedb",
				SSLMode:  "verify-full",
			},
			want: "postgres://secureuser:securepass@secure.db.com:5432/securedb?sslmode=verify-full",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.config.DSN()
			if got != tt.want {
				t.Errorf("DatabaseConfig.DSN() got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDatabaseConfig_RedactedDSN(t *testing.T) {
	tests := map[string]struct {
		config DatabaseConfig
		want   string
	}{
		"password is masked": {
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "super-secret-password",
				Name:     "spendwise",
				SSLMode:  "disable",
			},
			want: "postgres://postgres:***@localhost:5432/spendwise?sslmode=disable",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.config.RedactedDSN()
			if got != tt.want {
				t.Errorf("DatabaseConfig.RedactedDSN() got %q, want %q", got, tt.want)
			}
			if strings.Contains(got, tt.config.Password) {
				t.Error("RedactedDSN() must not contain the plaintext password")
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		envVars map[string]string
		check   func(*testing.T, Config)
	}{
		"default port is 8080": {
			envVars: map[string]string{},
			check: func(t *testing.T, cfg Config) {
				if cfg.Server.Port != 8080 {
					t.Errorf("Server.Port got %d, want 8080", cfg.Server.Port)
				}
			},
		},
		"default DB name is spendwise": {
			envVars: map[string]string{},
			check: func(t *testing.T, cfg Config) {
				if cfg.Database.Name != "spendwise" {
					t.Errorf("Database.Name got %q, want %q", cfg.Database.Name, "spendwise")
				}
			},
		},
		"default JWT secret matches defaultJWTSecret constant": {
			envVars: map[string]string{},
			check: func(t *testing.T, cfg Config) {
				if cfg.Auth.JWTSecret != defaultJWTSecret {
					t.Errorf("Auth.JWTSecret got %q, want %q", cfg.Auth.JWTSecret, defaultJWTSecret)
				}
			},
		},
		"custom server port": {
			envVars: map[string]string{
				"SERVER_PORT": "3000",
			},
			check: func(t *testing.T, cfg Config) {
				if cfg.Server.Port != 3000 {
					t.Errorf("Server.Port got %d, want 3000", cfg.Server.Port)
				}
			},
		},
		"custom database config": {
			envVars: map[string]string{
				"DB_HOST":     "custom.db.com",
				"DB_PORT":     "5433",
				"DB_USER":     "customuser",
				"DB_PASSWORD": "custompass",
				"DB_NAME":     "customdb",
			},
			check: func(t *testing.T, cfg Config) {
				if cfg.Database.Host != "custom.db.com" {
					t.Errorf("Database.Host got %q, want %q", cfg.Database.Host, "custom.db.com")
				}
				if cfg.Database.Port != 5433 {
					t.Errorf("Database.Port got %d, want 5433", cfg.Database.Port)
				}
				if cfg.Database.User != "customuser" {
					t.Errorf("Database.User got %q, want %q", cfg.Database.User, "customuser")
				}
				if cfg.Database.Password != "custompass" {
					t.Errorf("Database.Password got %q, want %q", cfg.Database.Password, "custompass")
				}
				if cfg.Database.Name != "customdb" {
					t.Errorf("Database.Name got %q, want %q", cfg.Database.Name, "customdb")
				}
			},
		},
		"custom JWT secret": {
			envVars: map[string]string{
				"JWT_SECRET": "my-super-secret-key",
			},
			check: func(t *testing.T, cfg Config) {
				if cfg.Auth.JWTSecret != "my-super-secret-key" {
					t.Errorf("Auth.JWTSecret got %q, want %q", cfg.Auth.JWTSecret, "my-super-secret-key")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Set environment variables for this test
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			tt.check(t, cfg)
		})
	}
}

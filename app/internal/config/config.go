package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port                   int
	ShutdownTimeoutSeconds int
}

type DatabaseConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	Name                   string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

type CacheConfig struct {
	TaskListTTLSeconds int
}

type RateLimitConfig struct {
	RequestsPerMinute int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Pagination PaginationConfig
	Cache      CacheConfig
	RateLimit  RateLimitConfig
	JWT        JWTConfig
}

type yamlConfig struct {
	Server struct {
		Port                   int `yaml:"port"`
		ShutdownTimeoutSeconds int `yaml:"shutdown_timeout_seconds"`
	} `yaml:"server"`
	Database struct {
		MaxOpenConns           int `yaml:"max_open_conns"`
		MaxIdleConns           int `yaml:"max_idle_conns"`
		ConnMaxLifetimeMinutes int `yaml:"conn_max_lifetime_minutes"`
	} `yaml:"database"`
	Pagination struct {
		DefaultPageSize int `yaml:"default_page_size"`
		MaxPageSize     int `yaml:"max_page_size"`
	} `yaml:"pagination"`
	Cache struct {
		TaskListTTLSeconds int `yaml:"task_list_ttl_seconds"`
	} `yaml:"cache"`
	RateLimit struct {
		RequestsPerMinute int `yaml:"requests_per_minute"`
	} `yaml:"rate_limit"`
	JWT struct {
		ExpiryHours int `yaml:"expiry_hours"`
	} `yaml:"jwt"`
}

func (d DatabaseConfig) ConnMaxLifetime() time.Duration {
	return time.Duration(d.ConnMaxLifetimeMinutes) * time.Minute
}

func (s ServerConfig) ShutdownTimeout() time.Duration {
	return time.Duration(s.ShutdownTimeoutSeconds) * time.Second
}

func (j JWTConfig) Expiry() time.Duration {
	return time.Duration(j.ExpiryHours) * time.Hour
}

func (c CacheConfig) TaskListTTL() time.Duration {
	return time.Duration(c.TaskListTTLSeconds) * time.Second
}

func Load(yamlPath, envPath string) (*Config, error) {
	_ = godotenv.Load(envPath)

	yc, err := loadYAML(yamlPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:                   yc.Server.Port,
			ShutdownTimeoutSeconds: yc.Server.ShutdownTimeoutSeconds,
		},
		Database: DatabaseConfig{
			Host:                   "localhost",
			Port:                   "3306",
			User:                   "root",
			Password:               "",
			Name:                   "teamtask",
			MaxOpenConns:           yc.Database.MaxOpenConns,
			MaxIdleConns:           yc.Database.MaxIdleConns,
			ConnMaxLifetimeMinutes: yc.Database.ConnMaxLifetimeMinutes,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
		Pagination: PaginationConfig{
			DefaultPageSize: yc.Pagination.DefaultPageSize,
			MaxPageSize:     yc.Pagination.MaxPageSize,
		},
		Cache: CacheConfig{
			TaskListTTLSeconds: yc.Cache.TaskListTTLSeconds,
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: yc.RateLimit.RequestsPerMinute,
		},
		JWT: JWTConfig{
			ExpiryHours: yc.JWT.ExpiryHours,
		},
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

func loadYAML(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	yc := &yamlConfig{}
	if err := yaml.Unmarshal(data, yc); err != nil {
		return nil, err
	}
	return yc, nil
}

func applyEnvOverrides(cfg *Config) {
	overrideInt(&cfg.Server.Port, "SERVER_PORT")

	overrideString(&cfg.Database.Host, "DB_HOST")
	overrideString(&cfg.Database.Port, "DB_PORT")
	overrideString(&cfg.Database.User, "DB_USER")
	overrideString(&cfg.Database.Password, "DB_PASSWORD")
	overrideString(&cfg.Database.Name, "DB_NAME")

	overrideString(&cfg.Redis.Host, "REDIS_HOST")
	overrideString(&cfg.Redis.Port, "REDIS_PORT")
	overrideString(&cfg.Redis.Password, "REDIS_PASSWORD")

	overrideString(&cfg.JWT.Secret, "JWT_SECRET")
}

func overrideString(field *string, envKey string) {
	if v := os.Getenv(envKey); v != "" {
		*field = v
	}
}

func overrideInt(field *int, envKey string) {
	if v := os.Getenv(envKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*field = n
		}
	}
}

func DefaultPaths() (yamlPath, envPath string) {
	return "config.yaml", filepath.Join(".", ".env")
}

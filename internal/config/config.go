package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	App      AppConfig      `yaml:"app"`
	Postgres PostgresConfig `yaml:"postgres"`
	Session  SessionConfig  `yaml:"session"`
}

// AppConfig 应用配置
type AppConfig struct {
	Env       string `yaml:"env"`        // development/production
	HTTPPort  int    `yaml:"http_port"`  // HTTP 端口
	GRPCPort  int    `yaml:"grpc_port"`  // gRPC 端口
	MachineID int    `yaml:"machine_id"` // 机器码 (0-63)
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	User        string        `yaml:"user"`
	Password    string        `yaml:"password"`
	Database    string        `yaml:"database"`
	MaxConns    int           `yaml:"max_conns"`
	MinConns    int           `yaml:"min_conns"`
	MaxConnLife time.Duration `yaml:"max_conn_life"`
	MaxConnIdle time.Duration `yaml:"max_conn_idle"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	TTL time.Duration `yaml:"ttl"` // 会话有效期
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Env:       "development",
			HTTPPort:  8080,
			GRPCPort:  9090,
			MachineID: 1,
		},
		Postgres: PostgresConfig{
			Host:        "localhost",
			Port:        5432,
			User:        "wallet",
			Password:    "wallet123",
			Database:    "wallet",
			MaxConns:    100,
			MinConns:    10,
			MaxConnLife: time.Hour,
			MaxConnIdle: time.Minute * 30,
		},
		Session: SessionConfig{
			TTL: time.Hour * 24,
		},
	}
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	config := DefaultConfig()

	if path == "" {
		// 尝试默认路径
		defaultPaths := []string{"config.yaml", "config/config.yaml", "/etc/wallet/config.yaml"}
		for _, p := range defaultPaths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// 环境变量覆盖
	config.loadFromEnv()

	return config, nil
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() {
	if env := os.Getenv("APP_ENV"); env != "" {
		c.App.Env = env
	}
	if port := os.Getenv("APP_HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.App.HTTPPort = p
		}
	}
	if port := os.Getenv("APP_GRPC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.App.GRPCPort = p
		}
	}
	if machineID := os.Getenv("APP_MACHINE_ID"); machineID != "" {
		if id, err := strconv.ParseInt(machineID, 10, 64); err == nil {
			c.App.MachineID = int(id)
		}
	}

	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		c.Postgres.Host = host
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Postgres.Port = p
		}
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		c.Postgres.User = user
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		c.Postgres.Password = password
	}
	if database := os.Getenv("POSTGRES_DB"); database != "" {
		c.Postgres.Database = database
	}
}

// DSN 返回 PostgreSQL 连接字符串
func (c *PostgresConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + strconv.Itoa(c.Port) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Database +
		" sslmode=disable"
}

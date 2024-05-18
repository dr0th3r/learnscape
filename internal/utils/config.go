package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const configPath = "../config/config.json"

type Config struct {
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"db"`
	App    AppConfig    `json:"app"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int16  `json:"port"`
}

type DBConfig struct {
	Host          string `json:"host"`
	Port          int16  `json:"port"`
	User          string `json:"user"`
	Password      string `json:"password"`
	Name          string `json:"name"`
	MigrationsDir string `json:"migrationsDir"`
}

func (c DBConfig) GetConnectionUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Name)
}

func (c DBConfig) GetConnectionUrlWithoutName() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}

type AppConfig struct {
	TimeFormat string `json:"timeFormat"`
	JwtSecret  string
}

func ParseConfig() (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func WithAppConfig(next http.Handler, config AppConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "config", config)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

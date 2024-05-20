package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const configPath = "config/config.json"

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
	JwtSecret string `json:"jwtSecret"`
}

func getProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dirName := filepath.Base(currentDir)

	if dirName == "test" {
		return "../", nil
	}

	return "", nil
}

func ParseConfig() (*Config, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fmt.Sprint(projectRoot + configPath))
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

	config.DB.MigrationsDir = projectRoot + config.DB.MigrationsDir

	return &config, nil
}

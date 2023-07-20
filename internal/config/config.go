package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type UrlConfig struct {
	LocalURL  string `yaml:"local_url"`
	DockerURL string `yaml:"docker_url"`
}

func (u UrlConfig) URL() string {
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "local" || mode == "" {
		return u.LocalURL
	}
	return u.DockerURL
}

type RedisConfig struct {
	LocalURL  string `yaml:"local_url"`
	DockerURL string `yaml:"docker_url"`
}

func (u RedisConfig) URL() string {
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "local" || mode == "" {
		return u.LocalURL
	}
	return u.DockerURL
}

type PostgresConfig struct {
	LocalURL  string `yaml:"local_url"`
	DockerURL string `yaml:"docker_url"`
}

func (u PostgresConfig) URL() string {
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "local" || mode == "" {
		return u.LocalURL
	}
	return u.DockerURL
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type VkConfig struct {
	Key      string `yaml:"secret_key"`
	DebugKey string `yaml:"debug_key"`
}

type TickerConfig struct {
	LocalURL  string `yaml:"local_url"`
	DockerURL string `yaml:"docker_url"`
}

func (u TickerConfig) URL() string {
	mode := strings.ToLower(os.Getenv("MODE"))
	if mode == "local" || mode == "" {
		return u.LocalURL
	}
	return u.DockerURL
}

type Config struct {
	Redis    RedisConfig    `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
	Server   ServerConfig   `yaml:"server"`
	VK       VkConfig       `yaml:"vk"`
	Ticker   TickerConfig   `yaml:"ticker"`
}

func New(
	redis RedisConfig,
	postgres PostgresConfig,
	server ServerConfig,
	ticker TickerConfig,
) *Config {
	return &Config{
		Redis:    redis,
		Postgres: postgres,
		Server:   server,
		Ticker:   ticker,
	}
}

func FromFile(filePath string) *Config {
	config := new(Config)
	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error while read file, %s", err)
	}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		log.Fatalf("error while unmarshal file into config, %s", err)
	}
	return config
}

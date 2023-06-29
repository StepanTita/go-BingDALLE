package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config interface {
	Logger
	Runtime
	Networker
	Authenticator
}

type config struct {
	Logger
	Runtime
	Networker
	Authenticator
}

type CliConfig struct {
	LogLevel string

	ApiUrl string

	Proxy string

	UCookie string
}

type YamlDALLEConfig struct {
	LogLevel string `yaml:"log_level"`

	ApiUrl string `yaml:"api_url"`
	Proxy  string `yaml:"proxy"`

	UCookie string `yaml:"u_cookie"`
}

type yamlConfig struct {
	Dalle YamlDALLEConfig `yaml:"dalle"`
}

func NewFromDALLEConfig(cfg YamlDALLEConfig) Config {
	return &config{
		Logger:        NewLogger(cfg.LogLevel),
		Runtime:       NewRuntime(Version),
		Networker:     NewNetworker(cfg.ApiUrl, cfg.Proxy),
		Authenticator: NewAuthenticator(cfg.UCookie),
	}
}

func NewFromFile(path string) Config {
	cfg := yamlConfig{}

	yamlConfig, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read config %s", path))
	}

	err = yaml.Unmarshal(yamlConfig, &cfg)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unmarshal config %s", path))
	}

	return NewFromDALLEConfig(cfg.Dalle)
}

func NewFromCLI(cfg CliConfig) Config {
	return &config{
		Logger:        NewLogger(cfg.LogLevel),
		Runtime:       NewRuntime(Version),
		Networker:     NewNetworker(cfg.ApiUrl, cfg.Proxy),
		Authenticator: NewAuthenticator(cfg.UCookie),
	}
}

package config

import (
	_ "embed"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/repo-scm/proxy/utils"
)

//go:embed proxy.yaml
var configData string

type Config struct {
	Gerrits []Gerrit `yaml:"gerrits"`
}

type Gerrit struct {
	SiteName string `yaml:"site_name"`
	ApiBase  string `yaml:"api_base"`
	UserName string `yaml:"user_name"`
	UserPass string `yaml:"user_pass"`
}

func LoadConfig(name string) (*Config, error) {
	var config Config

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if name != "" {
		viper.SetConfigFile(name)
	} else {
		viper.AddConfigPath(path.Join(home, ".repo-scm"))
		viper.SetConfigName("proxy")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if name == "" {
			name = path.Join(home, ".repo-scm", "proxy.yaml")
		}
		if err := createConfig(name); err != nil {
			return nil, errors.Wrap(err, "failed to read or create config\n")
		}
		viper.SetConfigFile(name)
		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config after creation\n")
		}
	}

	buf, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func createConfig(name string) error {
	if err := os.MkdirAll(path.Dir(name), utils.PermDir); err != nil {
		return err
	}

	if err := os.WriteFile(name, []byte(configData), utils.PermFile); err != nil {
		return err
	}

	return nil
}

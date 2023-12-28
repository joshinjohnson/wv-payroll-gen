package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	filename = ".payroll"
)

type Config struct {
	ServerAddress string   `mapstructure:"SERVER_ADDRESS"`
	AuthToken     string   `mapstructure:"AUTH_TOKEN"`
	LogMode       string   `mapstructure:"LOG_MODE"`
	DbConfig      DbConfig `mapstructure:"DB_CONFIG"`
}

type DbConfig struct {
	User         string `mapstructure:"USER"`
	Password     string `mapstructure:"PASSWORD"`
	Hostname     string `mapstructure:"HOSTNAME"`
	Port         string `mapstructure:"PORT"`
	DatabaseName string `mapstructure:"DB_NAME"`
	SSLMode      string `mapstructure:"SSL_MODE"`
	SchemaName   string `mapstructure:"SCHEMA_NAME"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(filename)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Errorf("error unmarshalling config file: %s", err)
		os.Exit(1)
	}

	return &cfg, nil
}

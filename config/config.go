package config

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	ApiUrl   string     `mapstructure:"api_url"`
	CronSpec string     `mapstructure:"cron_spec"`
	Auth     AuthConfig `mapstructure:"auth"`
	Log      LogConfig  `mapstructure:"log"`
}

type AuthConfig struct {
	NameEnv     string `mapstructure:"name_env"`
	PasswordEnv string `mapstructure:"password_env"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func setDefaults() {
	viper.SetDefault("api_url", "http://localhost:8080")
	viper.SetDefault("cron_spec", "@every 5m")
	viper.SetDefault("auth", AuthConfig{
		NameEnv:     "PINGER_NAME",
		PasswordEnv: "PINGER_PASSWORD",
	})
	viper.SetDefault("log", LogConfig{Level: "info"})
}

func (c *AuthConfig) GetCredentials() (string, string, error) {
	name := viper.GetString(c.NameEnv)
	password := viper.GetString(c.PasswordEnv)

	if name == "" || password == "" {
		return "", "", fmt.Errorf("config - AuthConfig.GetCredentials: missing credentials")
	}

	return name, password, nil
}

func LoadConfig(configPath string) (*Config, error) {
	setDefaults()

	if configPath != "" {
		// Use config file from the arguments.
		viper.SetConfigFile(configPath)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".docker-pinger" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".docker-pinger")
		err = viper.SafeWriteConfig()
		if err != nil {
			var configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
			if !errors.As(err, &configFileAlreadyExistsError) {
				return nil, fmt.Errorf("config - LoadConfig - SafeWriteConfig: %v", err)
			}
		}
	}

	viper.SetEnvPrefix("pinger")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config - LoadConfig - ReadInConfig: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("config - LoadConfig - viper.Unmarshal: %v", err)
	}

	return &config, nil
}

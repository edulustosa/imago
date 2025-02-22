package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Env struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	Addr      string `mapstructure:"SERVER_PORT"`
	JWTSecret string `mapstructure:"JWT_SECRET"`

	AWSSecretKey string `mapstructure:"AWS_SECRET_KEY"`
	AWSAccessKey string `mapstructure:"AWS_ACCESS_KEY"`
	BucketName   string `mapstructure:"BUCKET_NAME"`
	AWSRegion    string `mapstructure:"AWS_REGION"`
}

func LoadEnv(envPath string) (*Env, error) {
	viper.AutomaticEnv()

	viper.SetConfigType("env")
	viper.AddConfigPath(envPath)
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		envs := []string{
			"DATABASE_URL",
			"SERVER_PORT",
			"JWT_SECRET",
			"BUCKET_NAME",
			"AWS_REGION",
			"AWS_SECRET_KEY",
			"AWS_ACCESS_KEY",
		}

		for _, env := range envs {
			if err := viper.BindEnv(env); err != nil {
				return nil, fmt.Errorf("failed to bind env: %w", err)
			}
		}
	}

	var env Env
	if err := viper.Unmarshal(&env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}

	return &env, nil
}

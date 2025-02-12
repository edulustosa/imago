package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Env struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	Addr      string `mapstructure:"SERVER_PORT"`
	JWTSecret string `mapstructure:"JWT_SECRET"`

	BucketName string `mapstructure:"BUCKET_NAME"`
	AWSRegion  string `mapstructure:"AWS_REGION"`
}

func LoadEnv(envPath string) (*Env, error) {
	viper.SetConfigType("env")
	viper.AddConfigPath(envPath)
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	var env Env
	if err := viper.Unmarshal(&env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env: %w", err)
	}

	return &env, nil
}

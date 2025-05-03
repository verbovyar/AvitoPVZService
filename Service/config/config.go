package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Port             string `mapstructure:"PORT"`
	ConnectingString string `mapstructure:"CONNECTING_STRING"`
	GrpcPort         string `mapstructure:"GRPC_PORT"`
	NetworkType      string `mapstructure:"NETWORK_TYPE"`

	DBHost     string `env:"DB_HOST" env-default:"localhost"`
	DBPort     string `env:"DB_PORT" env-default:"5432"`
	DBUser     string `env:"DB_USER" env-default:"postgres"`
	DBPassword string `env:"DB_PASSWORD" env-default:"postgres"`
	DBName     string `env:"DB_NAME" env-default:"avito_pvz"`
	RunMigs    bool   `env:"RUN_MIGRATIONS" env-default:"false"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("conf")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		_ = fmt.Errorf("do not parse config file:%v", err)
	}

	err = viper.Unmarshal(&config)

	return
}

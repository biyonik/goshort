package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	URL      URLConfig
}

type ServerConfig struct {
	Port    string
	BaseURL string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

type RedisConfig struct {
	Host string
	Port string
	DB   int
}

type URLConfig struct {
	ShortCodeLength int
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("BASE_URL", "http://localhost:8080")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "goshort")
	viper.SetDefault("POSTGRES_PASSWORD", "goshort_secret")
	viper.SetDefault("POSTGRES_DB", "goshort")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("SHORT_CODE_LENGTH", 7)

	// .env dosyası yoksa hata verme, default'lar kullanılsın
	_ = viper.ReadInConfig()

	return &Config{
		Server: ServerConfig{
			Port:    viper.GetString("SERVER_PORT"),
			BaseURL: viper.GetString("BASE_URL"),
		},
		Postgres: PostgresConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetString("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			DB:       viper.GetString("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host: viper.GetString("REDIS_HOST"),
			Port: viper.GetString("REDIS_PORT"),
			DB:   viper.GetInt("REDIS_DB"),
		},
		URL: URLConfig{
			ShortCodeLength: viper.GetInt("SHORT_CODE_LENGTH"),
		},
	}, nil
}

// DSN returns PostgreSQL connection string
func (c *PostgresConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DB + "?sslmode=disable"
}

// Addr returns Redis address
func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

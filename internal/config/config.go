package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Load reads configuration
func Load(path string) error {
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

// DSN returns a PostgreSQL connection string
func DSN() string {
	return "host=" + viper.GetString("postgres.host") +
		" port=" + viper.GetString("postgres.port") +
		" user=" + viper.GetString("postgres.user") +
		" password=" + viper.GetString("postgres.password") +
		" dbname=" + viper.GetString("postgres.db") +
		" sslmode=" + viper.GetString("postgres.sslmode")
}

func HTTPAddr() string {
	host := viper.GetString("app.host")
	port := viper.GetString("app.port")
	return host + ":" + port
}

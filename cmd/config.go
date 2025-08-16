package cmd

import (
	"strings"

	"github.com/marcbran/versource/internal"
	"github.com/spf13/viper"
)

func LoadConfig() (*internal.Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	tfConfig, err := LoadTerraformConfig()
	if err != nil {
		return nil, err
	}

	httpConfig, err := LoadHttpConfig()
	if err != nil {
		return nil, err
	}

	return &internal.Config{
		Database:  dbConfig,
		Terraform: tfConfig,
		HTTP:      httpConfig,
	}, nil
}

func LoadDatabaseConfig() (*internal.DatabaseConfig, error) {
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "3306")
	viper.SetDefault("db.user", "versource")
	viper.SetDefault("db.password", "versource")
	viper.SetDefault("db.dbname", "versource")
	viper.SetDefault("db.sslmode", "false")

	viper.BindEnv("db.host", "DB_HOST")
	viper.BindEnv("db.port", "DB_PORT")
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.dbname", "DB_NAME")
	viper.BindEnv("db.sslmode", "DB_SSLMODE")

	config := &internal.DatabaseConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		User:     viper.GetString("db.user"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	}

	return config, nil
}

func LoadTerraformConfig() (*internal.TerraformConfig, error) {
	viper.SetDefault("tf.workdir", "terraform")

	viper.BindEnv("tf.workdir", "TF_WORKDIR")

	return &internal.TerraformConfig{
		WorkDir: viper.GetString("tf.workdir"),
	}, nil
}

func LoadHttpConfig() (*internal.HttpConfig, error) {
	viper.SetDefault("http.hostname", "localhost")
	viper.SetDefault("http.port", "8080")

	viper.BindEnv("http.hostname", "HTTP_HOSTNAME")
	viper.BindEnv("http.port", "HTTP_PORT")

	return &internal.HttpConfig{
		Hostname: viper.GetString("http.hostname"),
		Port:     viper.GetString("http.port"),
	}, nil
}

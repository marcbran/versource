package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func LoadConfig(cmd *cobra.Command) (*internal.Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		v.AddConfigPath(xdgConfig + "/versource")
	} else if home := os.Getenv("HOME"); home != "" {
		v.AddConfigPath(home + "/.config/versource")
	}
	v.SetEnvPrefix("VS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	configKeyFlag, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	configKeyEnv := os.Getenv("VS_CONFIG")
	configKey := configKeyFlag
	if configKeyFlag == "default" && configKeyEnv != "" {
		configKey = configKeyEnv
	}
	if sub := v.Sub(configKey); sub != nil {
		v = sub
	}

	dbConfig := LoadDatabaseConfig(v)
	tfConfig, err := LoadTerraformConfig(v)
	if err != nil {
		return nil, err
	}
	httpConfig := LoadHttpConfig(v)

	return &internal.Config{
		Database:  dbConfig,
		Terraform: tfConfig,
		HTTP:      httpConfig,
	}, nil
}

func LoadDatabaseConfig(v *viper.Viper) *internal.DatabaseConfig {
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "3306")
	v.SetDefault("database.user", "versource")
	v.SetDefault("database.password", "versource")
	v.SetDefault("database.dbname", "versource")
	v.SetDefault("database.sslmode", "false")

	return &internal.DatabaseConfig{
		Host:     v.GetString("database.host"),
		Port:     v.GetString("database.port"),
		User:     v.GetString("database.user"),
		Password: v.GetString("database.password"),
		DBName:   v.GetString("database.dbname"),
		SSLMode:  v.GetString("database.sslmode"),
	}
}

func LoadTerraformConfig(v *viper.Viper) (*internal.TerraformConfig, error) {
	v.SetDefault("terraform.workdir", "terraform")

	workDir := v.GetString("terraform.workdir")
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, err
	}

	return &internal.TerraformConfig{
		WorkDir: absWorkDir,
	}, nil
}

func LoadHttpConfig(v *viper.Viper) *internal.HttpConfig {
	v.SetDefault("http.scheme", "http")
	v.SetDefault("http.hostname", "localhost")
	v.SetDefault("http.port", "8080")

	return &internal.HttpConfig{
		Scheme:   v.GetString("http.scheme"),
		Hostname: v.GetString("http.hostname"),
		Port:     v.GetString("http.port"),
	}
}

package internal

type Config struct {
	Database  *DatabaseConfig
	Terraform *TerraformConfig
	HTTP      *HttpConfig
}

type HttpConfig struct {
	Hostname string
	Port     string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type TerraformConfig struct {
	WorkDir string
}

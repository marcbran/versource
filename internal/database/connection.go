package database

import (
	"database/sql"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDb(config *versource.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	return db, nil
}

func NewGormDb(config *versource.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

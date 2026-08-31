package database

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	DBPath string
}

func GetConnection(config DBConfig) (*gorm.DB, error) {
	// Abrir conexión con SQLite
	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("error conectando a SQLite: %v", err)
	}

	// ⚠️ CRÍTICO PARA SQLITE: Habilitar claves foráneas (necesario para tus modelos)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("error habilitando foreign keys en SQLite: %v", err)
	}

	return db, nil
}

// HealthCheck se simplifica para SQLite
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error obteniendo sql.DB: %v", err)
	}
	return sqlDB.Ping()
}

func CloseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error obteniendo sql.DB: %v", err)
	}
	return sqlDB.Close()
}
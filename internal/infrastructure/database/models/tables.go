package models

import (
	"time"
)

// PrismaMigration model
type Migration struct {
	ID                string     `gorm:"primaryKey;size:36" json:"id"`
	Checksum          string     `gorm:"size:64;not null" json:"checksum"`
	FinishedAt        *time.Time `json:"finished_at"`
	MigrationName     string     `gorm:"size:255;not null" json:"migration_name"`
	Logs              *string    `gorm:"type:text" json:"logs"`
	RolledBackAt      *time.Time `json:"rolled_back_at"`
	StartedAt         time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	AppliedStepsCount int        `gorm:"not null;default:0" json:"applied_steps_count"`
}

func (Migration) TableName() string {
	return "_migrations"
}

// Estado model
type Estado struct {
	ID           int     `gorm:"primaryKey;autoIncrement" json:"id"`
	NombreEstado string  `gorm:"type:text;not null" json:"nombre_estado"`
	Descripcion  *string `gorm:"type:text" json:"descripcion,omitempty"`
}

// Rol model
type Rol struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	NombreRol   string `gorm:"type:text;not null" json:"nombre_rol"`
	Descripcion string `gorm:"type:text" json:"descripcion"`
}

func (Rol) TableName() string {
	return "roles"
}

// Permiso model
type Permiso struct {
	ID            int    `gorm:"primaryKey;autoIncrement" json:"id"`
	NombrePermiso string `gorm:"type:text;not null;uniqueIndex:uni_permisos_nombre_permiso" json:"nombre_permiso"`
	Descripcion   string `gorm:"type:text" json:"descripcion"`
}

// RolXPermiso model (tabla intermedia)
type RolXPermiso struct {
	IDRol     int `gorm:"primaryKey" json:"id_rol"`
	IDPermiso int `gorm:"primaryKey" json:"id_permiso"`

	Rol     Rol     `gorm:"foreignKey:IDRol;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"rol"`
	Permiso Permiso `gorm:"foreignKey:IDPermiso;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permiso"`
}

// Usuario model
type Usuario struct {
	Documento     string    `gorm:"primaryKey" json:"documento"`
	NomUser       string    `gorm:"type:text;not null" json:"nom_user"`
	Pass          string    `gorm:"type:text;not null" json:"pass"`
	IDRol         int       `gorm:"not null" json:"id_rol"`
	EstadoID      int       `gorm:"not null" json:"estado_id"`
	FechaRegistro time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"fecha_registro"`
	Rol           Rol       `gorm:"foreignKey:IDRol;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"rol"`
	Estado        Estado    `gorm:"foreignKey:EstadoID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"estado"`
}

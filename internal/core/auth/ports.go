// internal/core/auth/ports.go
package auth

import (
	"api_go/internal/infrastructure/database/models"
	"context"
)

type AuthPort interface {
	RetrieveUser(ctx context.Context, authData AuthData) (*User, error)
	RetrieveUserByID(ctx context.Context, documento string) (*User, error)
	GetPermissionsByRoleID(ctx context.Context, roleID int) ([]string, error)
}

type UserRepositoryPort interface {
	CreateUser(ctx context.Context, usuario *Usuario) (string, error)
}

type RolRepositoryPort interface {
	FindRolIDByName(ctx context.Context, nombre string) (int, error)
	FindByID(ctx context.Context, id int) (*models.Rol, error)
	FindByExactName(ctx context.Context, nombre string) (*models.Rol, error)
	FindOrCreateGuest(ctx context.Context) (int, error)
}

type EstadoRepositoryPort interface {
	FindEstadoIDByName(ctx context.Context, nombre string) (int, error)
	FindByExactName(ctx context.Context, nombre string) (*models.Estado, error)
	FindOrCreateActive(ctx context.Context) (int, error)
}

type AuthData struct {
	Documento string `json:"documento,omitempty"`
	Username  string `json:"username,omitempty"`
}

type User struct {
	ID           string   `json:"id"`
	Username     string   `json:"username,omitempty"`
	PasswordHash string   `json:"password_hash,omitempty"`
	IDRol        *int     `json:"id_rol,omitempty"`
	IDEstado     *int     `json:"id_estado,omitempty"`
	Permisos     []string `json:"permisos,omitempty"`
	RolNombre    string   `json:"rol_nombre,omitempty"`
}

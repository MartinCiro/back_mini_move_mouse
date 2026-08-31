// internal/core/auth/entities.go
package auth

import (
	"api_go/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ExpiresAt alias para time.Time para mayor claridad
type ExpiresAt = time.Time

type Usuario struct {
	Documento         string `json:"documento"`
	Username          string `json:"username,omitempty"`
	IDRol             *int   `json:"id_rol,omitempty"`
	IDEstado          *int   `json:"id_estado,omitempty"`
	encryptedPassword string
}

// NewUsuario crea un nuevo usuario con password encriptado
func NewUsuario(documento, username, password string, idRol, idEstado *int) (*Usuario, error) {
	encryptedPassword, err := encodePassword(password)
	if err != nil {
		return nil, err
	}
	return &Usuario{
		Documento:         documento,
		Username:          username,
		IDRol:             idRol,
		IDEstado:          idEstado,
		encryptedPassword: encryptedPassword,
	}, nil
}

// NewUsuarioFromEncrypted crea un usuario con password ya encriptado
func NewUsuarioFromEncrypted(documento, username, encryptedPassword string, idRol, idEstado *int) *Usuario {
	return &Usuario{
		Documento:         documento,
		Username:          username,
		IDRol:             idRol,
		IDEstado:          idEstado,
		encryptedPassword: encryptedPassword,
	}
}

// GetEncryptedPassword retorna el password encriptado
func (u *Usuario) GetEncryptedPassword() string {
	return u.encryptedPassword
}

// ComparePassword verifica si el password plano coincide con el encriptado
func (u *Usuario) ComparePassword(plainPassword string, passwordService *utils.PasswordService) bool {
	return passwordService.ComparePasswords(plainPassword, u.encryptedPassword)
}

// encodePassword encripta un password usando bcrypt
func encodePassword(plainPassword string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

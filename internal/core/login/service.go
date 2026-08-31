package login

import (
	"context"
	"fmt"

	"api_go/config"
	"api_go/internal/core/auth"
	"api_go/internal/infrastructure/jwt"
	"api_go/pkg/logger"
	"api_go/pkg/utils"
)

type LoginService struct {
	authPort        auth.AuthPort
	jwtService      *jwt.JWTService
	passwordService *utils.PasswordService
	config          *config.Config
}

func NewLoginService(
	authPort auth.AuthPort,
	jwtService *jwt.JWTService,
	passwordService *utils.PasswordService,
	config *config.Config,
) *LoginService {
	return &LoginService{
		config:          config,
		authPort:        authPort,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"passwd"`
}

// LoginResult es un objeto del dominio, NO una respuesta HTTP
type LoginResult struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// Execute retorna objetos del dominio, NO estructuras HTTP
func (s *LoginService) Execute(ctx context.Context, credentials LoginCredentials) (*LoginResult, error) {
	// 1. Validar credenciales con AuthPort
	user, err := s.authPort.RetrieveUser(ctx, auth.AuthData{
		Username: credentials.Email,
	})

	if err != nil {
		logger.Error("❌ Error en RetrieveUser", "email", credentials.Email, "error", err)
		return nil, fmt.Errorf("credenciales inválidas")
	}

	if user == nil {
		logger.Warn("❌ Usuario no encontrado", "email", credentials.Email)
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// 2. Verificar contraseña
	passwordMatch := s.passwordService.ComparePasswords(credentials.Password, user.PasswordHash)
	if !passwordMatch {
		logger.Warn("❌ Contraseña incorrecta", "email", credentials.Email)
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// 3. Generar token JWT
	token, err := s.jwtService.GenerateJWT(jwt.JwtPayload{
		IDUser:   user.ID,
		Username: user.Username,
		IDRol:    *user.IDRol,
	})
	if err != nil {
		logger.Error("❌ Error generando token JWT", "error", err)
		return nil, fmt.Errorf("error generando token: %v", err)
	}

	// ✅ ELIMINADO: Paso 4 (Guardar en caché de Redis).
	// La autenticación ahora es 100% stateless basada en la cookie firmada y el JWT.

	// 5. Construir resultado del dominio
	return &LoginResult{
		Token:     token,
		ExpiresIn: s.config.JWTExpireTime, // En segundos
	}, nil
}

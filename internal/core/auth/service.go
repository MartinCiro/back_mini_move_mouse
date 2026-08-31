package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"api_go/config"
	"api_go/internal/infrastructure/cookies"
	"api_go/internal/interfaces/api/common"
	"api_go/pkg/logger"
	"api_go/pkg/utils"
)

type AuthService struct {
	authPort        AuthPort
	cookieSigner    *cookies.CookieSigner
	passwordService *utils.PasswordService
	userRepo        UserRepositoryPort
	rolRepo         RolRepositoryPort
	estadoRepo      EstadoRepositoryPort
	config          *config.Config
}

func NewAuthService(
	authPort AuthPort,
	cookieSigner *cookies.CookieSigner,
	passwordService *utils.PasswordService,
	userRepo UserRepositoryPort,
	rolRepo RolRepositoryPort,
	estadoRepo EstadoRepositoryPort,
	config *config.Config,
) *AuthService {
	return &AuthService{
		authPort:        authPort,
		cookieSigner:    cookieSigner,
		passwordService: passwordService,
		userRepo:        userRepo,
		rolRepo:         rolRepo,
		estadoRepo:      estadoRepo,
		config:          config,
	}
}

// ✅ CAMBIADO: Login ahora usa Documento en lugar de Email
type LoginRequest struct {
	Documento string `json:"documento"`
	Password  string `json:"password"`
}

type ProfileResponse struct {
	ID       string   `json:"id"`
	Nombre   string   `json:"nombre"`
	Rol      string   `json:"rol"`
	Permisos []string `json:"permisos,omitempty"`
}

type AuthResponse struct {
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

func (s *AuthService) LoginUser(ctx context.Context, req LoginRequest) (*AuthResponse, string, time.Time, error) {
	if req.Documento == "" {
		return nil, "", time.Time{}, fmt.Errorf("documento o contraseña inválida")
	}

	// ✅ CAMBIADO: Buscar por Documento
	usuarioRetrieved, err := s.authPort.RetrieveUser(ctx, AuthData{Documento: req.Documento})
	if err != nil {
		logger.Error("❌ Error buscando usuario", "documento", req.Documento, "error", err)
		return nil, "", time.Time{}, fmt.Errorf("documento o contraseña inválida")
	}

	if usuarioRetrieved == nil || usuarioRetrieved.ID == "" {
		logger.Error("❌ Usuario no encontrado", "documento", req.Documento)
		return nil, "", time.Time{}, fmt.Errorf("documento o contraseña inválida")
	}

	usuario := NewUsuarioFromEncrypted(
		usuarioRetrieved.ID,
		usuarioRetrieved.Username,
		usuarioRetrieved.PasswordHash,
		usuarioRetrieved.IDRol,
		usuarioRetrieved.IDEstado,
	)

	isPasswordValid := usuario.ComparePassword(req.Password, s.passwordService)
	if !isPasswordValid {
		return nil, "", time.Time{}, fmt.Errorf("documento o contraseña inválida")
	}

	// ✅ OBTENER PERMISOS DIRECTO DE LA BD (Sin Redis)
	permisos, err := s.authPort.GetPermissionsByRoleID(ctx, *usuarioRetrieved.IDRol)
	if err != nil {
		logger.Error("❌ Error obteniendo permisos durante login",
			"userID", usuarioRetrieved.ID,
			"roleID", *usuarioRetrieved.IDRol,
			"error", err)
		return nil, "", time.Time{}, fmt.Errorf("error obteniendo permisos del usuario: %v", err)
	}

	if len(permisos) == 0 {
		logger.Warn("⚠️ Usuario sin permisos asignados durante login",
			"userID", usuarioRetrieved.ID,
			"roleID", *usuarioRetrieved.IDRol)
		return nil, "", time.Time{}, fmt.Errorf("usuario sin permisos asignados")
	}

	cookieData := &cookies.SignedCookieData{
		UserID:    usuarioRetrieved.ID,
		Username:  usuarioRetrieved.Username,
		Role:      usuarioRetrieved.RolNombre,
		RoleID:    *usuarioRetrieved.IDRol,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWTExpireTime) * time.Second),
		IssuedAt:  time.Now(),
	}

	signedCookie, err := s.cookieSigner.Sign(cookieData)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("error generando cookie de autenticación: %v", err)
	}

	response := &AuthResponse{
		Message:   "Login exitoso",
		ExpiresAt: cookieData.ExpiresAt,
	}

	return response, signedCookie, cookieData.ExpiresAt, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*common.ResponseBody[ProfileResponse], error) {
	// ✅ CAMBIADO: userID ya es string (Documento), no necesita conversión a int
	usuarioRetrieved, err := s.authPort.RetrieveUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	profileData := ProfileResponse{
		ID:       usuarioRetrieved.ID,
		Nombre:   usuarioRetrieved.Username,
		Rol:      usuarioRetrieved.RolNombre,
		Permisos: usuarioRetrieved.Permisos,
	}

	return &common.ResponseBody[ProfileResponse]{
		Success: true,
		Code:    200,
		Data:    profileData,
	}, nil
}

type RegisterRequest struct {
	Documento string `json:"documento" validate:"required"`
	Username  string `json:"username" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=6"`
	RolID     *int   `json:"rol_id,omitempty"`
	EstadoID  *int   `json:"estado_id,omitempty"`
}

func (s *AuthService) RegisterUser(ctx context.Context, req RegisterRequest, currentUser *User) (string, string, time.Time, error) {
	// 1. Verificar permisos si hay un usuario actual (ej. un admin creando otro usuario)
	if currentUser != nil {
		hasPermission := s.hasPermission(currentUser.Permisos, "usuario:crear")
		if !hasPermission {
			return "", "", time.Time{}, fmt.Errorf("no tiene permisos para crear usuarios")
		}
	}

	// 2. Valores por defecto si no se proporcionan
	rolID := 2 // Asumimos 2 = usuario normal
	if req.RolID != nil {
		rolID = *req.RolID
	}

	estadoID := 1 // Asumimos 1 = activo
	if req.EstadoID != nil {
		estadoID = *req.EstadoID
	}

	// 3. Crear entidad de dominio
	newUser, err := NewUsuario(req.Documento, req.Username, req.Password, &rolID, &estadoID)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("error procesando usuario: %v", err)
	}

	// 4. Guardar en base de datos
	documento, err := s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate") {
			return "", "", time.Time{}, fmt.Errorf("El usuario ya existe")
		}
		logger.Error("No se pudo crear el usuario", "error", err)
		return "", "", time.Time{}, fmt.Errorf("ha ocurrido un error en el servidor")
	}

	// 5. Generar cookie de sesión inmediata para el nuevo usuario
	cookieData := &cookies.SignedCookieData{
		UserID:    documento,
		Username:  req.Username,
		Role:      "usuario",
		RoleID:    rolID,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWTExpireTime) * time.Second),
		IssuedAt:  time.Now(),
	}

	signedCookie, err := s.cookieSigner.Sign(cookieData)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("error generando cookie: %v", err)
	}

	msg := "Usuario registrado con éxito"
	if currentUser != nil {
		msg = "Usuario creado correctamente por administrador"
	}

	return msg, signedCookie, cookieData.ExpiresAt, nil
}

func (s *AuthService) ValidateCookie(cookieValue string) (*User, error) {
	cookieData, err := s.cookieSigner.Verify(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("cookie inválida: %v", err)
	}

	ctx := context.Background()

	// ✅ OBTENER PERMISOS DIRECTO DE LA BD (Sin Redis)
	permisos, err := s.authPort.GetPermissionsByRoleID(ctx, cookieData.RoleID)
	if err != nil {
		logger.Error("❌ Error crítico obteniendo permisos para validación de cookie",
			"userID", cookieData.UserID,
			"roleID", cookieData.RoleID,
			"error", err)
		return nil, fmt.Errorf("error validando permisos del usuario: %v", err)
	}

	if len(permisos) == 0 {
		logger.Warn("⚠️ Usuario sin permisos asignados",
			"userID", cookieData.UserID,
			"roleID", cookieData.RoleID)
		return nil, fmt.Errorf("usuario sin permisos asignados")
	}

	return &User{
		ID:        cookieData.UserID,
		Username:  cookieData.Username,
		RolNombre: cookieData.Role,
		IDRol:     &cookieData.RoleID,
		Permisos:  permisos,
	}, nil
}

// ✅ CAMBIADO: userID ahora es string (Documento) para coincidir con el modelo
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// ✅ La invalidación de la sesión ahora es stateless.
	// El middleware/handler se encarga de borrar la cookie del cliente.
	return nil
}

func (s *AuthService) hasPermission(permisos []string, permission string) bool {
	for _, perm := range permisos {
		if perm == permission {
			return true
		}
	}
	return false
}

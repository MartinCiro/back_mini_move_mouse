package app

import (
	"api_go/config"
	"api_go/internal/core/auth"
	"api_go/internal/core/dwd"
	"api_go/internal/core/estados"
	"api_go/internal/core/login"
	"api_go/internal/core/permisos"
	"api_go/internal/core/roles"
	"api_go/internal/core/usuarios"
	"api_go/internal/infrastructure/adapters"
	"api_go/internal/infrastructure/cookies"
	"api_go/internal/infrastructure/database"
	"api_go/internal/infrastructure/database/models"
	"api_go/internal/infrastructure/jwt"
	"api_go/internal/infrastructure/repositories"
	"api_go/pkg/logger"
	"api_go/pkg/utils"
	"fmt"

	"gorm.io/gorm"
)

type App struct {
	// Configuración
	Config *config.Config

	// Infraestructura
	DB           *gorm.DB
	JWTService   *jwt.JWTService
	CookieSigner *cookies.CookieSigner

	// Servicios del Core
	AuthService    *auth.AuthService
	EstadoService  *estados.EstadoService
	PermisoService *permisos.PermisoService
	RolService     *roles.RolService
	UsuarioService *usuarios.UsuarioService

	// Servicios de Descargas
	DescargaService *dwd.DescargaService

	// Servicios de Utilidad
	PasswordService *utils.PasswordService
	LoginService    *login.LoginService
}

func NewApp() *App {
	return &App{}
}

func (a *App) Initialize(cfg *config.Config) error {
	a.Config = cfg

	// Inicializar infraestructura
	if err := a.initializeInfrastructure(); err != nil {
		return err
	}

	// Inicializar servicios
	if err := a.initializeServices(); err != nil {
		return err
	}

	// Ejecutar migraciones si es necesario
	if err := a.runMigrations(); err != nil {
		logger.Error("❌ Error en migraciones", "error", err)
		if cfg.IsDevelopment() {
			return fmt.Errorf("error en migraciones: %v", err)
		}
	}

	return nil
}

func (a *App) initializeInfrastructure() error {
	// Inicializar base de datos (SQLite)
	db, err := a.initializeDatabase()
	if err != nil {
		return fmt.Errorf("error inicializando base de datos: %v", err)
	}
	a.DB = db

	// Inicializar JWT Service
	a.JWTService = a.initializeJWTService()

	// Inicializar Cookie Signer
	a.CookieSigner = a.initializeCookieSigner()

	return nil
}

func (a *App) initializeDatabase() (*gorm.DB, error) {
	dbConfig := database.DBConfig{
		DBPath: a.Config.DBPath,
	}

	db, err := database.GetConnection(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("error inicializando base de datos: %v", err)
	}

	if err := database.HealthCheck(db); err != nil {
		return nil, fmt.Errorf("error en health check de BD: %v", err)
	}

	return db, nil
}

func (a *App) initializeJWTService() *jwt.JWTService {
	if a.Config.JWTSecret == "" {
		logger.Warn("JWTSecret está vacío en la configuración")
	} else {
		logger.Info("JWTSecret cargado correctamente", "length", len(a.Config.JWTSecret))
	}
	return jwt.NewJWTService(
		a.Config.JWTSecret,
		a.Config.Env,
	)
}

func (a *App) initializeCookieSigner() *cookies.CookieSigner {
	return cookies.NewCookieSigner(a.Config.CookieSecret, a.Config.IsProduction())
}

func (a *App) initializeServices() error {
	// Inicializar servicios de utilidad
	a.PasswordService = utils.NewPasswordService(a.Config.JWTSalt)

	// Inicializar DB Manager
	dbManager := database.NewDBManager(a.DB)

	// Inicializar repositorios
	usuarioRepo := repositories.NewUsuarioRepository(dbManager)
	rolRepo := repositories.NewRolRepository(dbManager)
	permisoRepo := repositories.NewPermisoRepository(dbManager)
	estadoRepo := repositories.NewEstadoRepository(dbManager)

	// ✅ INICIALIZAR ADAPTERS
	estadosAdapter := adapters.NewEstadosAdapter(a.DB)
	a.EstadoService = estados.NewEstadoService(estadosAdapter)

	permisosAdapter := adapters.NewPermisosAdapter(a.DB)
	a.PermisoService = permisos.NewPermisoService(permisosAdapter)

	rolesAdapter := adapters.NewRolesAdapter(a.DB)
	a.RolService = roles.NewRolService(rolesAdapter)

	usuariosAdapter := adapters.NewUsuariosAdapter(a.DB)
	a.UsuarioService = usuarios.NewUsuarioService(usuariosAdapter)

	descargasAdapter := adapters.NewDescargasAdapter("dist")
	a.DescargaService = dwd.NewDescargaService(descargasAdapter)

	// AuthAdapter coordina los repositorios
	authAdapter := repositories.NewAuthAdapter(usuarioRepo, rolRepo, permisoRepo)

	// Inicializar servicios del core
	a.AuthService = auth.NewAuthService(
		authAdapter,
		a.CookieSigner,
		a.PasswordService,
		usuarioRepo,
		rolRepo,
		estadoRepo,
		a.Config,
	)

	a.LoginService = login.NewLoginService(
		authAdapter,
		a.JWTService,
		a.PasswordService,
		a.Config,
	)

	logger.Info("✅ Todos los servicios del core inicializados correctamente",
		"AuthService", a.AuthService != nil,
		"EstadoService", a.EstadoService != nil,
		"RolService", a.RolService != nil,
		"LoginService", a.LoginService != nil)

	return nil
}

func (a *App) runMigrations() error {
	logger.Info("🔄 Iniciando migraciones...")

	models := []interface{}{
		&models.Migration{},
		&models.Estado{},
		&models.Rol{},
		&models.Permiso{},
		&models.RolXPermiso{},
		&models.Usuario{},
	}

	for _, model := range models {
		logger.Info("Migrando modelo", "model", fmt.Sprintf("%T", model))
		if err := a.DB.AutoMigrate(model); err != nil {
			logger.Error("Error migrando modelo", "model", fmt.Sprintf("%T", model), "error", err)
			return err
		}
	}

	return nil
}

func (a *App) Shutdown() {
	// Cerrar base de datos
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Warn("error cerrando conexión de BD", "error", err)
			}
		}
	}

	logger.Info("todos los servicios cerrados correctamente")
}

// HealthCheck verifica el estado de todos los servicios
func (a *App) HealthCheck() error {
	// Verificar base de datos
	if err := database.HealthCheck(a.DB); err != nil {
		return fmt.Errorf("base de datos no saludable: %v", err)
	}

	return nil
}

// GetConfig retorna la configuración (útil para tests)
func (a *App) GetConfig() *config.Config {
	return a.Config
}

// GetDB retorna la instancia de BD (útil para tests y operaciones directas)
func (a *App) GetDB() *gorm.DB {
	return a.DB
}

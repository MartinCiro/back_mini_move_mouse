package routes

import (
	"log"
	"net/http"
	"time"

	"api_go/config"
	"api_go/internal/app"
	"api_go/internal/interfaces/api/common"
	"api_go/internal/interfaces/api/handlers/auth"
	common_handler "api_go/internal/interfaces/api/handlers/common"
	"api_go/internal/interfaces/api/handlers/dwd"
	"api_go/internal/interfaces/api/handlers/estados"
	"api_go/internal/interfaces/api/handlers/permisos"
	"api_go/internal/interfaces/api/handlers/roles"
	"api_go/internal/interfaces/api/handlers/usuarios"
	"api_go/internal/interfaces/api/middlewares"
)

// responseWriter wrapper para interceptar status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SetupRouter configura todas las rutas de la aplicación
func SetupRouter(app *app.App, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Configurar todas las rutas
	setupAllRoutes(mux, app)

	// Aplicar middlewares globales
	return withGlobalMiddleware(mux, cfg)
}

// setupAllRoutes configura todas las rutas (públicas y protegidas)
func setupAllRoutes(mux *http.ServeMux, app *app.App) {

	// ========== RUTAS ACME para Let's Encrypt ==========
	mux.HandleFunc("GET /.well-known/acme-challenge/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		log.Printf("ACME challenge request: %s", r.URL.Path)
	})

	// Inicializar middlewares
	authMiddleware := middlewares.NewAuthMiddleware(app.JWTService)

	// Configurar CookieSigner si está disponible
	if app.CookieSigner != nil {
		authMiddleware.SetCookieSigner(app.CookieSigner)
	}

	// Configurar auth service en el middleware
	authMiddleware.SetAuthService(app.AuthService)

	// Inicializar handlers
	estadosHandler := estados.NewEstadosHandler(app.EstadoService)
	permisosHandler := permisos.NewPermisosHandler(app.PermisoService)
	authHandler := auth.NewAuthHandler(app.AuthService, app.CookieSigner)
	profileHandler := auth.NewProfileHandler(app.AuthService)
	rolesHandler := roles.NewRolesHandler(app.RolService)
	usuariosHandler := usuarios.NewUsuariosHandler(app.UsuarioService)
	descargasHandler := dwd.NewDescargasHandler(app.DescargaService)

	// ========== RUTAS PÚBLICAS ==========

	// Health checks
	mux.HandleFunc("GET /{$}", common_handler.HealthHandler)
	mux.HandleFunc("GET /health", common_handler.HealthHandler)
	// ✅ CORREGIDO: Sin dependencia de Redis en ReadyHandler
	mux.HandleFunc("GET /ready", common_handler.ReadyHandler(app.DB))

	// Autenticación (públicas)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.Handle("POST /api/auth/register", authMiddleware.OptionalAuth(http.HandlerFunc(authHandler.Register)))

	// ========== RUTAS PROTEGIDAS ==========

	// Crear un subrouter para rutas protegidas
	protected := http.NewServeMux()

	// ✅ RUTAS DE USUARIOS (RefreshMiddleware eliminado)
	protected.Handle("GET /api/usuarios",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionUsuariosListar)(
			http.HandlerFunc(usuariosHandler.ObtenerUsuarios),
		))

	protected.Handle("GET /api/usuarios/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionUsuariosListarXid)(
			http.HandlerFunc(usuariosHandler.ObtenerUsuarioXid),
		))

	protected.Handle("POST /api/usuarios",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionUsuariosCrear)(
			http.HandlerFunc(authHandler.Register),
		))

	protected.Handle("PATCH /api/usuarios/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionUsuariosEditar)(
			http.HandlerFunc(usuariosHandler.ActualizarUsuario),
		))

	// ✅ RUTAS DE ESTADOS
	protected.Handle("GET /api/estados",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionEstadosListar)(
			http.HandlerFunc(estadosHandler.ObtenerEstados),
		))

	protected.Handle("GET /api/estados/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionEstadosVer)(
			http.HandlerFunc(estadosHandler.ObtenerEstadoXid),
		))

	protected.Handle("POST /api/estados",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionEstadosCrear)(
			http.HandlerFunc(estadosHandler.CrearEstado),
		))

	protected.Handle("PATCH /api/estados/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionEstadosEditar)(
			http.HandlerFunc(estadosHandler.ActualizarEstado),
		))

	protected.Handle("DELETE /api/estados/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionEstadosEliminar)(
			http.HandlerFunc(estadosHandler.EliminarEstado),
		))

	// ✅ RUTAS DE PERMISOS
	protected.Handle("GET /api/permisos",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionPermisosListar)(
			http.HandlerFunc(permisosHandler.ObtenerPermisos),
		))

	protected.Handle("GET /api/permisos/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionPermisosVer)(
			http.HandlerFunc(permisosHandler.ObtenerPermisoXid),
		))

	protected.Handle("POST /api/permisos",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionPermisosCrear)(
			http.HandlerFunc(permisosHandler.CrearPermiso),
		))

	protected.Handle("PATCH /api/permisos/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionPermisosEditar)(
			http.HandlerFunc(permisosHandler.ActualizarPermiso),
		))

	protected.Handle("DELETE /api/permisos/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionPermisosEliminar)(
			http.HandlerFunc(permisosHandler.EliminarPermiso),
		))

	// ✅ RUTAS DE ROLES
	protected.Handle("GET /api/roles",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesListar)(
			http.HandlerFunc(rolesHandler.ObtenerRoles),
		))

	protected.Handle("GET /api/roles/permisos",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesPermisos)(
			http.HandlerFunc(rolesHandler.ObtenerPermisos),
		))

	protected.Handle("GET /api/roles/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesVer)(
			http.HandlerFunc(rolesHandler.ObtenerRolXid),
		))

	protected.Handle("POST /api/roles",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesCrear)(
			http.HandlerFunc(rolesHandler.CrearRol),
		))

	protected.Handle("PATCH /api/roles/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesEditar)(
			http.HandlerFunc(rolesHandler.ActualizarRol),
		))

	protected.Handle("DELETE /api/roles/{id}",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionRolesEliminar)(
			http.HandlerFunc(rolesHandler.EliminarRol),
		))

	// ✅ RUTAS DE AUTH PROTEGIDAS
	protected.Handle("POST /api/auth/logout",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionLoginLogout)(
			http.HandlerFunc(authHandler.Logout),
		))

	protected.Handle("GET /api/profile",
		authMiddleware.Handler(
			http.HandlerFunc(profileHandler.GetProfile),
		))

	protected.Handle("POST /api/descargas/archivo",
		authMiddleware.RequireAuthAndPermission(middlewares.PermissionDescargasEjecutar)(
			http.HandlerFunc(descargasHandler.DescargarArchivo),
		))

	// ✅ APLICAR RUTAS PROTEGIDAS (ya tienen auth middleware individualmente, no se necesita envolver todo el mux)
	mux.Handle("/", protected)
}

// withGlobalMiddleware aplica middlewares globales
func withGlobalMiddleware(handler http.Handler, cfg *config.Config) http.Handler {
	// CORS - importante para cookies
	handler = withCORS(handler)

	// Logging
	handler = withLogging(handler, cfg)

	// Recovery (manejo de panics)
	handler = withRecovery(handler)

	return handler
}

// withCORS middleware para CORS (actualizado para cookies)
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, jwt, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")    // IMPORTANTE para cookies
		w.Header().Set("Access-Control-Expose-Headers", "Set-Cookie") // Para que el frontend vea las cookies

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// withLogging middleware para logging
func withLogging(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		if cfg.IsDevelopment() {
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.statusCode, duration)
		}
	})
}

// withRecovery middleware para recuperación de panics
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("⚠️ Panic recovered: %v", err)
				response := common.NewErrorResponse(500, "Internal Server Error")
				common.WriteJSONResponse(w, response, 500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

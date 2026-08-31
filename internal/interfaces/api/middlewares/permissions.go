package middlewares

import (
	"net/http"
	"runtime/debug"
	"strings"

	"api_go/internal/core/auth"
	"api_go/internal/interfaces/api/common"
	"api_go/pkg/logger"
)

// RequirePermissions es un middleware funcional.
// Al no tener estado, no necesita ser un struct.
func RequirePermissions(requiredPermissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("⚠️ Panic en middleware de permisos",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
						"next_is_nil", next == nil,
						"stack", debug.Stack())
					response := common.NewErrorResponse(500, "Error interno del servidor")
					common.WriteJSONResponse(w, response, 500)
				}
			}()

			// Si no hay permisos requeridos, continuar
			if len(requiredPermissions) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Obtener usuario del contexto (seteado por AuthMiddleware)
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || user == nil {
				logger.Warn("intento de acceso sin usuario en contexto")
				common.WriteJSONResponse(w, common.NewErrorResponse(401, "No autenticado"), 401)
				return
			}

			// Verificar permisos del usuario
			if len(user.Permisos) == 0 {
				common.WriteJSONResponse(w, common.NewErrorResponse(403, "No tiene permisos asignados"), 403)
				return
			}

			// Verificar si el usuario tiene al menos uno de los permisos requeridos
			if !hasAnyPermission(user.Permisos, requiredPermissions) {
				common.WriteJSONResponse(w, common.NewErrorResponse(403, "No posee permisos suficientes para realizar esta acción"), 403)
				return
			}

			if next == nil {
				logger.Error("❌ CRÍTICO: Next handler es NIL",
					"path", r.URL.Path,
					"method", r.Method)
				common.WriteJSONResponse(w, common.NewErrorResponse(500, "Error de configuración: handler no disponible"), 500)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasAnyPermission verifica si el usuario tiene al menos uno de los permisos requeridos
func hasAnyPermission(userPermisos []string, requiredPermissions []string) bool {
	for _, requiredPerm := range requiredPermissions {
		for _, userPerm := range userPermisos {
			if strings.EqualFold(trimPermission(userPerm), trimPermission(requiredPerm)) {
				return true
			}
		}
	}
	return false
}

// trimPermission limpia espacios y convierte a minúsculas
func trimPermission(perm string) string {
	return strings.ToLower(strings.TrimSpace(perm))
}

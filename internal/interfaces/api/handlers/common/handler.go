package common

import (
	"net/http"
	"time"

	"api_go/internal/interfaces/api/common"

	"gorm.io/gorm" // ✅ Necesario para tipar db correctamente
)

// HealthHandler maneja el endpoint raíz y health check
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response := common.NewErrorResponse(405, "Método no permitido")
		common.WriteJSONResponse(w, response, 405)
		return
	}

	response := common.NewSuccessResponse(map[string]string{
		"message": "Hello world",
		"status":  "running",
		"version": "1.0.0",
	})
	common.WriteJSONResponse(w, response, 200)
}

// ReadyHandler verifica que la base de datos esté lista (Redis eliminado)
func ReadyHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Obtener la conexión SQL subyacente de GORM
		sqlDB, err := db.DB()
		if err != nil {
			response := common.NewErrorResponse(503, "Error obteniendo conexión de BD")
			common.WriteJSONResponse(w, response, 503)
			return
		}

		// 2. Hacer un Ping real a SQLite para asegurar que está operativa
		if err := sqlDB.Ping(); err != nil {
			response := common.NewErrorResponse(503, "Base de datos no disponible")
			common.WriteJSONResponse(w, response, 503)
			return
		}

		// 3. Si todo está bien, responder con éxito
		response := common.NewSuccessResponse(map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		common.WriteJSONResponse(w, response, 200)
	}
}

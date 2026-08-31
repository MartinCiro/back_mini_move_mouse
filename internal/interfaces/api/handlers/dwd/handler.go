package dwd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	coreDescargas "api_go/internal/core/dwd"
	"api_go/internal/interfaces/api/common"
	dwdDTOs "api_go/internal/interfaces/api/dwd"
	"api_go/pkg/logger"
)

type DescargasHandler struct {
	descargaService *coreDescargas.DescargaService
}

func NewDescargasHandler(service *coreDescargas.DescargaService) *DescargasHandler {
	return &DescargasHandler{descargaService: service}
}

// DescargarArchivo maneja la solicitud de descarga y apaga el servidor
func (h *DescargasHandler) DescargarArchivo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Decodificar request
	var req dwdDTOs.DescargarArchivoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteSimpleError(w, "Solicitud inválida: formato JSON incorrecto", 400)
		return
	}

	// Validar request
	if errors := common.ValidateRequest(req); errors != nil {
		common.WriteValidationErrors(w, errors, 400)
		return
	}

	// Ejecutar lógica de negocio
	archivoData := dwdDTOs.ToArchivoData(req)
	archivo, err := h.descargaService.EjecutarDescarga(ctx, archivoData)
	if err != nil {
		common.WriteSimpleError(w, err.Error(), 400)
		return
	}

	// Preparar headers para descarga
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(archivo.Nombre)+"\"")
	w.Header().Set("Content-Type", archivo.ContentType)
	w.Header().Set("Content-Length", string(archivo.Tamano))

	// Abrir y enviar archivo
	file, err := os.Open(archivo.Ruta)
	if err != nil {
		common.WriteSimpleError(w, "Error abriendo archivo para descarga", 500)
		return
	}
	defer file.Close()

	// Copiar contenido del archivo a la respuesta
	http.ServeContent(w, r, archivo.Nombre, archivo.FechaDescarga, file)

	logger.Info("✅ Archivo enviado correctamente",
		"archivo", archivo.Nombre,
		"tamano", archivo.Tamano)

	// ⚠️ APAGAR EL SERVIDOR DESPUÉS DE LA DESCARGA
	go func() {
		logger.Warn("🛑 Iniciando apagado del servidor después de descarga...")

		// Enviar señal SIGTERM al proceso actual
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			logger.Error("❌ Error obteniendo proceso actual", "error", err)
			return
		}

		// Enviar señal de terminación
		if err := p.Signal(syscall.SIGTERM); err != nil {
			logger.Error("❌ Error enviando señal de apagado", "error", err)
		}
	}()
}

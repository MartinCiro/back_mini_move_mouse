package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"api_go/internal/core/dwd"
	"api_go/pkg/logger"
)

type DescargasAdapter struct {
	baseDir string // Directorio base donde se buscan los archivos
}

func NewDescargasAdapter(baseDir string) *DescargasAdapter {
	return &DescargasAdapter{baseDir: baseDir}
}

// DescargarArchivo implementa el puerto DescargasPort
func (a *DescargasAdapter) DescargarArchivo(ctx context.Context, data dwd.ArchivoData) (*dwd.Archivo, error) {
	// Construir ruta completa
	rutaCompleta := filepath.Join(a.baseDir, data.RutaArchivo, data.NombreArchivo)

	// Verificar que el archivo existe
	fileInfo, err := os.Stat(rutaCompleta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archivo no encontrado: %s", data.NombreArchivo)
		}
		return nil, fmt.Errorf("error accediendo al archivo: %v", err)
	}

	// Determinar content type basado en extensión
	contentType := a.determinarContentType(data.NombreArchivo)

	return &dwd.Archivo{
		Nombre:        data.NombreArchivo,
		Ruta:          rutaCompleta,
		Tamano:        fileInfo.Size(),
		ContentType:   contentType,
		FechaDescarga: time.Now(),
	}, nil
}

// RegistrarDescarga registra la descarga (por ahora solo log)
func (a *DescargasAdapter) RegistrarDescarga(ctx context.Context, archivo *dwd.Archivo) error {
	logger.Info("📥 Descarga registrada",
		"archivo", archivo.Nombre,
		"tamano", archivo.Tamano,
		"fecha", archivo.FechaDescarga)
	return nil
}

// determinarContentType retorna el MIME type basado en la extensión
func (a *DescargasAdapter) determinarContentType(nombre string) string {
	ext := filepath.Ext(nombre)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

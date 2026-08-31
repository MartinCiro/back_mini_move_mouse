package dwd

import (
	"context"
	"fmt"
)

type DescargaService struct {
	descargasPort DescargasPort
}

func NewDescargaService(port DescargasPort) *DescargaService {
	return &DescargaService{descargasPort: port}
}

// EjecutarDescarga ejecuta la lógica de negocio para descargar un archivo
func (s *DescargaService) EjecutarDescarga(ctx context.Context, data ArchivoData) (*Archivo, error) {
	// Validar datos de entrada
	if data.NombreArchivo == "" || data.RutaArchivo == "" {
		return nil, fmt.Errorf("nombre y ruta del archivo son obligatorios")
	}

	// Obtener archivo del adapter
	archivo, err := s.descargasPort.DescargarArchivo(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("error preparando archivo para descarga: %v", err)
	}

	// Registrar la descarga
	if err := s.descargasPort.RegistrarDescarga(ctx, archivo); err != nil {
		// No fallar si el registro falla, solo loggear
		// En producción podrías manejar esto diferente
	}

	return archivo, nil
}

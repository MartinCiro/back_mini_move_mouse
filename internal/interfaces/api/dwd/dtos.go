package dwd

import "api_go/internal/core/dwd"

// DescargarArchivoRequest - DTO para solicitar descarga
type DescargarArchivoRequest struct {
	NombreArchivo string `json:"nombre_archivo" validate:"required"`
	RutaArchivo   string `json:"ruta_archivo" validate:"required"`
}

// DescargarArchivoResponse - DTO para respuesta
type DescargarArchivoResponse struct {
	Exitoso       bool   `json:"exitoso"`
	NombreArchivo string `json:"nombre_archivo,omitempty"`
	Mensaje       string `json:"mensaje"`
}

// ToArchivoData convierte DTO de request a entidad del core
func ToArchivoData(dto DescargarArchivoRequest) dwd.ArchivoData {
	return dwd.ArchivoData{
		NombreArchivo: dto.NombreArchivo,
		RutaArchivo:   dto.RutaArchivo,
	}
}

// FromDescargaResult convierte resultado del core a DTO de respuesta
func FromDescargaResult(result *dwd.DescargaResult) DescargarArchivoResponse {
	return DescargarArchivoResponse{
		Exitoso:       result.Exitoso,
		NombreArchivo: result.NombreArchivo,
		Mensaje:       result.Mensaje,
	}
}

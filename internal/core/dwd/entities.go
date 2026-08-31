package dwd

import "time"

// Archivo representa la entidad de dominio para descarga
type Archivo struct {
	Nombre        string    `json:"nombre"`
	Ruta          string    `json:"ruta"`
	Tamano        int64     `json:"tamano"`
	ContentType   string    `json:"content_type"`
	FechaDescarga time.Time `json:"fecha_descarga"`
}

// ArchivoData contiene los datos para solicitar una descarga
type ArchivoData struct {
	NombreArchivo string `json:"nombre_archivo"`
	RutaArchivo   string `json:"ruta_archivo"`
}

// DescargaResult representa el resultado de una descarga
type DescargaResult struct {
	Exitoso       bool   `json:"exitoso"`
	NombreArchivo string `json:"nombre_archivo"`
	TamanoBytes   int64  `json:"tamano_bytes"`
	Mensaje       string `json:"mensaje"`
}

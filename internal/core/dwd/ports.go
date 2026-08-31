package dwd

import "context"

// DescargasPort define el contrato para la gestión de descargas
type DescargasPort interface {
	// DescargarArchivo prepara y retorna los datos del archivo para descarga
	DescargarArchivo(ctx context.Context, data ArchivoData) (*Archivo, error)

	// RegistrarDescarga registra en log/base de datos que se realizó una descarga
	RegistrarDescarga(ctx context.Context, archivo *Archivo) error
}

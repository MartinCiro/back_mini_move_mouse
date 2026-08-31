package adapters

import (
	"context"
	"fmt"
	"strings"

	"api_go/internal/core/estados"
	"api_go/internal/infrastructure/database/models"
	"api_go/pkg/utils"

	"gorm.io/gorm"
)

type EstadosAdapter struct {
	db *gorm.DB
}

func NewEstadosAdapter(db *gorm.DB) *EstadosAdapter {
	return &EstadosAdapter{
		db: db,
	}
}

// CrearEstados implementa el puerto EstadosPort
func (a *EstadosAdapter) CrearEstados(ctx context.Context, estadoData estados.EstadoData) (*estados.Estado, error) {
	if a.db == nil {
		return nil, fmt.Errorf("error de configuración: conexión a base de datos no disponible")
	}

	estadoDB := models.Estado{
		NombreEstado: estadoData.Nombre,
		Descripcion:  estadoData.Descripcion,
	}

	err := a.db.WithContext(ctx).Create(&estadoDB).Error
	if err != nil {
		return nil, a.handleCreateError(err, estadoData.Nombre)
	}

	return a.toEstadoEntity(&estadoDB), nil
}

// ObtenerEstados implementa el puerto EstadosPort
func (a *EstadosAdapter) ObtenerEstados(ctx context.Context) ([]estados.Estado, error) {
	var estadosDB []models.Estado

	err := a.db.WithContext(ctx).Find(&estadosDB).Error
	if err != nil {
		return nil, a.handleQueryError(err, "consultando estados")
	}

	if len(estadosDB) == 0 {
		return []estados.Estado{}, nil
	}

	estadosList := make([]estados.Estado, len(estadosDB))
	for i, estadoDB := range estadosDB {
		estadosList[i] = *a.toEstadoEntity(&estadoDB)
	}

	return estadosList, nil
}

// ObtenerEstadosXid implementa el puerto EstadosPort
func (a *EstadosAdapter) ObtenerEstadosXid(ctx context.Context, estadoData estados.EstadoDataXid) (*estados.Estado, error) {
	var estadoDB models.Estado

	err := a.db.WithContext(ctx).
		Where("id = ?", estadoData.ID).
		First(&estadoDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Estado no encontrado
		}
		return nil, a.handleQueryError(err, "consultando estado")
	}

	return a.toEstadoEntity(&estadoDB), nil
}

// DelEstado implementa el puerto EstadosPort
func (a *EstadosAdapter) DelEstado(ctx context.Context, estadoData estados.EstadoDataXid) error {
	result := a.db.WithContext(ctx).Where("id = ?", estadoData.ID).Delete(&models.Estado{})

	if result.Error != nil {
		return a.handleDeleteError(result.Error, estadoData.ID)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("el estado con ID %d no existe", estadoData.ID)
	}

	return nil
}

// ActualizaEstado implementa el puerto EstadosPort
func (a *EstadosAdapter) ActualizaEstado(ctx context.Context, estadoData estados.EstadoDataUpdate) (*estados.Estado, error) {
	// Verificar si el estado existe
	var estadoExistente models.Estado
	err := a.db.WithContext(ctx).Where("id = ?", estadoData.ID).First(&estadoExistente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("el estado solicitado no existe en la base de datos")
		}
		return nil, a.handleQueryError(err, "verificando estado existente")
	}

	// Preparar updates
	updates := map[string]interface{}{
		"nombre_estado": estadoData.Nombre,
		"descripcion":   estadoData.Descripcion,
	}

	// Actualizar en base de datos
	err = a.db.WithContext(ctx).Model(&models.Estado{}).
		Where("id = ?", estadoData.ID).
		Updates(updates).Error

	if err != nil {
		return nil, a.handleUpdateError(err, estadoData.Nombre)
	}

	// Obtener estado actualizado para retornarlo
	var estadoActualizado models.Estado
	err = a.db.WithContext(ctx).Where("id = ?", estadoData.ID).First(&estadoActualizado).Error
	if err != nil {
		return nil, a.handleQueryError(err, "obteniendo estado actualizado")
	}

	return a.toEstadoEntity(&estadoActualizado), nil
}

// Mapeo de DB a Entity
func (a *EstadosAdapter) toEstadoEntity(estadoDB *models.Estado) *estados.Estado {
	return &estados.Estado{
		ID:          estadoDB.ID,
		Nombre:      estadoDB.NombreEstado,
		Descripcion: estadoDB.Descripcion,
	}
}

// Manejo de errores
func (a *EstadosAdapter) handleCreateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf(validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error creando el estado")
}

func (a *EstadosAdapter) handleQueryError(err error, operation string) error {
	return fmt.Errorf("Ocurrió un error %s", operation)
}

func (a *EstadosAdapter) handleDeleteError(err error, id int) error {
	errStr := err.Error()
	if strings.Contains(errStr, "foreign") || strings.Contains(errStr, "constraint") || strings.Contains(errStr, "FOREIGN KEY") {
		return fmt.Errorf("No se puede eliminar el estado porque tiene registros asociados")
	}
	return fmt.Errorf("Ocurrió un error eliminando el estado")
}

func (a *EstadosAdapter) handleUpdateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf("status_cod:409, data:%s", validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error actualizando el estado")
}

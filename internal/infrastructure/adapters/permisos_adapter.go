package adapters

import (
	"context"
	"fmt"
	"strings"

	"api_go/internal/core/permisos"
	"api_go/internal/infrastructure/database/models"
	"api_go/pkg/utils"

	"gorm.io/gorm"
)

type PermisosAdapter struct {
	db *gorm.DB
}

func NewPermisosAdapter(db *gorm.DB) *PermisosAdapter {
	return &PermisosAdapter{
		db: db,
	}
}

// CrearPermisos implementa el puerto PermisosPort
func (a *PermisosAdapter) CrearPermisos(ctx context.Context, permisoData permisos.PermisoData) (*permisos.Permiso, error) {
	if a.db == nil {
		return nil, fmt.Errorf("error de configuración: conexión a base de datos no disponible")
	}

	descripcion := ""
	if permisoData.Descripcion != nil {
		descripcion = *permisoData.Descripcion
	}

	permisoDB := models.Permiso{
		NombrePermiso: permisoData.Nombre,
		Descripcion:   descripcion,
	}

	err := a.db.WithContext(ctx).Create(&permisoDB).Error
	if err != nil {
		return nil, a.handleCreateError(err, permisoData.Nombre)
	}

	return a.toPermisoEntity(&permisoDB), nil
}

// ObtenerPermisos implementa el puerto PermisosPort
func (a *PermisosAdapter) ObtenerPermisos(ctx context.Context) ([]permisos.Permiso, error) {
	var permisosDB []models.Permiso

	err := a.db.WithContext(ctx).Find(&permisosDB).Error
	if err != nil {
		return nil, a.handleQueryError(err, "consultando permisos")
	}

	if len(permisosDB) == 0 {
		return []permisos.Permiso{}, nil
	}

	permisosList := make([]permisos.Permiso, len(permisosDB))
	for i, permisoDB := range permisosDB {
		permisosList[i] = *a.toPermisoEntity(&permisoDB)
	}

	return permisosList, nil
}

// ObtenerPermisosXid implementa el puerto PermisosPort
func (a *PermisosAdapter) ObtenerPermisosXid(ctx context.Context, permisoData permisos.PermisoDataXid) (*permisos.Permiso, error) {
	var permisoDB models.Permiso

	err := a.db.WithContext(ctx).
		Where("id = ?", permisoData.ID).
		First(&permisoDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Permiso no encontrado
		}
		return nil, a.handleQueryError(err, "consultando permiso")
	}

	return a.toPermisoEntity(&permisoDB), nil
}

// DelPermiso implementa el puerto PermisosPort
func (a *PermisosAdapter) DelPermiso(ctx context.Context, permisoData permisos.PermisoDataXid) error {
	result := a.db.WithContext(ctx).Where("id = ?", permisoData.ID).Delete(&models.Permiso{})

	if result.Error != nil {
		return a.handleDeleteError(result.Error, permisoData.ID)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("el permiso con ID %d no existe", permisoData.ID)
	}

	return nil
}

// ActualizaPermiso implementa el puerto PermisosPort
func (a *PermisosAdapter) ActualizaPermiso(ctx context.Context, permisoData permisos.PermisoDataUpdate) (*permisos.Permiso, error) {
	var permisoExistente models.Permiso
	err := a.db.WithContext(ctx).Where("id = ?", permisoData.ID).First(&permisoExistente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("el permiso solicitado no existe en la base de datos")
		}
		return nil, a.handleQueryError(err, "verificando permiso existente")
	}

	updates := map[string]interface{}{
		"nombre_permiso": permisoData.Nombre,
		"descripcion":    permisoData.Descripcion,
	}

	err = a.db.WithContext(ctx).Model(&models.Permiso{}).
		Where("id = ?", permisoData.ID).
		Updates(updates).Error

	if err != nil {
		return nil, a.handleUpdateError(err, permisoData.Nombre)
	}

	var permisoActualizado models.Permiso
	err = a.db.WithContext(ctx).Where("id = ?", permisoData.ID).First(&permisoActualizado).Error
	if err != nil {
		return nil, a.handleQueryError(err, "obteniendo permiso actualizado")
	}

	return a.toPermisoEntity(&permisoActualizado), nil
}

// Mapeo de DB a Entity
func (a *PermisosAdapter) toPermisoEntity(permisoDB *models.Permiso) *permisos.Permiso {
	return &permisos.Permiso{
		ID: permisoDB.ID,
		PermisoBase: permisos.PermisoBase{
			Nombre:      permisoDB.NombrePermiso,
			Descripcion: &permisoDB.Descripcion,
		},
	}
}

// Manejo de errores
func (a *PermisosAdapter) handleCreateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf(validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error creando el permiso")
}

func (a *PermisosAdapter) handleQueryError(err error, operation string) error {
	return fmt.Errorf("Ocurrió un error %s", operation)
}

func (a *PermisosAdapter) handleDeleteError(err error, id int) error {
	errStr := err.Error()
	if strings.Contains(errStr, "foreign") || strings.Contains(errStr, "constraint") || strings.Contains(errStr, "FOREIGN KEY") {
		return fmt.Errorf("No se puede eliminar el permiso porque tiene registros asociados")
	}
	return fmt.Errorf("Ocurrió un error eliminando el permiso")
}

func (a *PermisosAdapter) handleUpdateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf("status_cod:409, data:%s", validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error actualizando el permiso")
}

package adapters

import (
	"context"
	"fmt"
	"strings"

	"api_go/internal/core/roles"
	"api_go/internal/infrastructure/database/models"
	"api_go/pkg/utils"

	"gorm.io/gorm"
)

type RolesAdapter struct {
	db *gorm.DB
}

func NewRolesAdapter(db *gorm.DB) *RolesAdapter {
	return &RolesAdapter{
		db: db,
	}
}

// CrearRol implementa el puerto RolesPort
func (a *RolesAdapter) CrearRol(ctx context.Context, rolData roles.RolData) (*roles.Rol, error) {
	// Verificar si el rol ya existe
	var rolExistente models.Rol
	err := a.db.WithContext(ctx).Where("nombre_rol = ?", rolData.Nombre).First(&rolExistente).Error
	if err == nil {
		return nil, fmt.Errorf("status_cod:409, data:Ya existe un rol con el nombre '%s'", rolData.Nombre)
	} else if err != gorm.ErrRecordNotFound {
		return nil, a.handleQueryError(err, "verificando rol existente")
	}

	// ✅ CORREGIDO: Validar que los IDs de permisos proporcionados existan en la BD
	validPermisosIDs, err := a.permisosIDs(ctx, rolData.Permisos)
	if err != nil {
		return nil, err
	}

	rolDB := models.Rol{
		NombreRol:   rolData.Nombre,
		Descripcion: rolData.Descripcion,
	}

	if err := a.db.WithContext(ctx).Create(&rolDB).Error; err != nil {
		return nil, a.handleCreateError(err, rolData.Nombre)
	}

	// Transacción para asignar permisos
	tx := a.db.WithContext(ctx).Begin()
	for _, permisoID := range validPermisosIDs {
		rolPermiso := models.RolXPermiso{
			IDRol:     rolDB.ID,
			IDPermiso: permisoID,
		}
		if err := tx.Create(&rolPermiso).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error asignando permisos")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error al guardar los permisos del rol")
	}

	// Obtener el rol creado con sus permisos
	return a.obtenerRolConPermisos(ctx, rolDB.ID)
}

// ObtenerRoles implementa el puerto RolesPort
func (a *RolesAdapter) ObtenerRoles(ctx context.Context) ([]roles.Rol, error) {
	var rolesDB []models.Rol
	err := a.db.WithContext(ctx).Find(&rolesDB).Error
	if err != nil {
		return nil, a.handleQueryError(err, "consultando roles")
	}

	if len(rolesDB) == 0 {
		return []roles.Rol{}, nil
	}

	rolesList := make([]roles.Rol, len(rolesDB))
	for i, rolDB := range rolesDB {
		rolCompleto, err := a.obtenerRolConPermisos(ctx, rolDB.ID)
		if err != nil {
			return nil, err
		}
		rolesList[i] = *rolCompleto
	}

	return rolesList, nil
}

// ObtenerRolXid implementa el puerto RolesPort
func (a *RolesAdapter) ObtenerRolXid(ctx context.Context, rolData roles.RolDataXid) (*roles.Rol, error) {
	var rolDB models.Rol
	err := a.db.WithContext(ctx).Where("id = ?", rolData.ID).First(&rolDB).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, a.handleQueryError(err, "consultando rol")
	}

	return a.obtenerRolConPermisos(ctx, rolDB.ID)
}

// ActualizarRol implementa el puerto RolesPort
func (a *RolesAdapter) ActualizarRol(ctx context.Context, rolData roles.RolDataUpdate) (*roles.Rol, error) {
	var rolExistente models.Rol
	err := a.db.WithContext(ctx).Where("id = ?", rolData.ID).First(&rolExistente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("status_cod:404, data:El rol solicitado no existe")
		}
		return nil, a.handleQueryError(err, "verificando rol existente")
	}

	if rolData.Nombre != nil {
		var rolConMismoNombre models.Rol
		err = a.db.WithContext(ctx).Where("nombre_rol = ? AND id != ?", *rolData.Nombre, rolData.ID).First(&rolConMismoNombre).Error
		if err == nil {
			return nil, fmt.Errorf("status_cod:409, data:Ya existe otro rol con el nombre '%s'", *rolData.Nombre)
		} else if err != gorm.ErrRecordNotFound {
			return nil, a.handleQueryError(err, "verificando nombre duplicado")
		}
	}

	tx := a.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updates := make(map[string]interface{})
	if rolData.Nombre != nil {
		updates["nombre_rol"] = *rolData.Nombre
	}
	if rolData.Descripcion != nil {
		updates["descripcion"] = *rolData.Descripcion
	}

	if len(updates) > 0 {
		if err := tx.Model(&models.Rol{}).Where("id = ?", rolData.ID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, a.handleUpdateError(err, "actualizando rol")
		}
	}

	// ✅ Actualizar permisos solo si se proporcionaron (rolData.Permisos es []int)
	if rolData.Permisos != nil {
		// Eliminar permisos actuales
		if err := tx.Where("id_rol = ?", rolData.ID).Delete(&models.RolXPermiso{}).Error; err != nil {
			tx.Rollback()
			return nil, a.handleUpdateError(err, "eliminando permisos anteriores")
		}

		// ✅ CORREGIDO: Validar que los nuevos IDs de permisos existan
		validPermisosIDs, err := a.permisosIDs(ctx, rolData.Permisos)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		for _, permisoID := range validPermisosIDs {
			rolPermiso := models.RolXPermiso{
				IDRol:     rolData.ID,
				IDPermiso: permisoID,
			}
			if err := tx.Create(&rolPermiso).Error; err != nil {
				tx.Rollback()
				return nil, a.handleUpdateError(err, "asignando nuevos permisos")
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error al actualizar el rol: %v", err)
	}

	return a.obtenerRolConPermisos(ctx, rolData.ID)
}

// EliminarRol implementa el puerto RolesPort
func (a *RolesAdapter) EliminarRol(ctx context.Context, rolData roles.RolDataXid) error {
	var rolExistente models.Rol
	err := a.db.WithContext(ctx).Where("id = ?", rolData.ID).First(&rolExistente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("status_cod:404, data:El rol solicitado no existe")
		}
		return a.handleQueryError(err, "verificando rol existente")
	}

	var countUsuarios int64
	err = a.db.WithContext(ctx).Model(&models.Usuario{}).Where("id_rol = ?", rolData.ID).Count(&countUsuarios).Error
	if err != nil {
		return a.handleDeleteError(err, rolData.ID)
	}

	if countUsuarios > 0 {
		return fmt.Errorf("no se puede eliminar el rol porque tiene usuarios asociados")
	}

	tx := a.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("id_rol = ?", rolData.ID).Delete(&models.RolXPermiso{}).Error; err != nil {
		tx.Rollback()
		return a.handleDeleteError(err, rolData.ID)
	}

	result := tx.Where("id = ?", rolData.ID).Delete(&models.Rol{})
	if result.Error != nil {
		tx.Rollback()
		return a.handleDeleteError(result.Error, rolData.ID)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("status_cod:404, data:El rol con ID %d no existe", rolData.ID)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error al eliminar el rol: %v", err)
	}

	return nil
}

// ObtenerTodosLosPermisos implementa el puerto RolesPort
func (a *RolesAdapter) ObtenerTodosLosPermisos(ctx context.Context) ([]string, error) {
	var permisosDB []models.Permiso
	err := a.db.WithContext(ctx).Find(&permisosDB).Error
	if err != nil {
		return nil, a.handleQueryError(err, "consultando permisos")
	}

	permisos := make([]string, len(permisosDB))
	for i, permisoDB := range permisosDB {
		permisos[i] = permisoDB.NombrePermiso
	}

	return permisos, nil
}

// obtenerRolConPermisos obtiene un rol con sus permisos
func (a *RolesAdapter) obtenerRolConPermisos(ctx context.Context, rolID int) (*roles.Rol, error) {
	var rolDB models.Rol
	err := a.db.WithContext(ctx).Where("id = ?", rolID).First(&rolDB).Error
	if err != nil {
		return nil, a.handleQueryError(err, "obteniendo rol")
	}

	var permisosDB []models.Permiso
	err = a.db.WithContext(ctx).
		Table("permisos p").
		Select("p.nombre_permiso").
		Joins("INNER JOIN rol_x_permisos rxp ON p.id = rxp.id_permiso").
		Where("rxp.id_rol = ?", rolID).
		Find(&permisosDB).Error

	if err != nil {
		return nil, a.handleQueryError(err, "obteniendo permisos del rol")
	}

	permisos := make([]string, len(permisosDB))
	for i, permisoDB := range permisosDB {
		permisos[i] = permisoDB.NombrePermiso
	}

	return &roles.Rol{
		ID:          rolDB.ID,
		Nombre:      rolDB.NombreRol,
		Descripcion: rolDB.Descripcion,
		Permisos:    permisos, // Nota: Asumiendo que Rol.Permisos en el core es []string con los nombres
	}, nil
}

// Manejo de errores
func (a *RolesAdapter) handleCreateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf(validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error creando el rol")
}

func (a *RolesAdapter) handleQueryError(err error, operation string) error {
	return fmt.Errorf("Ocurrió un error %s", operation)
}

func (a *RolesAdapter) handleUpdateError(err error, nombre string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "Duplicate") || strings.Contains(errStr, "UNIQUE constraint failed") {
		validacion := utils.ValidarExistente("P2002", nombre)
		if !validacion.OK {
			return fmt.Errorf("status_cod:409, data:%s", validacion.Data)
		}
	}
	return fmt.Errorf("Ocurrió un error actualizando el rol")
}

func (a *RolesAdapter) handleDeleteError(err error, id int) error {
	errStr := err.Error()
	if strings.Contains(errStr, "foreign") || strings.Contains(errStr, "constraint") || strings.Contains(errStr, "FOREIGN KEY") {
		return fmt.Errorf("No se puede eliminar el rol porque tiene registros asociados")
	}
	return fmt.Errorf("Ocurrió un error eliminando el rol")
}

// permisosIDs valida que los IDs de permisos proporcionados existan en la base de datos
func (a *RolesAdapter) permisosIDs(ctx context.Context, permisosIDs []int) ([]int, error) {
	if len(permisosIDs) == 0 {
		return []int{}, nil // Permitir rol sin permisos si así se desea, o cambiar a error si es obligatorio
	}

	var count int64
	err := a.db.WithContext(ctx).Model(&models.Permiso{}).
		Where("id IN ?", permisosIDs).
		Count(&count).Error

	if err != nil {
		return nil, a.handleQueryError(err, "validando permisos")
	}

	if int(count) != len(permisosIDs) {
		return nil, fmt.Errorf("uno o más permisos no existen")
	}

	return permisosIDs, nil
}

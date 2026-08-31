package adapters

import (
	"context"
	"fmt"
	"strings"

	"api_go/internal/core/usuarios"
	"api_go/internal/infrastructure/database/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UsuariosAdapter struct {
	db *gorm.DB
}

func NewUsuariosAdapter(db *gorm.DB) *UsuariosAdapter {
	return &UsuariosAdapter{db: db}
}

func (a *UsuariosAdapter) toUsuarioEntityConRelaciones(usuarioDB *models.Usuario) *usuarios.Usuario {
	return &usuarios.Usuario{
		Documento:     usuarioDB.Documento,
		Username:      usuarioDB.NomUser,
		RolID:         usuarioDB.IDRol,
		EstadoID:      usuarioDB.EstadoID,
		RolNombre:     usuarioDB.Rol.NombreRol,
		EstadoNombre:  usuarioDB.Estado.NombreEstado,
		FechaRegistro: usuarioDB.FechaRegistro,
	}
}

func (a *UsuariosAdapter) toUsuarioConRelacionesEntity(usuarioDB *models.Usuario) *usuarios.UsuarioConRelaciones {
	return &usuarios.UsuarioConRelaciones{
		Documento:     usuarioDB.Documento,
		Username:      usuarioDB.NomUser,
		RolID:         usuarioDB.IDRol,
		EstadoID:      usuarioDB.EstadoID,
		RolNombre:     usuarioDB.Rol.NombreRol,
		EstadoNombre:  usuarioDB.Estado.NombreEstado,
		FechaRegistro: usuarioDB.FechaRegistro,
	}
}

// ObtenerUsuarios implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ObtenerUsuarios(ctx context.Context) ([]usuarios.Usuario, error) {
	var usuariosDB []models.Usuario
	err := a.db.WithContext(ctx).Preload("Rol").Preload("Estado").Find(&usuariosDB).Error
	if err != nil {
		return nil, fmt.Errorf("error consultando usuarios: %v", err)
	}

	if len(usuariosDB) == 0 {
		return []usuarios.Usuario{}, nil
	}

	usuariosList := make([]usuarios.Usuario, len(usuariosDB))
	for i, usuarioDB := range usuariosDB {
		usuariosList[i] = *a.toUsuarioEntityConRelaciones(&usuarioDB)
	}
	return usuariosList, nil
}

// ObtenerUsuarioXid implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ObtenerUsuarioXid(ctx context.Context, usuarioData usuarios.UsuarioDataXid) (*usuarios.Usuario, error) {
	var usuarioDB models.Usuario
	err := a.db.WithContext(ctx).Preload("Rol").Preload("Estado").Where("documento = ?", usuarioData.Documento).First(&usuarioDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error consultando usuario: %v", err)
	}
	return a.toUsuarioEntityConRelaciones(&usuarioDB), nil
}

// ActualizarUsuario implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ActualizarUsuario(ctx context.Context, usuarioData usuarios.UsuarioDataUpdate) (*usuarios.Usuario, error) {
	var usuarioExistente models.Usuario
	err := a.db.WithContext(ctx).Where("documento = ?", usuarioData.Documento).First(&usuarioExistente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("el usuario solicitado no existe")
		}
		return nil, fmt.Errorf("error verificando usuario existente: %v", err)
	}

	updates := make(map[string]interface{})
	if usuarioData.Username != nil {
		updates["nom_user"] = *usuarioData.Username
	}
	if usuarioData.RolID != nil {
		updates["id_rol"] = *usuarioData.RolID
	}
	if usuarioData.EstadoID != nil {
		updates["estado_id"] = *usuarioData.EstadoID
	}

	if len(updates) == 0 {
		return a.toUsuarioEntityConRelaciones(&usuarioExistente), nil
	}

	err = a.db.WithContext(ctx).Model(&models.Usuario{}).Where("documento = ?", usuarioData.Documento).Updates(updates).Error
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && usuarioData.Username != nil {
			return nil, fmt.Errorf("status_cod:409, data:El nombre de usuario ya existe")
		}
		return nil, fmt.Errorf("error actualizando usuario: %v", err)
	}

	var usuarioActualizado models.Usuario
	err = a.db.WithContext(ctx).Preload("Rol").Preload("Estado").Where("documento = ?", usuarioData.Documento).First(&usuarioActualizado).Error
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuario actualizado: %v", err)
	}

	return a.toUsuarioEntityConRelaciones(&usuarioActualizado), nil
}

// EliminarUsuario implementa el puerto UsuariosPort
func (a *UsuariosAdapter) EliminarUsuario(ctx context.Context, usuarioData usuarios.UsuarioDataXid) error {
	result := a.db.WithContext(ctx).Where("documento = ?", usuarioData.Documento).Delete(&models.Usuario{})
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "FOREIGN KEY") || strings.Contains(result.Error.Error(), "constraint") {
			return fmt.Errorf("no se puede eliminar el usuario porque tiene registros asociados")
		}
		return fmt.Errorf("error eliminando usuario: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("el usuario con documento %s no existe", usuarioData.Documento)
	}
	return nil
}

// CambiarPassword implementa el puerto UsuariosPort
func (a *UsuariosAdapter) CambiarPassword(ctx context.Context, cambiarPasswordData usuarios.CambiarPasswordData) error {
	var usuario models.Usuario
	err := a.db.WithContext(ctx).Where("documento = ?", cambiarPasswordData.Documento).First(&usuario).Error
	if err != nil {
		return fmt.Errorf("usuario no encontrado")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Pass), []byte(cambiarPasswordData.PasswordActual)); err != nil {
		return fmt.Errorf("la contraseña actual es incorrecta")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cambiarPasswordData.NuevoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error procesando nueva contraseña")
	}

	err = a.db.WithContext(ctx).Model(&models.Usuario{}).Where("documento = ?", cambiarPasswordData.Documento).Update("pass", string(hashedPassword)).Error
	if err != nil {
		return fmt.Errorf("error actualizando contraseña: %v", err)
	}
	return nil
}

// ObtenerUsuarioPorUsername implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ObtenerUsuarioPorUsername(ctx context.Context, username string) (*usuarios.Usuario, error) {
	var usuarioDB models.Usuario
	err := a.db.WithContext(ctx).Where("nom_user = ?", username).First(&usuarioDB).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error consultando usuario por username: %v", err)
	}
	return a.toUsuarioEntityConRelaciones(&usuarioDB), nil
}

// ObtenerUsuariosConRelaciones implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ObtenerUsuariosConRelaciones(ctx context.Context) ([]usuarios.UsuarioConRelaciones, error) {
	var usuariosDB []models.Usuario
	err := a.db.WithContext(ctx).Preload("Rol").Preload("Estado").Find(&usuariosDB).Error
	if err != nil {
		return nil, fmt.Errorf("error consultando usuarios con relaciones: %v", err)
	}

	usuariosList := make([]usuarios.UsuarioConRelaciones, len(usuariosDB))
	for i, usuarioDB := range usuariosDB {
		usuariosList[i] = *a.toUsuarioConRelacionesEntity(&usuarioDB)
	}
	return usuariosList, nil
}

// ObtenerUsuarioConRelaciones implementa el puerto UsuariosPort
func (a *UsuariosAdapter) ObtenerUsuarioConRelaciones(ctx context.Context, usuarioData usuarios.UsuarioDataXid) (*usuarios.UsuarioConRelaciones, error) {
	var usuarioDB models.Usuario
	err := a.db.WithContext(ctx).Preload("Rol").Preload("Estado").Where("documento = ?", usuarioData.Documento).First(&usuarioDB).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error consultando usuario con relaciones: %v", err)
	}
	return a.toUsuarioConRelacionesEntity(&usuarioDB), nil
}

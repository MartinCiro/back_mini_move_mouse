package repositories

import (
	"context"
	"fmt"

	"api_go/internal/core/auth"
	"api_go/pkg/logger"
)

type AuthAdapter struct {
	usuarioRepo *UsuarioRepository
	rolRepo     *RolRepository
	permisoRepo *PermisoRepository
}

func NewAuthAdapter(
	usuarioRepo *UsuarioRepository,
	rolRepo *RolRepository,
	permisoRepo *PermisoRepository,
) *AuthAdapter {
	return &AuthAdapter{
		usuarioRepo: usuarioRepo,
		rolRepo:     rolRepo,
		permisoRepo: permisoRepo,
	}
}

// RetrieveUser busca un usuario por Documento o Username
func (a *AuthAdapter) RetrieveUser(ctx context.Context, authData auth.AuthData) (*auth.User, error) {
	var usuario *auth.User
	var err error

	// ✅ CORREGIDO: Ahora buscamos por Documento o Username
	if authData.Documento != "" {
		// FindByID ahora espera un string (el documento)
		usuario, err = a.usuarioRepo.FindByID(ctx, authData.Documento)
	} else if authData.Username != "" {
		usuario, err = a.usuarioRepo.FindByUsername(ctx, authData.Username)
	} else {
		return nil, fmt.Errorf("debe proporcionar documento o username")
	}

	if err != nil {
		return nil, fmt.Errorf("error recuperando usuario: %v", err)
	}

	if usuario == nil {
		return nil, nil // Usuario no encontrado, no es un error crítico
	}

	// Obtener permisos y nombre del rol si tiene un rol asignado
	if usuario.IDRol != nil {
		permisos, err := a.permisoRepo.FindByRolID(ctx, *usuario.IDRol)
		if err != nil {
			logger.Error("⚠️ Error obteniendo permisos (continuando sin permisos)", "error", err)
			usuario.Permisos = []string{}
		} else {
			usuario.Permisos = permisos
		}

		rol, err := a.rolRepo.FindByID(ctx, *usuario.IDRol)
		if err != nil {
			logger.Warn("⚠️ Error obteniendo rol (continuando)", "error", err)
		} else if rol != nil {
			usuario.RolNombre = rol.NombreRol
		}
	}

	return usuario, nil
}

// RetrieveUserByID busca un usuario específicamente por su Documento (que actúa como ID)
func (a *AuthAdapter) RetrieveUserByID(ctx context.Context, documento string) (*auth.User, error) {
	// ✅ CORREGIDO: El parámetro ahora es string (documento), no int
	usuario, err := a.usuarioRepo.FindByID(ctx, documento)
	if err != nil {
		return nil, err
	}

	if usuario == nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	if usuario.IDRol != nil {
		permisos, err := a.permisoRepo.FindByRolID(ctx, *usuario.IDRol)
		if err != nil {
			logger.Warn("⚠️ Error obteniendo permisos (continuando sin permisos)", "error", err)
			usuario.Permisos = []string{}
		} else {
			usuario.Permisos = permisos
		}

		rol, err := a.rolRepo.FindByID(ctx, *usuario.IDRol)
		if err != nil {
			logger.Warn("⚠️ Error obteniendo rol (continuando)", "error", err)
		} else if rol != nil {
			usuario.RolNombre = rol.NombreRol
		}
	}

	return usuario, nil
}

// GetPermissionsByRoleID obtiene los nombres de los permisos asociados a un rol
func (a *AuthAdapter) GetPermissionsByRoleID(ctx context.Context, roleID int) ([]string, error) {
	return a.permisoRepo.FindByRolID(ctx, roleID)
}

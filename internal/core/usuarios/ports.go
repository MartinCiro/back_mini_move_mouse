// internal/core/usuarios/ports.go
package usuarios

import "context"

// UsuariosPort define el contrato para la gestión de usuarios
type UsuariosPort interface {
	// ObtenerUsuarios obtiene todos los usuarios
	ObtenerUsuarios(ctx context.Context) ([]Usuario, error)

	// ObtenerUsuarioXid obtiene un usuario por su ID
	ObtenerUsuarioXid(ctx context.Context, usuarioData UsuarioDataXid) (*Usuario, error)

	// ActualizarUsuario actualiza un usuario existente
	ActualizarUsuario(ctx context.Context, usuarioData UsuarioDataUpdate) (*Usuario, error)

	// EliminarUsuario elimina un usuario por su ID
	EliminarUsuario(ctx context.Context, usuarioData UsuarioDataXid) error

	// CambiarPassword cambia la contraseña de un usuario
	CambiarPassword(ctx context.Context, cambiarPasswordData CambiarPasswordData) error

	// ObtenerUsuarioPorUsername obtiene un usuario por su nombre de usuario
	ObtenerUsuarioPorUsername(ctx context.Context, username string) (*Usuario, error)
	// ObtenerUsuarioConRelaciones obtiene un usuario con información de rol y estado
	ObtenerUsuarioConRelaciones(ctx context.Context, usuarioData UsuarioDataXid) (*UsuarioConRelaciones, error)

	// ObtenerUsuariosConRelaciones obtiene todos los usuarios con información de rol y estado
	ObtenerUsuariosConRelaciones(ctx context.Context) ([]UsuarioConRelaciones, error)
}

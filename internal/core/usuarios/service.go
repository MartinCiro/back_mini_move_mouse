// internal/core/usuarios/service.go
package usuarios

import (
	"context"
)

type UsuarioService struct {
	usuarioPort UsuariosPort
}

func NewUsuarioService(usuarioPort UsuariosPort) *UsuarioService {
	return &UsuarioService{
		usuarioPort: usuarioPort,
	}
}

func (s *UsuarioService) ObtenerUsuarios(ctx context.Context) ([]Usuario, error) {
	return s.usuarioPort.ObtenerUsuarios(ctx)
}

func (s *UsuarioService) ObtenerUsuarioXid(ctx context.Context, usuarioData UsuarioDataXid) (*Usuario, error) {
	return s.usuarioPort.ObtenerUsuarioXid(ctx, usuarioData)
}

func (s *UsuarioService) UpUsuario(ctx context.Context, usuarioData UsuarioDataUpdate) (*Usuario, error) {
	return s.usuarioPort.ActualizarUsuario(ctx, usuarioData)
}

func (s *UsuarioService) DelUsuario(ctx context.Context, usuarioData UsuarioDataXid) error {
	return s.usuarioPort.EliminarUsuario(ctx, usuarioData)
}

func (s *UsuarioService) CambiarPassword(ctx context.Context, cambiarPasswordData CambiarPasswordData) error {
	return s.usuarioPort.CambiarPassword(ctx, cambiarPasswordData)
}

func (s *UsuarioService) ObtenerUsuariosConRelaciones(ctx context.Context) ([]UsuarioConRelaciones, error) {
	return s.usuarioPort.ObtenerUsuariosConRelaciones(ctx)
}

func (s *UsuarioService) ObtenerUsuarioConRelaciones(ctx context.Context, usuarioData UsuarioDataXid) (*UsuarioConRelaciones, error) {
	return s.usuarioPort.ObtenerUsuarioConRelaciones(ctx, usuarioData)
}

func (s *UsuarioService) ObtenerUsuarioPorUsername(ctx context.Context, username string) (*Usuario, error) {
	return s.usuarioPort.ObtenerUsuarioPorUsername(ctx, username)
}

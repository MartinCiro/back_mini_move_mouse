package usuarios

import "api_go/internal/core/usuarios"

func ToUsuarioData(dto CreateUsuarioRequest) usuarios.UsuarioData {
	return usuarios.UsuarioData{
		Username: dto.Username,
		Password: dto.Password,
		RolID:    dto.RolID,
		EstadoID: dto.EstadoID,
	}
}

func ToUsuarioDataUpdate(dto UpdateUsuarioRequest) usuarios.UsuarioDataUpdate {
	return usuarios.UsuarioDataUpdate{
		Documento: dto.Documento,
		Username:  dto.Username,
		RolID:     dto.RolID,
		EstadoID:  dto.EstadoID,
	}
}

func ToUsuarioDataXid(dto GetUsuarioRequest) usuarios.UsuarioDataXid {
	return usuarios.UsuarioDataXid{
		Documento: dto.Documento,
	}
}

func ToCambiarPasswordData(dto CambiarPasswordRequest) usuarios.CambiarPasswordData {
	return usuarios.CambiarPasswordData{
		Documento:      dto.Documento,
		PasswordActual: dto.PasswordActual,
		NuevoPassword:  dto.NuevoPassword,
	}
}

func FromUsuario(usuario *usuarios.Usuario) UsuarioResponse {
	return UsuarioResponse{
		Documento: usuario.Documento,
		Username:  usuario.Username,
		Rol:       usuario.RolNombre,
		Estado:    usuario.EstadoNombre,
		CreadoAt:  usuario.FechaRegistro,
	}
}

func FromUsuarios(usuariosList []usuarios.Usuario) []UsuarioResponse {
	responses := make([]UsuarioResponse, len(usuariosList))
	for i, usuario := range usuariosList {
		responses[i] = FromUsuario(&usuario)
	}
	return responses
}

func FromUsuarioConRelaciones(usuario *usuarios.UsuarioConRelaciones) UsuarioConRelacionesResponse {
	return UsuarioConRelacionesResponse{
		Documento: usuario.Documento,
		Username:  usuario.Username,
		RolID:     usuario.RolID,
		EstadoID:  usuario.EstadoID,
		RolNombre: usuario.RolNombre,
		Estado:    usuario.EstadoNombre,
		CreadoAt:  usuario.FechaRegistro,
	}
}

func FromUsuariosConRelaciones(usuariosList []usuarios.UsuarioConRelaciones) []UsuarioConRelacionesResponse {
	responses := make([]UsuarioConRelacionesResponse, len(usuariosList))
	for i, usuario := range usuariosList {
		responses[i] = FromUsuarioConRelaciones(&usuario)
	}
	return responses
}

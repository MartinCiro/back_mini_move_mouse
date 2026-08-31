// internal/interfaces/api/usuarios/dtos.go
package usuarios

import "time"

// --- REQUESTS ---

type CreateUsuarioRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
	RolID    *int   `json:"rol_id,omitempty"`
	EstadoID *int   `json:"estado_id,omitempty"`
}

type GetUsuarioRequest struct {
	Documento string `json:"documento" validate:"required"`
}

type UpdateUsuarioRequest struct {
	Documento string  `json:"-"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	RolID     *int    `json:"rol_id,omitempty"`
	EstadoID  *int    `json:"estado_id,omitempty"`
}

type CambiarPasswordRequest struct {
	Documento      string `json:"-"`
	PasswordActual string `json:"password_actual" validate:"required"`
	NuevoPassword  string `json:"nuevo_password" validate:"required,min=6"`
}

// --- RESPONSES ---

type UsuarioResponse struct {
	Documento string    `json:"documento"`
	Username  string    `json:"username"`
	Rol       string    `json:"rol"`
	Estado    string    `json:"estado"`
	CreadoAt  time.Time `json:"creado_at"`
}

type UsuarioConRelacionesResponse struct {
	Documento string    `json:"documento"`
	Username  string    `json:"username"`
	RolID     int       `json:"rol_id"`
	EstadoID  int       `json:"estado_id"`
	RolNombre string    `json:"rol_nombre"`
	Estado    string    `json:"estado"`
	CreadoAt  time.Time `json:"created_at"`
}

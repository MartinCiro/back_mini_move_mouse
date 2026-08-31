package usuarios

import "time"

// Usuario representa la entidad de dominio usuario
type Usuario struct {
	Documento     string    `json:"documento"`
	Username      string    `json:"username"`
	RolID         int       `json:"rol_id"`
	EstadoID      int       `json:"estado_id"`
	RolNombre     string    `json:"rol"`
	EstadoNombre  string    `json:"estado"`
	FechaRegistro time.Time `json:"fecha_registro"`
}

// UsuarioData contiene los datos para crear un nuevo usuario
type UsuarioData struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	RolID    *int   `json:"rol_id,omitempty"`
	EstadoID *int   `json:"estado_id,omitempty"`
}

// UsuarioDataXid contiene el ID para buscar o eliminar un usuario
type UsuarioDataXid struct {
	Documento string `json:"documento" validate:"required"`
}

// UsuarioDataUpdate contiene los datos para actualizar un usuario
type UsuarioDataUpdate struct {
	Documento string  `json:"documento" validate:"required"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	RolID     *int    `json:"rol_id,omitempty"`
	EstadoID  *int    `json:"estado_id,omitempty"`
}

// CambiarPasswordData contiene los datos para cambiar contraseña
type CambiarPasswordData struct {
	Documento      string `json:"documento" validate:"required"`
	PasswordActual string `json:"password_actual" validate:"required"`
	NuevoPassword  string `json:"nuevo_password" validate:"required,min=6"`
}

// UsuarioConRelaciones representa un usuario con información de rol y estado
type UsuarioConRelaciones struct {
	Documento     string    `json:"documento"`
	Username      string    `json:"username"`
	RolID         int       `json:"rol_id"`
	EstadoID      int       `json:"estado_id"`
	RolNombre     string    `json:"rol_nombre"`
	EstadoNombre  string    `json:"estado"`
	FechaRegistro time.Time `json:"fecha_registro"`
}

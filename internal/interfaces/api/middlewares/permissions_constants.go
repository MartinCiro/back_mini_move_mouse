// internal/interfaces/api/middlewares/permissions_constants.go
package middlewares

// Permisos para Estados
var (
	PermissionEstadosListar   = []string{"estado:listar", "estado:leer"}
	PermissionEstadosCrear    = []string{"estado:escribir", "estado:crear"}
	PermissionEstadosEditar   = []string{"estado:editar", "estado:actualizar"}
	PermissionEstadosEliminar = []string{"estado:eliminar", "estado:escribir"}
	PermissionEstadosVer      = []string{"estado:ver", "estado:leer"}
)

// Permisos para Estados
var (
	PermissionPermisosListar   = []string{"permiso:listar", "permiso:leer"}
	PermissionPermisosCrear    = []string{"permiso:escribir", "permiso:crear"}
	PermissionPermisosEditar   = []string{"permiso:editar", "permiso:actualizar"}
	PermissionPermisosEliminar = []string{"permiso:eliminar", "permiso:escribir"}
	PermissionPermisosVer      = []string{"permiso:ver", "permiso:leer"}
)

// Permisos para Roles
var (
	PermissionRolesVer      = []string{"rol:ver", "rol:leer"}
	PermissionRolesListar   = []string{"rol:listar", "rol:leer"}
	PermissionRolesCrear    = []string{"rol:escribir", "rol:crear"}
	PermissionRolesEliminar = []string{"rol:eliminar", "rol:elimina"}
	PermissionRolesEditar   = []string{"rol:editar", "rol:actualizar"}
	PermissionRolesPermisos = []string{"rol:permisos", "rol:leer_permisos"}
)

// Permisos para Usuarios
var (
	PermissionUsuariosListar    = []string{"usuario:listar", "usuario:leer"}
	PermissionUsuariosCrear     = []string{"usuario:crear", "usuario:escribir"}
	PermissionUsuariosEditar    = []string{"usuario:editar", "usuario:escribir"}
	PermissionUsuariosListarXid = []string{"usuario:listar_xid", "usuario:leer"}
)

// Permisos para login
var (
	PermissionLoginLogout = []string{"login:logout", "auth:logout"}
)

// Permisos administrativos
var (
	PermissionAdmin = []string{"admin", "administrador"}
)

// Permisos para Descargas
var (
	PermissionDescargasEjecutar = []string{"dwd:exec", "descarga:ejecutar"}
)

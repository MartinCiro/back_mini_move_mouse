package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"api_go/internal/core/auth"
	"api_go/internal/infrastructure/cookies"
	"api_go/internal/interfaces/api/common"
)

type AuthHandler struct {
	authService  *auth.AuthService
	cookieSigner *cookies.CookieSigner
}

func NewAuthHandler(authService *auth.AuthService, cookieSigner *cookies.CookieSigner) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		cookieSigner: cookieSigner,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Documento string `json:"documento"`
		Password  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteSimpleError(w, "Solicitud inválida: formato JSON incorrecto", 400)
		return
	}

	if req.Documento == "" || req.Password == "" {
		common.WriteSimpleError(w, "El documento y la contraseña son requeridos", 400)
		return
	}

	ctx := r.Context()
	loginReq := auth.LoginRequest{
		Documento: req.Documento,
		Password:  req.Password,
	}

	authResponse, signedCookie, expiresAt, err := h.authService.LoginUser(ctx, loginReq)
	if err != nil {
		common.WriteSimpleError(w, err.Error(), 401)
		return
	}

	if signedCookie != "" {
		h.cookieSigner.SetAuthCookie(w, signedCookie, expiresAt)
	}

	successResponse := common.NewSuccessResponse(authResponse.Message)
	common.WriteJSONResponse(w, successResponse, 200)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteSimpleError(w, "Solicitud inválida: formato JSON incorrecto", 400)
		return
	}

	// ✅ CORREGIDO: Validar Documento, Username y Password
	if req.Documento == "" || req.Username == "" || req.Password == "" {
		common.WriteSimpleError(w, "Documento, username y password son requeridos", 400)
		return
	}

	if len(req.Password) < 6 {
		common.WriteSimpleError(w, "El password debe tener al menos 6 caracteres", 400)
		return
	}

	ctx := r.Context()

	var currentUser *auth.User
	if user, ok := ctx.Value("user").(*auth.User); ok {
		currentUser = user
	}

	// ✅ CORREGIDO: RegisterUser ahora devuelve (msg string, cookie string, expires time.Time, err error)
	msg, signedCookie, expiresAt, err := h.authService.RegisterUser(ctx, req, currentUser)
	if err != nil {
		statusCode := 400
		if strings.Contains(err.Error(), "ha ocurrido un error en el servidor") {
			statusCode = 500
		} else if strings.Contains(err.Error(), "no tiene permisos") {
			statusCode = 403
		}

		common.WriteSimpleError(w, err.Error(), statusCode)
		return
	}

	if signedCookie != "" {
		h.cookieSigner.SetAuthCookie(w, signedCookie, expiresAt)
	}

	// ✅ SIMPLIFICADO: msg ya es un string, no necesitamos el switch
	successResponse := common.NewSuccessResponse(msg)
	common.WriteJSONResponse(w, successResponse, 201)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// ✅ CORREGIDO: userID ahora es string (Documento), no necesita conversión a int
	if userID, ok := ctx.Value("userID").(string); ok {
		h.authService.Logout(ctx, userID)
	}

	h.cookieSigner.ClearAuthCookie(w)

	successResponse := common.NewSuccessResponse("Sesión cerrada exitosamente")
	common.WriteJSONResponse(w, successResponse, 200)
}

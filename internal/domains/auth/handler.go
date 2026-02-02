package auth

import (
	"encoding/json"
	"go-web-template/internal/middleware"
	"go-web-template/internal/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service        AuthServiceInterface
	authMiddleware middleware.AuthMiddlewareInterface
}

func NewAuthHandler(
	srv AuthServiceInterface,
	authMiddleware middleware.AuthMiddlewareInterface,
) *AuthHandler {
	return &AuthHandler{
		service:        srv,
		authMiddleware: authMiddleware,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware.WebClientAuthentication)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.GetMe)
	})

	return r
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.ValidateCredentials(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authMiddleware.GenerateLoginTokens(user.ID, req.RememberMe)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.authMiddleware.SetLoginCookies(w, accessToken, refreshToken, req.RememberMe)

	utils.RespondJSON(w, http.StatusOK, MeResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Password != req.PasswordConfirmation {
		utils.RespondError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.DisplayName, req.Email, req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authMiddleware.GenerateLoginTokens(user.ID, false)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.authMiddleware.SetLoginCookies(w, accessToken, refreshToken, false)

	utils.RespondJSON(w, http.StatusCreated, MeResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, MeResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authMiddleware.ClearLoginCookies(w)
	utils.RespondSuccess(w, http.StatusOK, "Logged out successfully")
}

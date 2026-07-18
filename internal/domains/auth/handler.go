package auth

import (
	"encoding/json"
	"errors"
	"go-web-template/internal/middleware"
	"go-web-template/internal/sessions"
	"go-web-template/internal/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service        AuthServiceInterface
	authMiddleware middleware.AuthMiddlewareInterface
	logger         *zap.Logger
}

func NewAuthHandler(
	srv AuthServiceInterface,
	authMiddleware middleware.AuthMiddlewareInterface,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		service:        srv,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Get("/confirm-email", h.ConfirmEmail)
	r.Post("/resend-confirmation", h.ResendConfirmation)
	r.Post("/request-password-reset", h.RequestPasswordReset)
	r.Post("/reset-password", h.ResetPassword)
	// Logout reads the session cookie directly, so it stays public: an expired
	// session must still be able to clear its cookie.
	r.Post("/logout", h.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware.WebClientAuthentication)
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
		if errors.Is(err, ErrEmailNotConfirmed) {
			utils.RespondError(w, http.StatusForbidden, err.Error())
			return
		}
		utils.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, maxAge, err := h.authMiddleware.CreateLoginSession(r.Context(), user.ID, req.RememberMe)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.authMiddleware.SetSessionCookie(w, sessionID, maxAge)

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

	if err := h.service.Register(r.Context(), req.DisplayName, req.Email, req.Password); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// No auto-login: the account is unconfirmed until the emailed link is used.
	utils.RespondSuccess(w, http.StatusCreated, "Registration successful. Please check your email to confirm your account.")
}

func (h *AuthHandler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.RespondError(w, http.StatusBadRequest, "missing token")
		return
	}

	if err := h.service.ConfirmEmail(r.Context(), token); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	utils.RespondSuccess(w, http.StatusOK, "Email confirmed. You can now log in.")
}

func (h *AuthHandler) ResendConfirmation(w http.ResponseWriter, r *http.Request) {
	var req ConfirmResendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ResendConfirmation(r.Context(), req.Email); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, "If the account exists and is unconfirmed, a confirmation email has been sent.")
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.RequestPasswordReset(r.Context(), req.Email); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, "If the account exists, a password reset email has been sent.")
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := h.service.ResetPassword(r.Context(), req.Token, req.Password, req.PasswordConfirmation)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Revoke existing sessions so an old one can't outlive the password change.
	if err := h.authMiddleware.RevokeAllSessions(r.Context(), userID); err != nil {
		h.logger.Error("failed to revoke sessions after password reset", zap.Error(err))
	}

	utils.RespondSuccess(w, http.StatusOK, "Password updated. Please log in with your new password.")
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
	// Best effort: if the delete fails the session still dies at its TTL, so the
	// cookie is cleared either way.
	if cookie, err := r.Cookie(sessions.CookieName); err == nil && cookie.Value != "" {
		if err := h.authMiddleware.DestroySession(r.Context(), cookie.Value); err != nil {
			h.logger.Error("failed to destroy session", zap.Error(err))
		}
	}

	h.authMiddleware.ClearSessionCookie(w)
	utils.RespondSuccess(w, http.StatusOK, "Logged out successfully")
}

package handlers

import (
	"net/http"
	"strconv"

	"go-web-template/internal/services"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserHandler struct {
	service services.UserServiceInterface
	logger  *zap.Logger
}

func NewUserHandler(service services.UserServiceInterface, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger.Named("user-handler"),
	}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListUsers)
	return r
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := h.service.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		RespondError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	RespondJSON(w, http.StatusOK, result)
}

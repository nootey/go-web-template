package user

import (
	"go-web-template/internal/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserHandler struct {
	service UserServiceInterface
	logger  *zap.Logger
}

func NewUserHandler(srv UserServiceInterface, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: srv,
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
		utils.RespondError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

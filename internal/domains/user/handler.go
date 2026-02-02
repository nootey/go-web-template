package user

import (
	"go-web-template/internal/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(srv UserServiceInterface) *UserHandler {
	return &UserHandler{
		service: srv,
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
		utils.RespondError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

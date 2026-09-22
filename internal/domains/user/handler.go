package user

import (
	"go-web-template/internal/apperr"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service UserServiceInterface
	resp    *apperr.Responder
}

func NewUserHandler(srv UserServiceInterface, resp *apperr.Responder) *UserHandler {
	return &UserHandler{
		service: srv,
		resp:    resp,
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
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, result)
}

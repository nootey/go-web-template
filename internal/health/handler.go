package health

import (
	"go-web-template/internal/apperr"
	"net/http"
)

var statusCodes = map[Status]int{
	StatusOK:       http.StatusOK,
	StatusDegraded: http.StatusServiceUnavailable,
}

type HealthHandler struct {
	service *Service
	resp    *apperr.Responder
}

func NewHealthHandler(srv *Service, resp *apperr.Responder) *HealthHandler {
	return &HealthHandler{
		service: srv,
		resp:    resp,
	}
}

func (h *HealthHandler) Report(w http.ResponseWriter, _ *http.Request) {
	report := h.service.Report()
	h.resp.JSON(w, statusCodes[report.Status], report)
}

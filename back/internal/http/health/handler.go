package health

import (
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/http/response"
)

type Handler struct{}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "ok", http.StatusOK)
}

func NewHandler() *Handler {
	return &Handler{}
}

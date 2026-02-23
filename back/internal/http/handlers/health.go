package handlers

import (
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/http/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	DB *pgxpool.Pool
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "ok", http.StatusOK)
}

func (h *HealthHandler) HealthDB(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(r.Context()); err != nil {
		response.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	response.Success(w, "db available", http.StatusOK)
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		DB: db,
	}
}

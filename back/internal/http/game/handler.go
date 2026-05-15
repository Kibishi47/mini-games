package game

import (
	"net/http"

	domaingame "github.com/Kibishi47/mini-games/back/internal/domain/game"
	"github.com/Kibishi47/mini-games/back/internal/http/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	gameService domaingame.Service
}

func NewHandler(gameService domaingame.Service) *Handler {
	return &Handler{gameService: gameService}
}

func Routes(r chi.Router, handler *Handler) {
	r.Get("/games", handler.ListGames)
}

func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.gameService.ListGames(r.Context())
	if err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, games)
}

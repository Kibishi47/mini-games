package room

import (
	"encoding/json"
	"net/http"

	domainroom "github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/Kibishi47/mini-games/back/internal/http/middlewares"
	"github.com/Kibishi47/mini-games/back/internal/http/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	roomService domainroom.Service
}

func NewHandler(roomService domainroom.Service) *Handler {
	return &Handler{roomService: roomService}
}

func Routes(r chi.Router, handler *Handler, authMw func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Post("/rooms", handler.CreateRoom)
		r.Post("/rooms/join", handler.JoinRoom)
		r.Get("/rooms/{code}", handler.GetRoom)
	})
}

type JoinRoomRequest struct {
	Code string `json:"code"`
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		response.Error(w, "missing room code", http.StatusBadRequest)
		return
	}

	room, err := h.roomService.GetByCode(r.Context(), code)
	if err != nil {
		if err == domainroom.ErrRoomNotFound {
			response.Error(w, "room not found", http.StatusNotFound)
			return
		}
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, room)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	room, err := h.roomService.CreateRoom(r.Context(), userID)
	if err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusCreated, room)
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	room, err := h.roomService.JoinRoom(r.Context(), userID, req.Code)
	if err != nil {
		if err == domainroom.ErrRoomNotFound {
			response.Error(w, "room not found", http.StatusNotFound)
			return
		}
		if err == domainroom.ErrRoomNotLobby {
			response.Error(w, "room is already running or closed", http.StatusBadRequest)
			return
		}
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, room)
}

package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/service"
)

type NoteHandler struct {
	noteService *service.NoteService
}

func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.CreateNoteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	note, err := h.noteService.Create(r.Context(), userID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrInvalidNote) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Println("create note error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	writeJSON(w, http.StatusCreated, toNoteResponse(*note))
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notes, err := h.noteService.List(r.Context(), userID)
	if err != nil {
		log.Println("list notes error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response := []dto.NoteResponse{}
	for _, n := range notes {
		response = append(response, toNoteResponse(n))
	}
	writeJSON(w, http.StatusOK, response)
}

func toNoteResponse(n model.Note) dto.NoteResponse {
	return dto.NoteResponse{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

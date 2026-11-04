package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type TutorHandler struct {
	tutorService *service.TutorService
}

func NewTutorHandler(tutorService *service.TutorService) *TutorHandler {
	return &TutorHandler{tutorService: tutorService}
}

func (h *TutorHandler) Ask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var req dto.AskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	question := strings.TrimSpace(req.Question)
	if question == "" || len(question) > 500 {
		writeError(w, http.StatusBadRequest, "question is required (max 500 characters)")
		return
	}

	result, err := h.tutorService.Ask(r.Context(), userID, question)
	if err != nil {
		log.Println("ask error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	sources := []dto.PassageResponse{}
	for _, c := range result.Sources {
		sources = append(sources, dto.PassageResponse{
			NoteID:    c.NoteID,
			NoteTitle: c.NoteTitle,
			Content:   c.Content,
			Score:     c.Rank,
		})
	}
	writeJSON(w, http.StatusOK, dto.AskResponse{Answer: result.Answer, Sources: sources})
}

package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/sashank33v/recallengine/internal/apperror"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type GenerationHandler struct {
	generationService *service.GenerationService
}

func NewGenerationHandler(generationService *service.GenerationService) *GenerationHandler {
	return &GenerationHandler{generationService: generationService}
}

func (h *GenerationHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	noteID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	jobID, err := h.generationService.RequestGeneration(r.Context(), userID, noteID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrNoteNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, apperror.ErrJobInProgress):
			writeError(w, http.StatusConflict, err.Error())
		default:
			log.Println("request generation error:", err)
			writeError(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{"job_id": jobID, "status": "pending"})
}

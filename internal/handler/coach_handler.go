package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type CoachHandler struct {
	coachService *service.CoachService
}

func NewCoachHandler(coachService *service.CoachService) *CoachHandler {
	return &CoachHandler{coachService: coachService}
}

func (h *CoachHandler) Ask(w http.ResponseWriter, r *http.Request) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	answer, err := h.coachService.Ask(ctx, userID, question)
	if err != nil {
		log.Println("coach error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	writeJSON(w, http.StatusOK, dto.CoachResponse{Answer: answer})
}

package handler

import (
	"log"
	"net/http"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type MasteryHandler struct {
	masteryservice *service.MasteryService
}

func NewMasterHandler(masteryService *service.MasteryService) *MasteryHandler {
	return &MasteryHandler{masteryservice: masteryService}
}

func (h *MasteryHandler) List(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.masteryservice.List(r.Context(), userID)
	if err != nil {
		log.Println("list mastery error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response := []dto.TopicMasteryResponse{}
	for _, m := range items {
		response = append(response, dto.TopicMasteryResponse{
			Topic:          m.Topic,
			TotalAnswers:   m.TotalAnswers,
			CorrectAnswers: m.CorrectAnswers,
			MasteryPercent: m.CorrectAnswers * 100 / m.TotalAnswers,
		})

	}
	writeJSON(w, http.StatusOK, response)
}

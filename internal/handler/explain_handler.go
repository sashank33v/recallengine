package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type ExplainHandler struct {
	explainService *service.ExplainService
}

func NewExplainHandler(explainService *service.ExplainService) *ExplainHandler {
	return &ExplainHandler{explainService: explainService}
}

func (h *ExplainHandler) Explain(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)

	var req dto.ExplainRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.QuestionID <= 0 {
		writeError(w, http.StatusBadRequest, "question_id is required")
		return
	}

	explanation, err := h.explainService.Explain(r.Context(), userID, req.QuestionID)
	if err != nil {
		if errors.Is(err, service.ErrQuestionNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		log.Println("explain error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, dto.ExplainResponse{Explaination: explanation})
}

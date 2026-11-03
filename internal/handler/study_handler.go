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

type StudyHandler struct {
	studyService *service.StudyService
}

func NewStudyHandler(studyService *service.StudyService) *StudyHandler {
	return &StudyHandler{studyService: studyService}
}

func (h *StudyHandler) Due(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	questions, err := h.studyService.DueQuestions(r.Context(), userID)
	if err != nil {
		log.Println("due questions error:", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	response := []dto.QuestionResponse{}
	for _, ques := range questions {
		response = append(response, dto.QuestionResponse{
			ID:        ques.ID,
			Question:  ques.QuestionText,
			Answer:    ques.AnswerText,
			Source:    ques.Source,
			Status:    ques.Status,
			CreatedAt: ques.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *StudyHandler) Answer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.AnswerRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	state, dueAt, err := h.studyService.Answer(r.Context(), userID, int64(req.QuestionID), req.Quality)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuality):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrQuestionNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			log.Println("answer error:", err)
			writeError(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.AnswerResponse{
		QuestionID:   int64(req.QuestionID),
		Quality:      req.Quality,
		IntervalDays: state.IntervalDays,
		Repetitions:  state.Repetitions,
		EaseFactor:   state.EaseFactor,
		NextDueAt:    dueAt,
	})
}

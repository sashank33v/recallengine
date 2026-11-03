package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type QuestionHandler struct {
	questionService *service.QuestionService
}

func NewQuestionHandler(questionService *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{questionService: questionService}
}

func (h *QuestionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.CreateQuestionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	q, err := h.questionService.Create(r.Context(), userID, req.Question, req.Answer)
	if err != nil {
		if errors.Is(err, service.ErrInvalidQuestion) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Println("create question error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	writeJSON(w, http.StatusCreated, dto.QuestionResponse{
		ID:        q.ID,
		Question:  q.QuestionText,
		Answer:    q.AnswerText,
		Source:    q.Source,
		Status:    q.Status,
		CreatedAt: q.CreatedAt,
	})
}
func (h *QuestionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	questions, err := h.questionService.List(r.Context(), userID, status, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Println("list questions error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	response := []dto.QuestionResponse{}
	for _, q := range questions {
		response = append(response, dto.QuestionResponse{
			ID:        q.ID,
			Question:  q.QuestionText,
			Answer:    q.AnswerText,
			Source:    q.Source,
			Status:    q.Status,
			CreatedAt: q.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

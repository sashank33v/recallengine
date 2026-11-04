package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/sashank33v/recallengine/internal/apperror"
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

	q, err := h.questionService.Create(r.Context(), userID, req.Question, req.Answer, req.Topic)
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
func (h *QuestionHandler) Approve(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	questionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}
	err = h.questionService.Approve(r.Context(), userID, questionID)
	if err != nil {
		if errors.Is(err, apperror.ErrDraftNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		log.Println("approve question error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": questionID, "status": "approved"})
}

func (h *QuestionHandler) Reject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	questionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}
	err = h.questionService.Reject(r.Context(), userID, questionID)
	if err != nil {
		if errors.Is(err, apperror.ErrDraftNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Println("reject question error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": questionID, "status": "rejected"})
}
func (h *QuestionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	questionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.CreateQuestionRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	q, err := h.questionService.UpdateDraft(r.Context(), userID, questionID, req.Question, req.Answer)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuestion):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, apperror.ErrDraftNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			log.Println("update question error:", err)
			writeError(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.QuestionResponse{
		ID:        q.ID,
		Question:  q.QuestionText,
		Answer:    q.AnswerText,
		Source:    q.Source,
		Status:    q.Status,
		CreatedAt: q.CreatedAt,
	})
}

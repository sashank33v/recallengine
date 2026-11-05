package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/sashank33v/recallengine/internal/apperror"
	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type ConversationHandler struct {
	conversationService *service.ConversationService
}

func NewConversationHandler(conversationService *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{conversationService: conversationService}
}

func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversations, err := h.conversationService.List(r.Context(), userID)
	if err != nil {
		log.Println("list conversations error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response := []dto.ConversationResponse{}
	for _, c := range conversations {
		response = append(response, dto.ConversationResponse{
			ID:        c.ID,
			Title:     c.Title,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ConversationHandler) Messages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	messages, err := h.conversationService.Messages(r.Context(), userID, conversationID)
	if err != nil {
		if errors.Is(err, apperror.ErrConversationNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Println("list messages error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response := []dto.MessageResponse{}
	for _, m := range messages {
		response = append(response, dto.MessageResponse{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

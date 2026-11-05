package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sashank33v/recallengine/internal/apperror"
	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/service"
)

type TutorHandler struct {
	tutorService        *service.TutorService
	conversationService *service.ConversationService
}

func NewTutorHandler(tutorService *service.TutorService, conversationService *service.ConversationService) *TutorHandler {
	return &TutorHandler{tutorService: tutorService, conversationService: conversationService}
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
func toPassages(chunks []model.Chunk) []dto.PassageResponse {
	passages := []dto.PassageResponse{}
	for _, c := range chunks {
		passages = append(passages, dto.PassageResponse{
			NoteID:    c.NoteID,
			NoteTitle: c.NoteTitle,
			Content:   c.Content,
			Score:     c.Rank,
		})
	}
	return passages
}

func (h *TutorHandler) Chat(w http.ResponseWriter, r *http.Request) {
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}
	conversationID, err := h.conversationService.Start(r.Context(), userID, req.ConversationID, question)
	if err != nil {
		if errors.Is(err, apperror.ErrConversationFull) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
	}

	if err != nil {
		if errors.Is(err, apperror.ErrConversationNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Println("start conversation error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	err = h.conversationService.AddMessage(r.Context(), conversationID, "user", question)
	if err != nil {
		log.Println("save question error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	send := func(event string, data any) error {
		payload, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	send("conversation", map[string]int64{"id": conversationID})
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	result, err := h.tutorService.AskStream(ctx, userID, question,
		func(sources []model.Chunk) error { return send("sources", toPassages(sources)) },
		func(token string) error { return send("token", token) },
	)
	if err != nil {
		log.Println("chat error:", err)
		send("error", "something went wrong")
		return
	}

	if result.Replaced {
		send("replace", result.Answer)
	}
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer saveCancel()

	err = h.conversationService.AddMessage(saveCtx, conversationID, "assistant", result.Answer)
	if err != nil {
		log.Println("save answer error:", err)
	}
	send("done", map[string]bool{"ok": true})
}

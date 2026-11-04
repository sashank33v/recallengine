package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" || len(q) > 300 {
		writeError(w, http.StatusBadRequest, "q is required (max 300 characters)")
		return
	}

	chunks, err := h.searchService.FindPassages(r.Context(), userID, q, 5)
	if err != nil {
		log.Println("search error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response := []dto.PassageResponse{}
	for _, c := range chunks {
		response = append(response, dto.PassageResponse{
			NoteID:    c.NoteID,
			NoteTitle: c.NoteTitle,
			Content:   c.Content,
			Score:     c.Rank,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

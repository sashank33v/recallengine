package service

import (
	"context"
	"strings"
	"unicode"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
)

type SearchService struct {
	chunkRepo *repository.ChunkRepository
}

func NewSearchService(chunkRepo *repository.ChunkRepository) *SearchService {
	return &SearchService{chunkRepo: chunkRepo}
}

func (s *SearchService) FindPassages(ctx context.Context, userID int64, question string, limit int) ([]model.Chunk, error) {
	searchQuery := buildSearchQuery(question)
	if searchQuery == "" {
		return []model.Chunk{}, nil
	}
	return s.chunkRepo.Search(ctx, userID, searchQuery, limit)
}
func buildSearchQuery(question string) string {
	words := strings.FieldsFunc(question, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(words) > 20 {
		words = words[:20]
	}
	return strings.Join(words, " or ")
}

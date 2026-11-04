package service

import (
	"context"

	"github.com/sashank33v/recallengine/internal/repository"
)

type GenerationService struct {
	generationRepo *repository.GenerationRepository
}

func NewGenerationService(generationRepo *repository.GenerationRepository) *GenerationService {
	return &GenerationService{generationRepo: generationRepo}
}

func (s *GenerationService) RequestGeneration(ctx context.Context, userID, noteID int64) (int64, error) {
	return s.generationRepo.CreateJob(ctx, userID, noteID)
}

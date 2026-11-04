package service

import (
	"context"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
)

type MasteryService struct {
	masteryrepo *repository.MasteryRepository
}

func NewMasterService(masteryrepo *repository.MasteryRepository) *MasteryService {
	return &MasteryService{masteryrepo: masteryrepo}
}

func (s *MasteryService) List(ctx context.Context, userID int64) ([]model.TopicMastery, error) {
	return s.masteryrepo.ListByUser(ctx, userID)
}

package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/premium/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type PremiumService interface {
	GetLatest(ctx context.Context) (*dto.PremiumLatestResponse, error)
	GetHistory(ctx context.Context, req dto.PremiumHistoryRequest) (*dto.PremiumHistoryResponse, error)
	SaveTick(ctx context.Context, snapshot model.PremiumSnapshot) error
	Hub() *Hub
}

type premiumService struct {
	repo   repository.PremiumRepository
	hub    *Hub
	logger *slog.Logger
}

func NewPremiumService(repo repository.PremiumRepository, hub *Hub, logger *slog.Logger) PremiumService {
	return &premiumService{
		repo:   repo,
		hub:    hub,
		logger: logger,
	}
}

func (s *premiumService) GetLatest(ctx context.Context) (*dto.PremiumLatestResponse, error) {
	ticks, err := s.repo.FindLatest(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.PremiumTickResponse, 0, len(ticks))
	for _, t := range ticks {
		resp = append(resp, toTickResponse(t))
	}

	return &dto.PremiumLatestResponse{Ticks: resp}, nil
}

func (s *premiumService) GetHistory(ctx context.Context, req dto.PremiumHistoryRequest) (*dto.PremiumHistoryResponse, error) {
	pair := req.Pair
	if pair == "" {
		pair = model.PairBTCKRW
	}

	var from, to time.Time
	if req.From != "" {
		parsed, err := time.Parse(time.RFC3339, req.From)
		if err == nil {
			from = parsed
		}
	}
	if req.To != "" {
		parsed, err := time.Parse(time.RFC3339, req.To)
		if err == nil {
			to = parsed
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}

	ticks, err := s.repo.FindHistory(ctx, pair, from, to, limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	points := make([]dto.PremiumTickResponse, 0, len(ticks))
	for _, t := range ticks {
		points = append(points, toTickResponse(t))
	}

	return &dto.PremiumHistoryResponse{
		Pair:   pair,
		Points: points,
	}, nil
}

func (s *premiumService) SaveTick(ctx context.Context, snapshot model.PremiumSnapshot) error {
	tick := &model.PremiumTick{
		Pair:              snapshot.Pair,
		KoreanPrice:       snapshot.KoreanPrice,
		GlobalPrice:       snapshot.GlobalPrice,
		PremiumPct:        snapshot.PremiumPct,
		ReversePremiumPct: snapshot.ReversePremiumPct,
		SourceKR:          snapshot.SourceKR,
		SourceGL:          snapshot.SourceGL,
	}

	if err := s.repo.Create(ctx, tick); err != nil {
		return err
	}

	s.hub.BroadcastTick(snapshot)

	return nil
}

func (s *premiumService) Hub() *Hub {
	return s.hub
}

func toTickResponse(t model.PremiumTick) dto.PremiumTickResponse {
	return dto.PremiumTickResponse{
		Pair:              t.Pair,
		KoreanPrice:       t.KoreanPrice,
		GlobalPrice:       t.GlobalPrice,
		PremiumPct:        t.PremiumPct,
		ReversePremiumPct: t.ReversePremiumPct,
		SourceKR:          t.SourceKR,
		SourceGL:          t.SourceGL,
		Timestamp:         t.CreatedAt.Format(time.RFC3339),
	}
}

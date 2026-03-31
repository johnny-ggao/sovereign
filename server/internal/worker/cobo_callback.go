package worker

import (
	"context"
	"log/slog"

	"github.com/sovereign-fund/sovereign/internal/modules/wallet/service"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
)

// WebhookProcessor handles incoming Cobo webhook callbacks
// In production, this would be triggered by the HTTP webhook endpoint
// and process events asynchronously via a queue
type WebhookProcessor struct {
	walletSvc service.WalletService
	provider  cobo.WalletProvider
	logger    *slog.Logger
}

func NewWebhookProcessor(
	ws service.WalletService,
	provider cobo.WalletProvider,
	logger *slog.Logger,
) *WebhookProcessor {
	return &WebhookProcessor{
		walletSvc: ws,
		provider:  provider,
		logger:    logger,
	}
}

func (p *WebhookProcessor) ProcessCallback(ctx context.Context, signature string, payload []byte, data cobo.WebhookPayload) error {
	valid, err := p.provider.VerifyWebhook(signature, payload)
	if err != nil || !valid {
		p.logger.Warn("invalid webhook signature")
		return err
	}

	if err := p.walletSvc.HandleWebhook(ctx, data); err != nil {
		p.logger.Error("webhook processing failed",
			slog.String("type", data.Type),
			slog.String("id", data.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	p.logger.Info("webhook processed",
		slog.String("type", data.Type),
		slog.String("id", data.ID),
		slog.String("status", data.Status),
	)
	return nil
}

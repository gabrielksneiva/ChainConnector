package service

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NetworkService struct {
	repo     ports.NetworkRepositoryPort
	producer ports.NetworkProducerPort
	logger   *zap.Logger
}

func NewNetworkService(repo ports.NetworkRepositoryPort, producer ports.NetworkProducerPort, logger *zap.Logger) *NetworkService {
	return &NetworkService{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (s *NetworkService) RegisterNetwork(ctx context.Context, network *entity.Network) error {
	if err := normalizeNetwork(network); err != nil {
		return err
	}

	if s.producer != nil && s.producer.Enabled() {
		if err := s.producer.EnqueueNetworkRegistration(ctx, network); err != nil {
			return err
		}
		s.logger.Info("network registration queued", zap.String("network_id", network.ID), zap.String("name", network.Name), zap.Int64("chain_id", network.ChainID))
		return nil
	}

	if s.repo == nil {
		return errors.New("network repository is required")
	}
	if err := s.repo.SaveNetwork(ctx, network); err != nil {
		return err
	}
	s.logger.Info("network registered directly", zap.String("network_id", network.ID), zap.String("name", network.Name), zap.Int64("chain_id", network.ChainID))
	return nil
}

func (s *NetworkService) ListNetworks(ctx context.Context) ([]*entity.Network, error) {
	if s.repo == nil {
		return nil, errors.New("network repository is required")
	}
	return s.repo.ListNetworks(ctx)
}

func (s *NetworkService) FindNetworkByID(ctx context.Context, id string) (*entity.Network, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("network id is required")
	}
	if s.repo == nil {
		return nil, errors.New("network repository is required")
	}
	return s.repo.FindNetworkByID(ctx, id)
}

func normalizeNetwork(network *entity.Network) error {
	if network == nil {
		return errors.New("network is required")
	}

	network.Name = strings.TrimSpace(strings.ToLower(network.Name))
	network.RPCURL = strings.TrimSpace(network.RPCURL)
	network.CurrencySymbol = strings.TrimSpace(strings.ToUpper(network.CurrencySymbol))
	network.ExplorerURL = strings.TrimSpace(network.ExplorerURL)

	if network.Name == "" {
		return errors.New("network name is required")
	}
	if network.ChainID <= 0 {
		return errors.New("network chain_id must be greater than zero")
	}
	if err := validateOptionalHTTPURL("rpc_url", network.RPCURL); err != nil {
		return err
	}
	if err := validateOptionalHTTPURL("explorer_url", network.ExplorerURL); err != nil {
		return err
	}
	if network.ID == "" {
		network.ID = uuid.NewString()
	}

	now := time.Now().UTC()
	if network.CreatedAt.IsZero() {
		network.CreatedAt = now
	}
	network.UpdatedAt = now
	return nil
}

func validateOptionalHTTPURL(field string, value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid URL", field)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return nil
	default:
		return fmt.Errorf("%s must use http or https", field)
	}
}

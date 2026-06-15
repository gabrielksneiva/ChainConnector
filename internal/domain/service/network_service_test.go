package service

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

type fakeNetworkRepo struct {
	saved  *entity.Network
	findID string
}

func (f *fakeNetworkRepo) SaveNetwork(ctx context.Context, network *entity.Network) error {
	cp := *network
	f.saved = &cp
	return nil
}

func (f *fakeNetworkRepo) FindNetworkByID(ctx context.Context, id string) (*entity.Network, error) {
	f.findID = id
	return &entity.Network{ID: id, Name: "sepolia", ChainID: 11155111, Enabled: true}, nil
}

func (f *fakeNetworkRepo) ListNetworks(ctx context.Context) ([]*entity.Network, error) {
	return []*entity.Network{{ID: "network-1", Name: "sepolia", ChainID: 11155111, Enabled: true}}, nil
}

type fakeNetworkProducer struct {
	enabled bool
	sent    *entity.Network
	err     error
}

func (f *fakeNetworkProducer) Enabled() bool {
	return f.enabled
}

func (f *fakeNetworkProducer) EnqueueNetworkRegistration(ctx context.Context, network *entity.Network) error {
	if f.err != nil {
		return f.err
	}
	cp := *network
	f.sent = &cp
	return nil
}

func TestNetworkServiceRegisterDirect(t *testing.T) {
	repo := &fakeNetworkRepo{}
	svc := NewNetworkService(repo, nil, zap.NewNop())

	network := &entity.Network{
		Name:           " Sepolia ",
		ChainID:        11155111,
		RPCURL:         "http://localhost:8545",
		CurrencySymbol: " eth ",
		ExplorerURL:    "https://sepolia.etherscan.io",
	}
	if err := svc.RegisterNetwork(context.Background(), network); err != nil {
		t.Fatalf("RegisterNetwork returned error: %v", err)
	}

	if repo.saved == nil {
		t.Fatalf("expected network to be saved")
	}
	if repo.saved.Name != "sepolia" {
		t.Fatalf("expected normalized name, got %q", repo.saved.Name)
	}
	if repo.saved.CurrencySymbol != "ETH" {
		t.Fatalf("expected normalized currency symbol, got %q", repo.saved.CurrencySymbol)
	}
	if repo.saved.ID == "" {
		t.Fatalf("expected generated network id")
	}
	if repo.saved.CreatedAt.IsZero() || repo.saved.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestNetworkServiceRegisterQueued(t *testing.T) {
	repo := &fakeNetworkRepo{}
	producer := &fakeNetworkProducer{enabled: true}
	svc := NewNetworkService(repo, producer, zap.NewNop())

	err := svc.RegisterNetwork(context.Background(), &entity.Network{Name: "sepolia", ChainID: 11155111})
	if err != nil {
		t.Fatalf("RegisterNetwork returned error: %v", err)
	}
	if producer.sent == nil {
		t.Fatalf("expected network to be sent to producer")
	}
	if repo.saved != nil {
		t.Fatalf("expected queued registration not to save directly")
	}
}

func TestNetworkServiceRegisterProducerError(t *testing.T) {
	producer := &fakeNetworkProducer{enabled: true, err: errors.New("sqs unavailable")}
	svc := NewNetworkService(&fakeNetworkRepo{}, producer, zap.NewNop())

	err := svc.RegisterNetwork(context.Background(), &entity.Network{Name: "sepolia", ChainID: 11155111})
	if err == nil {
		t.Fatalf("expected producer error")
	}
}

func TestNetworkServiceRejectsInvalidURL(t *testing.T) {
	svc := NewNetworkService(&fakeNetworkRepo{}, nil, zap.NewNop())

	err := svc.RegisterNetwork(context.Background(), &entity.Network{
		Name:    "sepolia",
		ChainID: 11155111,
		RPCURL:  "file:///tmp/socket",
	})
	if err == nil {
		t.Fatalf("expected invalid URL error")
	}
}

func TestNetworkServiceFindTrimsID(t *testing.T) {
	repo := &fakeNetworkRepo{}
	svc := NewNetworkService(repo, nil, zap.NewNop())

	_, err := svc.FindNetworkByID(context.Background(), " network-1 ")
	if err != nil {
		t.Fatalf("FindNetworkByID returned error: %v", err)
	}
	if repo.findID != "network-1" {
		t.Fatalf("expected trimmed id, got %q", repo.findID)
	}
}

package monitor

import (
	"strings"
	"sync"

	"ChainConnector/internal/domain/entity"
)

type InterestStore struct {
	mu        sync.RWMutex
	addresses map[string]struct{}
	topics    [][]string
	txHashes  map[string]struct{}
}

func NewInterestStore() *InterestStore {
	return &InterestStore{
		addresses: make(map[string]struct{}),
		topics:    make([][]string, 0),
		txHashes:  make(map[string]struct{}),
	}
}

func normalizeAddress(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

func normalizeTopic(topic string) string {
	return strings.ToLower(strings.TrimSpace(topic))
}

func (s *InterestStore) AddAddress(address string) {
	if address == "" {
		return
	}
	addr := normalizeAddress(address)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addresses[addr] = struct{}{}
}

func (s *InterestStore) AddTxHash(txHash string) {
	if txHash == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.txHashes[normalizeTopic(txHash)] = struct{}{}
}

func (s *InterestStore) AddTopics(topics []string) {
	if len(topics) == 0 {
		return
	}
	normalized := make([]string, 0, len(topics))
	for _, t := range topics {
		t = normalizeTopic(t)
		if t != "" {
			normalized = append(normalized, t)
		}
	}
	if len(normalized) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.topics = append(s.topics, normalized)
}

func (s *InterestStore) HasTxHash(txHash string) bool {
	if txHash == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.txHashes[normalizeTopic(txHash)]
	return ok
}

func (s *InterestStore) GetAddresses() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, 0, len(s.addresses))
	for addr := range s.addresses {
		result = append(result, addr)
	}
	return result
}

func (s *InterestStore) GetTopics() [][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([][]string, 0, len(s.topics))
	for _, topicSet := range s.topics {
		copySet := make([]string, len(topicSet))
		copy(copySet, topicSet)
		result = append(result, copySet)
	}
	return result
}

func (s *InterestStore) GetTxHashes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, 0, len(s.txHashes))
	for txHash := range s.txHashes {
		result = append(result, txHash)
	}
	return result
}

func (s *InterestStore) ToLogFilter(fromBlock, toBlock *uint64) entity.LogFilter {
	return entity.LogFilter{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: s.GetAddresses(),
		Topics:    s.GetTopics(),
	}
}

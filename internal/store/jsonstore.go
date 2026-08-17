package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"water-quality/internal/domain"
)

type DB struct {
	Points      map[string]*domain.Point            `json:"points"`
	Batches     map[string]*domain.Batch            `json:"batches"`
	Bottles     map[string]*domain.Bottle           `json:"bottles"`
	Handovers   map[string][]*domain.HandoverRecord `json:"handovers"`
	Conclusions map[string]*domain.Conclusion       `json:"conclusions"`
}

type JSONStore struct {
	mu   sync.RWMutex
	file string
	db   *DB
}

func NewJSONStore(file string) (*JSONStore, error) {
	s := &JSONStore{file: file, db: &DB{
		Points:      make(map[string]*domain.Point),
		Batches:     make(map[string]*domain.Batch),
		Bottles:     make(map[string]*domain.Bottle),
		Handovers:   make(map[string][]*domain.HandoverRecord),
		Conclusions: make(map[string]*domain.Conclusion),
	}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *JSONStore) load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, s.db)
}

func (s *JSONStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *JSONStore) saveLocked() error {
	data, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.file)
}

func (s *JSONStore) Close() error {
	return s.save()
}

func (s *JSONStore) SavePoint(p *domain.Point) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Points[p.ID] = p
	return s.saveLocked()
}

func (s *JSONStore) GetPoint(id string) (*domain.Point, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.db.Points[id]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (s *JSONStore) ListPoints() ([]*domain.Point, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*domain.Point, 0, len(s.db.Points))
	for _, p := range s.db.Points {
		res = append(res, p)
	}
	return res, nil
}

func (s *JSONStore) SaveBatch(b *domain.Batch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Batches[b.ID] = b
	return s.saveLocked()
}

func (s *JSONStore) GetBatch(id string) (*domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if b, ok := s.db.Batches[id]; ok {
		return b, nil
	}
	return nil, ErrNotFound
}

func (s *JSONStore) ListBatches() ([]*domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*domain.Batch, 0, len(s.db.Batches))
	for _, b := range s.db.Batches {
		res = append(res, b)
	}
	return res, nil
}

func (s *JSONStore) SaveBottle(b *domain.Bottle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Bottles[b.ID] = b
	return s.saveLocked()
}

func (s *JSONStore) GetBottle(id string) (*domain.Bottle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if b, ok := s.db.Bottles[id]; ok {
		return b, nil
	}
	return nil, ErrNotFound
}

func (s *JSONStore) ListBottlesByBatch(batchID string) ([]*domain.Bottle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*domain.Bottle, 0)
	for _, b := range s.db.Bottles {
		if b.BatchID == batchID {
			res = append(res, b)
		}
	}
	return res, nil
}

func (s *JSONStore) AddHandover(r *domain.HandoverRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Handovers[r.BottleID] = append(s.db.Handovers[r.BottleID], r)
	return s.saveLocked()
}

func (s *JSONStore) ListHandoverByBottle(bottleID string) ([]*domain.HandoverRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Handovers[bottleID], nil
}

func (s *JSONStore) SaveConclusion(c *domain.Conclusion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Conclusions[c.BatchID] = c
	return s.saveLocked()
}

func (s *JSONStore) GetConclusion(batchID string) (*domain.Conclusion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.db.Conclusions[batchID]; ok {
		return c, nil
	}
	return nil, ErrNotFound
}

// 确保io.Reader被使用（为避免未使用导入）
var _ io.Reader

func debug(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

package store

import (
	"errors"

	"water-quality/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store interface {
	SavePoint(p *domain.Point) error
	GetPoint(id string) (*domain.Point, error)
	ListPoints() ([]*domain.Point, error)

	SaveBatch(b *domain.Batch) error
	GetBatch(id string) (*domain.Batch, error)
	ListBatches() ([]*domain.Batch, error)

	SaveBottle(b *domain.Bottle) error
	GetBottle(id string) (*domain.Bottle, error)
	ListBottlesByBatch(batchID string) ([]*domain.Bottle, error)

	AddHandover(r *domain.HandoverRecord) error
	ListHandoverByBottle(bottleID string) ([]*domain.HandoverRecord, error)

	SaveConclusion(c *domain.Conclusion) error
	GetConclusion(batchID string) (*domain.Conclusion, error)

	Close() error
}

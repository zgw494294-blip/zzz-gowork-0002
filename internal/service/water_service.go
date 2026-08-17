package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"water-quality/internal/domain"
	"water-quality/internal/store"
)

var (
	ErrInvalidState = errors.New("invalid state")
	ErrInvalidInput = errors.New("invalid input")
)

type WaterService struct {
	st  store.Store
	now func() time.Time
}

func NewWaterService(st store.Store) *WaterService {
	return &WaterService{
		st:  st,
		now: time.Now,
	}
}

func (s *WaterService) SetNowFunc(f func() time.Time) {
	s.now = f
}

func (s *WaterService) genID(prefix string) (string, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b)), nil
}

// CreatePoint creates a new sampling point
func (s *WaterService) CreatePoint(name, location string) (*domain.Point, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	id, err := s.genID("pt")
	if err != nil {
		return nil, err
	}
	p := &domain.Point{ID: id, Name: name, Location: location}
	if err := s.st.SavePoint(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *WaterService) GetPoint(id string) (*domain.Point, error) {
	return s.st.GetPoint(id)
}

func (s *WaterService) ListPoints() ([]*domain.Point, error) {
	return s.st.ListPoints()
}

// CreateBatch creates a batch attached to a point
func (s *WaterService) CreateBatch(pointID, batchNo string) (*domain.Batch, error) {
	if batchNo == "" {
		return nil, fmt.Errorf("%w: batch_no required", ErrInvalidInput)
	}
	if _, err := s.st.GetPoint(pointID); err != nil {
		return nil, fmt.Errorf("point %s: %w", pointID, err)
	}
	id, err := s.genID("batch")
	if err != nil {
		return nil, err
	}
	now := s.now()
	b := &domain.Batch{
		ID:        id,
		PointID:   pointID,
		BatchNo:   batchNo,
		Status:    domain.BatchCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.st.SaveBatch(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *WaterService) GetBatch(id string) (*domain.Batch, error) {
	return s.st.GetBatch(id)
}

func (s *WaterService) GetBottle(id string) (*domain.Bottle, error) {
	return s.st.GetBottle(id)
}

func (s *WaterService) ListBottlesByBatch(batchID string) ([]*domain.Bottle, error) {
	return s.st.ListBottlesByBatch(batchID)
}

func (s *WaterService) ListBatches() ([]*domain.Batch, error) {
	return s.st.ListBatches()
}

// AddBottle adds a bottle to a batch. Only allowed when batch is created.
func (s *WaterService) AddBottle(batchID, code string) (*domain.Bottle, error) {
	if code == "" {
		return nil, fmt.Errorf("%w: code required", ErrInvalidInput)
	}
	b, err := s.st.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	if b.Status != domain.BatchCreated {
		return nil, fmt.Errorf("%w: batch is %s, only created can add bottles", ErrInvalidState, b.Status)
	}
	id, err := s.genID("btl")
	if err != nil {
		return nil, err
	}
	btl := &domain.Bottle{
		ID:      id,
		BatchID: batchID,
		Code:    code,
		Status:  domain.BottlePending,
	}
	if err := s.st.SaveBottle(btl); err != nil {
		return nil, err
	}
	return btl, nil
}

// CompleteSampling marks all bottles in batch as sampled and updates batch status to sampled.
func (s *WaterService) CompleteSampling(batchID string) error {
	b, err := s.st.GetBatch(batchID)
	if err != nil {
		return err
	}
	if b.Status != domain.BatchCreated {
		return fmt.Errorf("%w: batch is %s", ErrInvalidState, b.Status)
	}
	bottles, err := s.st.ListBottlesByBatch(batchID)
	if err != nil {
		return err
	}
	if len(bottles) == 0 {
		return fmt.Errorf("%w: batch has no bottles", ErrInvalidState)
	}
	now := s.now()
	for _, btl := range bottles {
		if btl.Status != domain.BottlePending {
			return fmt.Errorf("%w: bottle %s is %s", ErrInvalidState, btl.Code, btl.Status)
		}
		btl.Status = domain.BottleSampled
		btl.SampledAt = &now
		if err := s.st.SaveBottle(btl); err != nil {
			return err
		}
	}
	b.Status = domain.BatchSampled
	b.UpdatedAt = now
	return s.st.SaveBatch(b)
}

// HandoverBottle adds a handover record and updates bottle status if needed.
// When all bottles in the batch are handed over, batch status becomes handed_over.
func (s *WaterService) HandoverBottle(bottleID, from, to, note string) (*domain.HandoverRecord, error) {
	if from == "" || to == "" {
		return nil, fmt.Errorf("%w: from and to required", ErrInvalidInput)
	}
	btl, err := s.st.GetBottle(bottleID)
	if err != nil {
		return nil, err
	}
	if btl.Status != domain.BottleSampled && btl.Status != domain.BottleHandedOver {
		return nil, fmt.Errorf("%w: bottle status is %s", ErrInvalidState, btl.Status)
	}
	id, err := s.genID("hdo")
	if err != nil {
		return nil, err
	}
	now := s.now()
	rec := &domain.HandoverRecord{
		ID:           id,
		BottleID:     bottleID,
		From:         from,
		To:           to,
		HandoverTime: now,
		Note:         note,
	}
	if err := s.st.AddHandover(rec); err != nil {
		return nil, err
	}
	if btl.Status == domain.BottleSampled {
		btl.Status = domain.BottleHandedOver
		btl.HandoverAt = &now
		if err := s.st.SaveBottle(btl); err != nil {
			return nil, err
		}
		// update batch status if all bottles now handed over
		batch, err := s.st.GetBatch(btl.BatchID)
		if err != nil {
			return nil, err
		}
		if batch.Status == domain.BatchSampled {
			bottles, err := s.st.ListBottlesByBatch(batch.ID)
			if err != nil {
				return nil, err
			}
			allHanded := true
			for _, b := range bottles {
				if b.Status != domain.BottleHandedOver {
					allHanded = false
					break
				}
			}
			if allHanded {
				batch.Status = domain.BatchHandedOver
				batch.UpdatedAt = now
				if err := s.st.SaveBatch(batch); err != nil {
					return nil, err
				}
			}
		}
	}
	return rec, nil
}

// ConfirmConclusion confirms conclusion for batch. Requires all bottles handed over.
func (s *WaterService) ConfirmConclusion(batchID, conclusion string) (*domain.Conclusion, error) {
	if conclusion == "" {
		return nil, fmt.Errorf("%w: conclusion required", ErrInvalidInput)
	}
	b, err := s.st.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	if b.Status != domain.BatchHandedOver {
		return nil, fmt.Errorf("%w: batch status is %s, need handed_over", ErrInvalidState, b.Status)
	}
	bottles, err := s.st.ListBottlesByBatch(batchID)
	if err != nil {
		return nil, err
	}
	for _, btl := range bottles {
		if btl.Status != domain.BottleHandedOver {
			return nil, fmt.Errorf("%w: bottle %s is %s", ErrInvalidState, btl.Code, btl.Status)
		}
	}
	now := s.now()
	c := &domain.Conclusion{
		BatchID:     batchID,
		Conclusion:  conclusion,
		ConcludedAt: now,
	}
	if err := s.st.SaveConclusion(c); err != nil {
		return nil, err
	}
	b.Status = domain.BatchConcluded
	b.UpdatedAt = now
	if err := s.st.SaveBatch(b); err != nil {
		return nil, err
	}
	// update bottles to concluded (optional but good)
	for _, btl := range bottles {
		btl.Status = domain.BottleConcluded
		if err := s.st.SaveBottle(btl); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (s *WaterService) GetHandoverRecords(bottleID string) ([]*domain.HandoverRecord, error) {
	return s.st.ListHandoverByBottle(bottleID)
}

// SampleSummary returns pending samples.
type SampleSummary struct {
	PendingHandover   []*domain.Bottle `json:"pending_handover"`
	PendingConclusion []*domain.Bottle `json:"pending_conclusion"`
}

func (s *WaterService) SampleSummary() (*SampleSummary, error) {
	batches, err := s.st.ListBatches()
	if err != nil {
		return nil, err
	}
	summary := &SampleSummary{}
	for _, b := range batches {
		bottles, err := s.st.ListBottlesByBatch(b.ID)
		if err != nil {
			return nil, err
		}
		for _, btl := range bottles {
			switch btl.Status {
			case domain.BottleSampled:
				summary.PendingHandover = append(summary.PendingHandover, btl)
			case domain.BottleHandedOver:
				summary.PendingConclusion = append(summary.PendingConclusion, btl)
			}
		}
	}
	return summary, nil
}

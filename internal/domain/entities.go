package domain

import "time"

type Point struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type BatchStatus string

const (
	BatchCreated   BatchStatus = "created"
	BatchSampled   BatchStatus = "sampled"
	BatchHandedOver BatchStatus = "handed_over"
	BatchConcluded BatchStatus = "concluded"
)

type Batch struct {
	ID        string      `json:"id"`
	PointID   string      `json:"point_id"`
	BatchNo   string      `json:"batch_no"`
	Status    BatchStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type BottleStatus string

const (
	BottlePending   BottleStatus = "pending"
	BottleSampled   BottleStatus = "sampled"
	BottleHandedOver BottleStatus = "handed_over"
	BottleConcluded BottleStatus = "concluded"
)

type Bottle struct {
	ID        string       `json:"id"`
	BatchID   string       `json:"batch_id"`
	Code      string       `json:"code"`
	Status    BottleStatus `json:"status"`
	SampledAt *time.Time   `json:"sampled_at,omitempty"`
	HandoverAt *time.Time  `json:"handover_at,omitempty"`
}

type HandoverRecord struct {
	ID           string    `json:"id"`
	BottleID     string    `json:"bottle_id"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	HandoverTime time.Time `json:"handover_time"`
	Note         string    `json:"note"`
}

type Conclusion struct {
	BatchID    string    `json:"batch_id"`
	Conclusion string    `json:"conclusion"`
	ConcludedAt time.Time `json:"concluded_at"`
}

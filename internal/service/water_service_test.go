package service

import (
	"testing"
	"time"

	"water-quality/internal/store"
)

func newTestService(t *testing.T) (*WaterService, *store.JSONStore) {
	t.Helper()
	st, err := store.NewJSONStore(t.TempDir() + "/test.json")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewWaterService(st)
	fixedNow := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.SetNowFunc(func() time.Time { return fixedNow })
	return svc, st
}

func TestFullBusinessChain(t *testing.T) {
	svc, _ := newTestService(t)

	// create point
	point, err := svc.CreatePoint("采样点A", "位置A")
	if err != nil {
		t.Fatal(err)
	}

	// create batch
	batch, err := svc.CreateBatch(point.ID, "B001")
	if err != nil {
		t.Fatal(err)
	}

	// add bottles
	b1, err := svc.AddBottle(batch.ID, "S001")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := svc.AddBottle(batch.ID, "S002")
	if err != nil {
		t.Fatal(err)
	}

	// complete sampling
	if err := svc.CompleteSampling(batch.ID); err != nil {
		t.Fatal(err)
	}

	// handover b1
	if _, err := svc.HandoverBottle(b1.ID, "采集员", "运输员", "冷链"); err != nil {
		t.Fatal(err)
	}

	// batch status should still be sampled
	b, _ := svc.GetBatch(batch.ID)
	if b.Status != "sampled" {
		t.Fatalf("expected sampled, got %s", b.Status)
	}

	// handover b2
	if _, err := svc.HandoverBottle(b2.ID, "采集员", "运输员", "冷链"); err != nil {
		t.Fatal(err)
	}

	// batch should now be handed_over
	b, _ = svc.GetBatch(batch.ID)
	if b.Status != "handed_over" {
		t.Fatalf("expected handed_over, got %s", b.Status)
	}

	// confirm conclusion
	conclusion, err := svc.ConfirmConclusion(batch.ID, "合格")
	if err != nil {
		t.Fatal(err)
	}
	if conclusion.Conclusion != "合格" {
		t.Fatalf("unexpected conclusion: %s", conclusion.Conclusion)
	}
}

func TestConfirmBeforeHandoverFailsAndStateUnchanged(t *testing.T) {
	svc, _ := newTestService(t)
	point, _ := svc.CreatePoint("P", "L")
	batch, _ := svc.CreateBatch(point.ID, "B002")
	bottle, _ := svc.AddBottle(batch.ID, "S003")
	if err := svc.CompleteSampling(batch.ID); err != nil {
		t.Fatal(err)
	}

	// Failed confirmations
	_, err := svc.ConfirmConclusion(batch.ID, "任意")
	if err == nil {
		t.Fatal("expected error")
	}

	// state unchanged
	b, _ := svc.GetBatch(batch.ID)
	if b.Status != "sampled" {
		t.Fatalf("batch status changed: %s", b.Status)
	}
	bt, _ := svc.GetBottle(bottle.ID)
	if bt.Status != "sampled" {
		t.Fatalf("bottle status changed: %s", bt.Status)
	}
}

func TestAddBottleToNonCreatedBatchFails(t *testing.T) {
	svc, _ := newTestService(t)
	point, _ := svc.CreatePoint("P", "L")
	batch, _ := svc.CreateBatch(point.ID, "B003")
	if _, err := svc.AddBottle(batch.ID, "S004"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteSampling(batch.ID); err != nil {
		t.Fatal(err)
	}

	// adding bottle after sampling should fail
	if _, err := svc.AddBottle(batch.ID, "S004"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSampleSummary(t *testing.T) {
	svc, _ := newTestService(t)
	point, _ := svc.CreatePoint("P", "L")
	batch1, _ := svc.CreateBatch(point.ID, "B004")
	btl1, _ := svc.AddBottle(batch1.ID, "S005")
	btl2, _ := svc.AddBottle(batch1.ID, "S006")
	if err := svc.CompleteSampling(batch1.ID); err != nil {
		t.Fatal(err)
	}
	// handover only one bottle
	if _, err := svc.HandoverBottle(btl1.ID, "A", "B", ""); err != nil {
		t.Fatal(err)
	}

	summary, err := svc.SampleSummary()
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.PendingHandover) != 1 {
		t.Fatalf("expected 1 pending handover, got %d", len(summary.PendingHandover))
	}
	if len(summary.PendingConclusion) != 1 {
		t.Fatalf("expected 1 pending conclusion, got %d", len(summary.PendingConclusion))
	}
	// the handed over one should be pending conclusion
	if summary.PendingConclusion[0].ID != btl1.ID {
		t.Fatalf("unexpected pending conclusion bottle")
	}
	// the sampled one should be pending handover
	if summary.PendingHandover[0].ID != btl2.ID {
		t.Fatalf("unexpected pending handover bottle")
	}
}

func TestAllHandedBottlesPermitConclusion(t *testing.T) {
	svc, _ := newTestService(t)
	point, err := svc.CreatePoint("下游河段", "北岸")
	if err != nil {
		t.Fatal(err)
	}
	batch, err := svc.CreateBatch(point.ID, "B005")
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.AddBottle(batch.ID, "S007")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.AddBottle(batch.ID, "S008")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteSampling(batch.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.HandoverBottle(first.ID, "采集组", "运输组", "冷链交接"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.HandoverBottle(second.ID, "采集组", "运输组", "冷链交接"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConfirmConclusion(batch.ID, "合格"); err != nil {
		t.Fatalf("all handed-over bottles should allow a conclusion: %v", err)
	}
}

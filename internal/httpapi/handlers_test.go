package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"water-quality/internal/service"
	"water-quality/internal/store"
)

func TestHandoverRejectsBottleFromAnotherBatch(t *testing.T) {
	st, err := store.NewJSONStore(t.TempDir() + "/water.json")
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewWaterService(st)
	point, err := svc.CreatePoint("上游点", "东岸")
	if err != nil {
		t.Fatal(err)
	}
	firstBatch, err := svc.CreateBatch(point.ID, "A-001")
	if err != nil {
		t.Fatal(err)
	}
	secondBatch, err := svc.CreateBatch(point.ID, "B-001")
	if err != nil {
		t.Fatal(err)
	}
	bottle, err := svc.AddBottle(firstBatch.ID, "S-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteSampling(firstBatch.ID); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/batches/"+secondBatch.ID+"/handover",
		bytes.NewBufferString(`{"bottle_id":"`+bottle.ID+`","from":"采集组","to":"运输组","note":"错误批次"}`),
	)
	NewHandler(svc).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("foreign bottle handover status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

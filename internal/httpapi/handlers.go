package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"water-quality/internal/service"
)

type Handler struct {
	svc *service.WaterService
}

func NewHandler(svc *service.WaterService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/":
		h.serveWeb(w, r)
	case path == "/api/points" && r.Method == http.MethodPost:
		h.createPoint(w, r)
	case path == "/api/points" && r.Method == http.MethodGet:
		h.listPoints(w, r)
	case strings.HasPrefix(path, "/api/points/") && r.Method == http.MethodGet:
		h.getPoint(w, r)
	case path == "/api/batches" && r.Method == http.MethodPost:
		h.createBatch(w, r)
	case path == "/api/batches" && r.Method == http.MethodGet:
		h.listBatches(w, r)
	case strings.HasPrefix(path, "/api/batches/") && strings.HasSuffix(path, "/samples") && r.Method == http.MethodGet:
		h.getBatchSamples(w, r)
	case strings.HasPrefix(path, "/api/batches/") && strings.HasSuffix(path, "/bottles") && r.Method == http.MethodPost:
		h.addBottle(w, r)
	case strings.HasPrefix(path, "/api/batches/") && strings.HasSuffix(path, "/sampling") && r.Method == http.MethodPost:
		h.completeSampling(w, r)
	case strings.HasPrefix(path, "/api/batches/") && strings.HasSuffix(path, "/handover") && r.Method == http.MethodPost:
		h.handoverBottle(w, r)
	case strings.HasPrefix(path, "/api/batches/") && strings.HasSuffix(path, "/conclusion") && r.Method == http.MethodPost:
		h.confirmConclusion(w, r)
	case strings.HasPrefix(path, "/api/batches/") && r.Method == http.MethodGet:
		h.getBatch(w, r)
	case path == "/api/samples/summary" && r.Method == http.MethodGet:
		h.sampleSummary(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	h.writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (h *Handler) createPoint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	p, err := h.svc.CreatePoint(req.Name, req.Location)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) listPoints(w http.ResponseWriter, r *http.Request) {
	points, err := h.svc.ListPoints()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, points)
}

func (h *Handler) getPoint(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/points/")
	p, err := h.svc.GetPoint(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err)
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) createBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PointID string `json:"point_id"`
		BatchNo string `json:"batch_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	b, err := h.svc.CreateBatch(req.PointID, req.BatchNo)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) listBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := h.svc.ListBatches()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, batches)
}

func (h *Handler) getBatch(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/batches/", "")
	b, err := h.svc.GetBatch(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err)
		return
	}
	h.writeJSON(w, http.StatusOK, b)
}

func (h *Handler) addBottle(w http.ResponseWriter, r *http.Request) {
	batchID := extractID(r.URL.Path, "/api/batches/", "/bottles")
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	b, err := h.svc.AddBottle(batchID, req.Code)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) getBatchSamples(w http.ResponseWriter, r *http.Request) {
	batchID := extractID(r.URL.Path, "/api/batches/", "/samples")
	bottles, err := h.svc.ListBottlesByBatch(batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err)
		return
	}
	h.writeJSON(w, http.StatusOK, bottles)
}

func (h *Handler) completeSampling(w http.ResponseWriter, r *http.Request) {
	batchID := extractID(r.URL.Path, "/api/batches/", "/sampling")
	if err := h.svc.CompleteSampling(batchID); err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handoverBottle(w http.ResponseWriter, r *http.Request) {
	batchID := extractID(r.URL.Path, "/api/batches/", "/handover")
	var req struct {
		BottleID string `json:"bottle_id"`
		From     string `json:"from"`
		To       string `json:"to"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	// verify bottle belongs to batch
	btl, err := h.svc.GetBottle(req.BottleID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err)
		return
	}
	if btl.BatchID == batchID {
		h.writeError(w, http.StatusBadRequest, errMismatch)
		return
	}
	rec, err := h.svc.HandoverBottle(req.BottleID, req.From, req.To, req.Note)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, rec)
}

func (h *Handler) confirmConclusion(w http.ResponseWriter, r *http.Request) {
	batchID := extractID(r.URL.Path, "/api/batches/", "/conclusion")
	var req struct {
		Conclusion string `json:"conclusion"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}
	c, err := h.svc.ConfirmConclusion(batchID, req.Conclusion)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) sampleSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := h.svc.SampleSummary()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, sum)
}

var errMismatch = &httpError{msg: "bottle does not belong to batch"}

type httpError struct{ msg string }

func (e *httpError) Error() string { return e.msg }

func extractID(path, prefix, suffix string) string {
	// path like /api/batches/{id}/samples
	id := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		id = strings.TrimSuffix(id, suffix)
	}
	// cut leading / and trailing / just in case
	id = strings.Trim(id, "/")
	return id
}

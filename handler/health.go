package handler

import (
	"database/sql"
	"net/http"
)

// HealthHandler provides liveness and readiness probe endpoints for
// Kubernetes health checks. No authentication is required.
type HealthHandler struct {
	db *sql.DB // Database connection used by the readiness check.
}

// NewHealthHandler constructs a HealthHandler with the given database connection.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Healthz handles GET /healthz. It always returns 200 OK — it confirms the
// process is alive and can serve HTTP traffic.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz handles GET /readyz. It pings the database and returns 200 if
// reachable, or 503 Service Unavailable if the database is down.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database not ready", "NOT_READY")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

package handler_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-service-starter/handler"
)

// fakeDriver / fakeConn implement the minimum database/sql/driver interfaces
// needed to create a *sql.DB that can succeed or fail Ping.
type fakeDriver struct{ pingErr error }
type fakeConn struct{ pingErr error }

func (d *fakeDriver) Open(_ string) (driver.Conn, error) { return &fakeConn{pingErr: d.pingErr}, nil }
func (c *fakeConn) Prepare(_ string) (driver.Stmt, error) { return nil, nil }
func (c *fakeConn) Close() error                          { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)             { return nil, nil }
func (c *fakeConn) Ping(_ context.Context) error          { return c.pingErr }

var _ driver.Pinger = (*fakeConn)(nil)

// fakeDriverSeq ensures each test gets a uniquely named driver registration,
// since sql.Register panics on duplicate names.
var fakeDriverSeq int

// newFakeDB returns a *sql.DB backed by fakeDriver. When pingErr is non-nil,
// calls to db.PingContext will return that error.
func newFakeDB(pingErr error) *sql.DB {
	fakeDriverSeq++
	name := fmt.Sprintf("fakedriver_%d", fakeDriverSeq)
	sql.Register(name, &fakeDriver{pingErr: pingErr})
	db, _ := sql.Open(name, "")
	return db
}

func TestHealthHandler_Healthz(t *testing.T) {
	h := handler.NewHealthHandler(newFakeDB(nil))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHealthHandler_Readyz(t *testing.T) {
	t.Run("returns 200 when database is reachable", func(t *testing.T) {
		h := handler.NewHealthHandler(newFakeDB(nil))
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		h.Readyz(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("returns 503 when database is unreachable", func(t *testing.T) {
		h := handler.NewHealthHandler(newFakeDB(fmt.Errorf("connection refused")))
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		h.Readyz(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

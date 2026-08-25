package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lacsar712/vialfill/internal/app"
	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/config"
	"github.com/lacsar712/vialfill/internal/model"
	"github.com/lacsar712/vialfill/internal/web/api"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("DRUM-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	drum := a.Snapshot().Vialtray
	drum.LevelPercent = model.MinVialtrayLevelPercent - 1
	if err := a.Store().UpdateVialtray(cfg.UnitID, drum); err != nil {
		t.Fatal(err)
	}
	s := api.NewServer(a)
	req := httptest.NewRequest(http.MethodGet, "/vialtray/level", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "vialtray_level_low" {
		t.Fatalf("expected vialtray_level_low code, got %q", body.Error)
	}
}

package dashboard_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/dashboard"
)

type fakeDashboardService struct {
	resp *dashboard.DashboardResponse
	err  error
}

func (f *fakeDashboardService) GetDashboard(_ context.Context, _ dashboard.GetDashboardRequest) (*dashboard.DashboardResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.resp != nil {
		return f.resp, nil
	}
	return &dashboard.DashboardResponse{TopVehicleTypes: []dashboard.VehicleTypeMetrics{}}, nil
}

func newTestEngine(svc dashboard.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := dashboard.NewHandler(svc)
	dashboard.RegisterRoutes(engine.Group("/api/v1"), h)
	return engine
}

func TestGetDashboardSuccess(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"success":true`) {
		t.Fatalf("expected success response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"sales_summary"`) {
		t.Fatalf("expected sales_summary in response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"inventory_summary"`) {
		t.Fatalf("expected inventory_summary in response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"expense_summary"`) {
		t.Fatalf("expected expense_summary in response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"top_vehicle_types"`) {
		t.Fatalf("expected top_vehicle_types in response, got %s", resp.Body.String())
	}
}

func TestGetDashboardDefaultDurationLifetime(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestGetDashboardInvalidDurationReturns400(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{err: dashboard.ErrInvalidDuration})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?duration=invalid", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":false`) {
		t.Fatalf("expected error response, got %s", resp.Body.String())
	}
}

func TestGetDashboardInvalidShowroomIDReturns400(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?showroom_id=abc", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

func TestGetDashboardValidShowroomIDSucceeds(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?showroom_id=1", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetDashboardServiceErrorPropagated(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{err: errors.New("unexpected db error")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}

func TestGetDashboardResponseMessage(t *testing.T) {
	engine := newTestEngine(&fakeDashboardService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if !strings.Contains(resp.Body.String(), "dashboard data fetched") {
		t.Fatalf("expected 'dashboard data fetched' message, got %s", resp.Body.String())
	}
}

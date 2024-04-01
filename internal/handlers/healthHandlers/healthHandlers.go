package healthHandlers

import (
	"github.com/17HIERARCH70/SocialManager/internal/services"
	"golang.org/x/exp/slog"
	"net/http"
)

// HealthCheckHandler handles HTTP requests for health checks.
type HealthCheckHandler struct {
	HealthService services.HealthServiceMethods
	Log           *slog.Logger
}

func (h *HealthCheckHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.HealthService.Check(); err != nil {
		h.Log.Error("Health check failed", "error", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Package health implements Kubernetes-compatible health check endpoints
package health

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Checker implements liveness and readiness probes
type Checker struct {
	mu    sync.RWMutex
	ready bool
}

// HealthResponse is the JSON response for health endpoints
type HealthResponse struct {
	Status  string `json:"status"`
	Ready   bool   `json:"ready,omitempty"`
	Version string `json:"version,omitempty"`
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		ready: false,
	}
}

// SetReady sets the readiness state
func (c *Checker) SetReady(ready bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ready = ready
}

// IsReady returns the current readiness state
func (c *Checker) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

// LivenessHandler handles /health/live requests
// Returns 200 if the process is alive
func (c *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "alive",
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessHandler handles /health/ready requests
// Returns 200 if the server is ready to accept traffic, 503 otherwise
func (c *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ready := c.IsReady()

	response := HealthResponse{
		Ready:   ready,
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")

	if ready {
		response.Status = "ready"
		w.WriteHeader(http.StatusOK)
	} else {
		response.Status = "not_ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// PingResponse represents the response
// structure for the ping endpoint
type PingResponse struct {
	Message  string `json:"message"`
	Hostname string `json:"hostname"`
}

// pingHandler handles the /api/ping endpoint
func PingHandler(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("failed to get hostname", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	response := PingResponse{
		Message:  "pong",
		Hostname: hostname,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("Ping request handled", "hostname", hostname, "remote_addr", r.RemoteAddr)
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	id := flag.String("id", envOr("BACKEND_ID", "backend"), "backend identifier")
	port := flag.Int("port", envInt("PORT", 8081), "listen port")
	latency := flag.Duration("latency", envDuration("LATENCY", 25*time.Millisecond), "base response latency")
	errorRate := flag.Float64("error-rate", envFloat("ERROR_RATE", 0), "failure probability between 0 and 1")
	flag.Parse()
	if *errorRate < 0 || *errorRate > 1 {
		slog.Error("error-rate must be between 0 and 1")
		os.Exit(2)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "backend": *id})
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(*latency)
		if rand.Float64() < *errorRate {
			writeJSON(writer, http.StatusInternalServerError, map[string]any{"code": "SIMULATED_FAILURE", "message": "The mock backend generated a configured failure.", "statusCode": 500})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"backend": *id, "method": request.Method, "path": request.URL.Path, "latency": latency.String()})
	})
	server := &http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	slog.Info("mock backend started", "id", *id, "address", server.Addr, "latency", *latency, "error_rate", *errorRate)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("mock backend stopped", "error", err)
		os.Exit(1)
	}
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err == nil {
		return value
	}
	return fallback
}
func envFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err == nil {
		return value
	}
	return fallback
}
func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err == nil {
		return value
	}
	return fallback
}
